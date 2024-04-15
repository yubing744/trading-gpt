package pkg

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/chat"
	"github.com/yubing744/trading-gpt/pkg/chat/feishu"
	"github.com/yubing744/trading-gpt/pkg/llms"
	"github.com/yubing744/trading-gpt/pkg/prompt"

	"github.com/yubing744/trading-gpt/pkg/agents"
	"github.com/yubing744/trading-gpt/pkg/agents/keeper"
	"github.com/yubing744/trading-gpt/pkg/agents/trading"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/env"
	"github.com/yubing744/trading-gpt/pkg/env/exchange"
	"github.com/yubing744/trading-gpt/pkg/env/fng"
	"github.com/yubing744/trading-gpt/pkg/utils"

	nfeishu "github.com/yubing744/trading-gpt/pkg/notify/feishu"
	feishu_hook "github.com/yubing744/trading-gpt/pkg/notify/feishu-hook"
	ttypes "github.com/yubing744/trading-gpt/pkg/types"
)

const MaxRetryTime = 1

// ID is the unique strategy ID, it needs to be in all lower case
// For example, grid strategy uses "grid"
const ID = "jarvis"

// log is a logrus.Entry that will be reused.
// This line attaches the strategy field to the logger with our ID, so that the logs from this strategy will be tagged with our ID
var log = logrus.WithField("jarvis", ID)

// init is a special function of golang, it will be called when the program is started
// importing this package will trigger the init function call.
func init() {
	// Register our struct type to BBGO
	// Note that you don't need to field the fields.
	// BBGO uses reflect to parse your type information.
	bbgo.RegisterStrategy(ID, &Strategy{})
}

// Strategy is a struct that contains the settings of your strategy.
// These settings will be loaded from the BBGO YAML config file "bbgo.yaml" automatically.
type Strategy struct {
	config.Config
	Environment *bbgo.Environment
	Market      types.Market

	// persistence fields
	Position *types.Position

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor

	// StrategyController
	bbgo.StrategyController

	// jarvis model
	llm          *llms.LLMManager
	world        *env.Environment
	agent        agents.IAgent
	chatSessions *chat.ChatSessions
}

// ID should return the identity of this strategy
func (s *Strategy) ID() string {
	return ID
}

// InstanceID returns the identity of the current instance of this strategy.
// You may have multiple instance of a strategy, with different symbols and settings.
// This value will be used for persistence layer to separate the storage.
//
// Run:
//
//	redis-cli KEYS "*"
//
// And you will see how this instance ID is used in redis.
func (s *Strategy) InstanceID() string {
	return ID + ":" + s.Symbol
}

// Subscribe method subscribes specific market data from the given session.
// Before BBGO is connected to the exchange, we need to collect what we want to subscribe.
// Here the strategy needs kline data, so it adds the kline subscription.
func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	// We want 1m kline data of the symbol
	// It will be BTCUSDT 1m if our s.Symbol is BTCUSDT
	log.Info("subscribe KLineChannel")
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: s.Interval})
}

// This strategy simply spent all available quote currency to buy the symbol whenever kline gets closed
func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	log.Info("Strategy_Run")

	s.session = session

	// calculate group id for orders
	instanceID := s.InstanceID()

	// If position is nil, we need to allocate a new position for calculation
	if s.Position == nil {
		s.Position = types.NewPositionFromMarket(s.Market)
	}
	// Always update the position fields
	s.Position.Strategy = ID
	s.Position.StrategyInstanceID = instanceID

	// Set fee rate
	if s.session.MakerFeeRate.Sign() > 0 || s.session.TakerFeeRate.Sign() > 0 {
		s.Position.SetExchangeFeeRate(s.session.ExchangeName, types.ExchangeFee{
			MakerFeeRate: s.session.MakerFeeRate,
			TakerFeeRate: s.session.TakerFeeRate,
		})
	}

	// Setup order executor
	s.orderExecutor = bbgo.NewGeneralOrderExecutor(session, s.Symbol, ID, instanceID, s.Position)
	s.orderExecutor.BindEnvironment(s.Environment)
	s.orderExecutor.Bind()

	// Sync position to redis on trade
	s.orderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		log.WithField("position", position).Info("Strategy_OnPositionUpdate")
		bbgo.Sync(ctx, s)
	})

	// Setup LLM
	err := s.setupLLM(ctx)
	if err != nil {
		return err
	}

	// Setup Environment
	err = s.setupWorld(ctx)
	if err != nil {
		return err
	}

	// Setup Agent
	err = s.setupAgent(ctx)
	if err != nil {
		return err
	}

	// Setup Notify
	err = s.setupNotify(ctx)
	if err != nil {
		return err
	}

	// Setup Chat
	err = s.setupChat(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Strategy) setupLLM(ctx context.Context) error {
	llm := llms.NewLLMManager(&s.LLM)
	err := llm.Init()
	if err != nil {
		return errors.Wrap(err, "Init LLM fail")
	}

	s.llm = llm
	return nil
}

func (s *Strategy) setupWorld(ctx context.Context) error {
	world := env.NewEnvironment(&s.Env)
	world.RegisterEntity(exchange.NewExchangeEntity(
		s.Symbol,
		s.Interval,
		s.Leverage,
		s.Env.ExchangeConfig,
		s.session,
		s.orderExecutor,
		s.Position,
	))
	world.RegisterEntity(fng.NewFearAndGreedEntity())

	err := world.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "Error in start env")
	}

	s.world = world
	return nil
}

func (s *Strategy) setupAgent(ctx context.Context) error {
	var tradingAgent *trading.TradingAgent
	tradingCfg := &s.Agent.Trading
	if tradingCfg != nil && tradingCfg.Enabled {
		tradingAgent := trading.NewTradingAgent(tradingCfg, s.llm)
		s.agent = tradingAgent
	}

	keeperCfg := &s.Agent.Keeper
	if keeperCfg != nil && keeperCfg.Enabled {
		agents := make(map[string]agents.IAgent, 0)

		if tradingCfg != nil && tradingCfg.Enabled {
			agents["trading"] = tradingAgent
		}

		agentKeeper := keeper.NewAgentKeeper(keeperCfg, agents)
		s.agent = agentKeeper
	}

	if s.agent == nil {
		return errors.New("No agent enabled")
	}

	err := s.agent.Start()
	if err != nil {
		return errors.Wrap(err, "Error in init agent")
	}

	return nil
}

func (s *Strategy) setupNotify(ctx context.Context) error {
	feishuNotifyCfg := s.Notify.Feishu
	if feishuNotifyCfg != nil && feishuNotifyCfg.Enabled {
		if os.Getenv("NOTIFY_FEISHU_APP_ID") != "" {
			feishuNotifyCfg.AppId = os.Getenv("NOTIFY_FEISHU_APP_ID")
			feishuNotifyCfg.AppSecret = os.Getenv("NOTIFY_FEISHU_APP_SECRET")
		}

		feishuNotifyChannel := nfeishu.NewFeishuNotifyChannel(feishuNotifyCfg)
		chatSession := chat.NewChatSession(feishuNotifyChannel)
		s.setupAdminSession(ctx, chatSession)
		s.agentAction(ctx, chatSession, []*ttypes.Message{{
			Text: "wait",
		}}, MaxRetryTime)

		log.Info("init feishu notify channel ok!")
	}

	hookNotifyCfg := s.Notify.FeishuHook
	if hookNotifyCfg != nil && hookNotifyCfg.Enabled {
		feishuHookNotifyChannel := feishu_hook.NewFeishuHookNotifyChannel(hookNotifyCfg)
		chatSession := chat.NewChatSession(feishuHookNotifyChannel)
		s.setupAdminSession(ctx, chatSession)
		s.agentAction(ctx, chatSession, []*ttypes.Message{{
			Text: "wait",
		}}, MaxRetryTime)

		log.Info("init feishu hook notify channel ok!")
	}

	return nil
}

func (s *Strategy) setupChat(ctx context.Context) error {
	feishuCfg := s.Chat.Feishu
	if feishuCfg != nil && feishuCfg.Enabled {
		if feishuCfg != nil && os.Getenv("CHAT_FEISHU_APP_ID") != "" {
			feishuCfg.AppId = os.Getenv("CHAT_FEISHU_APP_ID")
			feishuCfg.AppSecret = os.Getenv("CHAT_FEISHU_APP_SECRET")
			feishuCfg.EventEncryptKey = os.Getenv("CHAT_FEISHU_EVENT_ENCRYPT_KEY")
			feishuCfg.VerificationToken = os.Getenv("CHAT_FEISHU_VERIFICATION_TOKEN")
		}

		chatProvider := feishu.NewFeishuChatProvider(feishuCfg)
		sessions := chat.NewChatSessions()
		adminInit := &sync.Once{}

		go func() {
			err := chatProvider.Listen(func(ch ttypes.IChannel) {
				log.WithField("channel", ch).Info("new channel")

				chatSession := chat.NewChatSession(ch)
				sessions.AddChatSession(chatSession)

				ch.OnMessage(func(msg *ttypes.Message) {
					s.handleChatMessage(context.Background(), chatSession, msg)
				})

				adminInit.Do(func() {
					s.setupAdminSession(ctx, chatSession)
				})
			})
			if err != nil {
				log.WithError(err).Error("listen chat error")
			}
		}()
		s.chatSessions = sessions
	}

	return nil
}

func (s *Strategy) setupAdminSession(ctx context.Context, chatSession ttypes.ISession) {
	chatSession.SetRoles([]string{ttypes.RoleAdmin})

	s.world.OnEvent(func(evt *ttypes.Event) {
		s.handleEnvEvent(context.Background(), chatSession, evt)
	})
}

func (s *Strategy) replyMsg(ctx context.Context, chatSession ttypes.ISession, msg string) {
	err := chatSession.Reply(ctx, &ttypes.Message{
		ID:   uuid.NewString(),
		Text: msg,
	})
	if err != nil {
		log.WithError(err).Error("reply message error")
	}
}

func (s *Strategy) feedbackCmdExecuteResult(ctx context.Context, chatSession ttypes.ISession, msg string) {
	s.replyMsg(ctx, chatSession, msg)
}

func (s *Strategy) emergencyClosePosition(ctx context.Context, chatSession ttypes.ISession, reason string) {
	log.Warn("emergency close position")

	err := s.world.SendCommand(ctx, "exchange.close_position", map[string]string{})
	if err != nil {
		log.WithError(err).Error("env send cmd error")
		return
	}

	log.Warn("emergency close position ok")
	s.replyMsg(ctx, chatSession, fmt.Sprintf("emergency close position, for %s", reason))
}

func (s *Strategy) agentAction(ctx context.Context, chatSession ttypes.ISession, msgs []*ttypes.Message, retryTime int) {
	s.replyMsg(ctx, chatSession, fmt.Sprintf("The agent start action at %s, and the msgs:", time.Now().Format(time.RFC3339)))
	for _, msg := range msgs {
		s.replyMsg(ctx, chatSession, msg.Text)
	}

	resp, err := s.agent.GenActions(ctx, chatSession, msgs)
	if err != nil {
		log.WithError(err).Error("gen action error")
		s.replyMsg(ctx, chatSession, fmt.Sprintf("gen action error: %s", err.Error()))

		if chatSession.HasRole(ttypes.RoleAdmin) {
			s.emergencyClosePosition(ctx, chatSession, "agent error")
		}

		return
	}

	actions := make([]*ttypes.Action, 0)

	if len(resp.Texts) > 0 {
		resultText := strings.TrimSpace(strings.Join(resp.Texts, ""))

		if strings.HasPrefix(resultText, "{") || strings.Contains(resultText, "```json") {
			result, err := utils.ParseResult(resultText)
			if err != nil {
				log.WithError(err).WithField("resultText", resultText).Error("parse resp error")
				s.replyMsg(ctx, chatSession, fmt.Sprintf("parse resp error: %s, resultText: %s", err.Error(), resultText))
				return
			}

			if result.Thoughts != nil {
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Text: %s", result.Thoughts.Text))
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Analyze: %s", result.Thoughts.Analyze))
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Criticism: %s", result.Thoughts.Criticism))
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Speak: %s", result.Thoughts.Speak))
			}

			if result.Action != nil {
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Action: %s", result.Action.JSON()))

				if result.Action.Name != "" {
					actions = append(actions, result.Action)
				}
			}
		} else {
			s.replyMsg(ctx, chatSession, resultText)
		}
	}

	if len(actions) > 0 {
		if chatSession.HasRole(ttypes.RoleAdmin) {
			if len(actions) > 1 {
				log.Info("skip handle actions for too many actions")
				return
			}

			for _, action := range actions {
				actionName := action.Name
				if !strings.Contains(action.Name, ".") {
					actionName = "exchange." + actionName
				}

				err := s.world.SendCommand(ctx, actionName, action.Args)

				if err != nil {
					log.WithError(err).Error("env send cmd error")
					errMsg := fmt.Sprintf("Command: %s failed to execute by entity, reason: %s", action.JSON(), err.Error())
					s.feedbackCmdExecuteResult(ctx, chatSession, errMsg)

					if retryTime > 0 {
						time.Sleep(time.Second * 5)

						newMsgs := append(msgs, []*ttypes.Message{
							{
								Text: errMsg,
							},
							{
								Text: "Please try to fix the above error by responding with JSON again.",
							},
						}...)
						s.agentAction(ctx, chatSession, newMsgs, retryTime-1)
					}
				} else {
					s.feedbackCmdExecuteResult(ctx, chatSession, fmt.Sprintf("Command: %s executed successfully by entity.", action.JSON()))
				}
			}
		} else {
			log.Info("skip handle actions for not have RoleAdmin")
		}
	}
}

func (s *Strategy) handleChatMessage(ctx context.Context, chatSession *chat.ChatSession, msg *ttypes.Message) {
	log.WithField("msg", msg).Info("new message")
	s.agentAction(ctx, chatSession, []*ttypes.Message{msg}, MaxRetryTime)
}

func (s *Strategy) handleEnvEvent(ctx context.Context, session ttypes.ISession, evt *ttypes.Event) {
	log.WithField("event", evt).Info("handle env event")

	switch evt.Type {
	case "position_changed":
		position, ok := evt.Data.(*exchange.PositionX)
		if ok {
			s.handlePositionChanged(ctx, session, position)
		} else {
			log.WithField("eventType", evt.Type).Warn("event data Type not match")
		}
	case "kline_changed":
		klineWindow, ok := evt.Data.(*types.KLineWindow)
		if ok {
			s.handleKlineChanged(ctx, session, klineWindow)
		} else {
			log.WithField("eventType", evt.Type).Warn("event data Type not match")
		}
	case "indicator_changed":
		indicator, ok := evt.Data.(*exchange.ExchangeIndicator)
		if ok {
			s.handleExchangeIndicatorChanged(ctx, session, indicator)
		} else {
			log.WithField("eventType", evt.Type).Warn("event data Type not match")
		}
	case "fng_changed":
		fng, ok := evt.Data.(*string)
		if ok {
			s.handleFngChanged(ctx, session, fng)
		} else {
			log.WithField("eventType", evt.Type).Warn("event data Type not match")
		}
	case "update_finish":
		s.handleUpdateFinish(ctx, session)
	default:
		log.WithField("eventType", evt.Type).Warn("no match event type")
	}
}

func (s *Strategy) handleKlineChanged(ctx context.Context, session ttypes.ISession, klineWindow *types.KLineWindow) {
	log.WithField("kline", klineWindow).Info("handle klineWindow values changed")

	msg := fmt.Sprintf("KLine data changed: Open:[%s], Close:[%s], High:[%s], Low:[%s], Volume:[%s], and the current close price is: %.3f",
		utils.JoinFloatSeries(klineWindow.Open(), " "),
		utils.JoinFloatSeries(klineWindow.Close(), " "),
		utils.JoinFloatSeries(klineWindow.High(), " "),
		utils.JoinFloatSeries(klineWindow.Low(), " "),
		utils.JoinFloatSeries(klineWindow.Volume(), " "),
		klineWindow.GetClose().Float64(),
	)

	session.SetAttribute("kline", klineWindow)
	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleExchangeIndicatorChanged(ctx context.Context, session ttypes.ISession, indicator *exchange.ExchangeIndicator) {
	log.WithField("indicator", indicator).Info("handle indicator changed")

	messages := indicator.ToPrompts(s.MaxWindowSize)

	for _, msg := range messages {
		s.stashMsg(ctx, session, msg)
	}
}

func (s *Strategy) handleFngChanged(ctx context.Context, session ttypes.ISession, fng *string) {
	log.WithField("fng", fng).Info("handle FNG values changed")

	msg := fmt.Sprintf("The current Fear and Greed Index value is: %s", *fng)
	session.SetAttribute("fng_msg", &ttypes.Message{
		Text: msg,
	})
}

func (s *Strategy) handlePositionChanged(_ctx context.Context, session ttypes.ISession, position *exchange.PositionX) {
	log.WithField("position", position).Info("handle position changed")

	msg := "There are currently no open positions"

	kline, ok := s.getKline(session)
	if ok {
		if position.IsOpened(kline.GetClose()) {
			side := "short"
			if position.IsLong() {
				side = "long"
			}

			msg = fmt.Sprintf("The current position is %s with %dx leverage, average cost: %.3f, and accumulated profit: %.3f%s", side, s.Leverage.Int(), position.AverageCost.Float64(), position.AccumulatedProfit.Float64(), "%")

			profits := position.GetProfitValues()
			if len(profits) > s.MaxWindowSize {
				profits = profits[len(profits)-s.MaxWindowSize:]
			}

			msg = msg + fmt.Sprintf("\nThe profits of the recent %d periods: [%s], and the holding period: %d",
				s.MaxWindowSize,
				utils.JoinFloatSlicePercentage([]float64(profits), " "),
				position.GetHoldingPeriod())
		}

		session.SetAttribute("position_msg", &ttypes.Message{
			Text: msg,
		})
	}
}

func (s *Strategy) handleUpdateFinish(ctx context.Context, session ttypes.ISession) {
	tempMsgs, ok := s.popMsgs(ctx, session)
	log.WithField("tempMsgs", tempMsgs).Info("session tmp msgs")

	if ok {
		// fng
		fngMsg, ok := session.GetAttribute("fng_msg")
		if ok {
			tempMsgs = append(tempMsgs, fngMsg.(*ttypes.Message))
		}

		// position
		posMsg, ok := s.getPositionMsg(session)
		if ok {
			tempMsgs = append(tempMsgs, posMsg)
		}

		actionTips := make([]string, 0)
		for i, ac := range s.world.Actions() {
			actionTips = append(actionTips, fmt.Sprintf(`%d. %s`, i+1, ac.String()))
		}

		tempMsgs = append(tempMsgs, &ttypes.Message{
			Text: fmt.Sprintf(prompt.Thought, strings.Join(actionTips, "\n"), s.Strategy),
		})

		s.agentAction(ctx, session, tempMsgs, MaxRetryTime)
	}

	session.RemoveAttribute("tempMsgs")
}

func (s *Strategy) stashMsg(ctx context.Context, session ttypes.ISession, msg string) {
	tempMsgsRef, _ := session.GetAttribute("tempMsgs")
	tempMsgs, _ := tempMsgsRef.([]*ttypes.Message)
	tempMsgs = append(tempMsgs, &ttypes.Message{
		Text: msg,
	})

	log.WithField("tempMsgs", tempMsgs).Info("session tmp msgs")
	session.SetAttribute("tempMsgs", tempMsgs)
}

func (s *Strategy) popMsgs(ctx context.Context, session ttypes.ISession) ([]*ttypes.Message, bool) {
	tempMsgsRef, ok := session.GetAttribute("tempMsgs")
	if ok {
		return tempMsgsRef.([]*ttypes.Message), ok
	}

	return []*ttypes.Message{}, false
}

func (s *Strategy) getKline(session ttypes.ISession) (*types.KLineWindow, bool) {
	kline, ok := session.GetAttribute("kline")
	if ok {
		return kline.(*types.KLineWindow), ok
	}

	return nil, false
}

func (s *Strategy) getPositionMsg(session ttypes.ISession) (*ttypes.Message, bool) {
	positionMsg, ok := session.GetAttribute("position_msg")
	if ok {
		return positionMsg.(*ttypes.Message), ok
	}

	return nil, false
}
