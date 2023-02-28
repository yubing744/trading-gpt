package pkg

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/agent/chatgpt"
	"github.com/yubing744/trading-bot/pkg/agent/keeper"
	"github.com/yubing744/trading-bot/pkg/agent/openai"
	"github.com/yubing744/trading-bot/pkg/chat"
	"github.com/yubing744/trading-bot/pkg/chat/feishu"

	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/env/exchange"
	"github.com/yubing744/trading-bot/pkg/env/fng"
	"github.com/yubing744/trading-bot/pkg/utils"

	nfeishu "github.com/yubing744/trading-bot/pkg/notify/feishu"
	feishu_hook "github.com/yubing744/trading-bot/pkg/notify/feishu-hook"
	ttypes "github.com/yubing744/trading-bot/pkg/types"
)

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
	Position *types.Position `persistence:"position"`

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor

	// StrategyController
	bbgo.StrategyController

	// jarvis model
	world        *env.Environment
	agent        agent.IAgent
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
		bbgo.Sync(ctx, s)
	})

	// Setup Environment
	err := s.setupWorld(ctx)
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
	var openaiAgent *openai.OpenAIAgent
	openaiCfg := &s.Agent.OpenAI
	if openaiCfg != nil && openaiCfg.Enabled {
		token := os.Getenv("AGENT_OPENAI_TOKEN")
		if token == "" {
			return errors.New("AGENT_OPENAI_TOKEN not set in .env.local")
		}

		openaiCfg.Token = token
		openaiAgent = openai.NewOpenAIAgent(openaiCfg)
		s.agent = openaiAgent
	}

	var chatgptAgent *chatgpt.ChatGPTAgent
	chatgptCfg := &s.Agent.ChatGPT
	if chatgptCfg != nil && chatgptCfg.Enabled {
		email := os.Getenv("AGENT_CHATGPT_EMAIL")
		password := os.Getenv("AGENT_CHATGPT_PASSWORD")
		if email == "" || password == "" {
			return errors.New("AGENT_CHATGPT_EMAIL or AGENT_CHATGPT_PASSWORD not set in .env.local")
		}

		chatgptCfg.Email = email
		chatgptCfg.Password = password

		chatgptAgent = chatgpt.NewChatGPTAgent(chatgptCfg)
		s.agent = chatgptAgent
	}

	keeperCfg := &s.Agent.Keeper
	if keeperCfg != nil && keeperCfg.Enabled {
		agents := make(map[string]agent.IAgent, 0)

		if openaiCfg != nil && openaiCfg.Enabled {
			agents["openai"] = openaiAgent
		}

		if chatgptCfg != nil && chatgptCfg.Enabled {
			agents["chatgpt"] = chatgptAgent
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
		}})

		log.Info("init feishu notify channel ok!")
	}

	hookNotifyCfg := s.Notify.FeishuHook
	if hookNotifyCfg != nil && hookNotifyCfg.Enabled {
		feishuHookNotifyChannel := feishu_hook.NewFeishuHookNotifyChannel(hookNotifyCfg)
		chatSession := chat.NewChatSession(feishuHookNotifyChannel)
		s.setupAdminSession(ctx, chatSession)
		s.agentAction(ctx, chatSession, []*ttypes.Message{{
			Text: "wait",
		}})

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

	result, err := s.agent.GenActions(ctx, chatSession, []*ttypes.Message{
		{
			Text: msg,
		},
	})
	if err != nil {
		log.WithError(err).Error("gen action error")
		s.replyMsg(ctx, chatSession, fmt.Sprintf("gen action error: %s", err.Error()))
		return
	}

	log.WithField("result", result).Info("feedback result")

	if len(result.Texts) > 0 {
		s.replyMsg(ctx, chatSession, strings.Join(result.Texts, ""))
	}
}

func (s *Strategy) emergencyClosePosition(ctx context.Context, chatSession ttypes.ISession, reason string) {
	log.Warn("emergency close position")

	err := s.world.SendCommand(ctx, "exchange", "close_position", []string{})
	if err != nil {
		log.WithError(err).Error("env send cmd error")
		return
	}

	log.Warn("emergency close position ok")
	s.replyMsg(ctx, chatSession, fmt.Sprintf("emergency close position, for %s", reason))
}

func (s *Strategy) agentAction(ctx context.Context, chatSession ttypes.ISession, msgs []*ttypes.Message) {
	s.replyMsg(ctx, chatSession, fmt.Sprintf("The agent start action at %s, and the msgs:", time.Now().Format(time.RFC3339)))
	for _, msg := range msgs {
		s.replyMsg(ctx, chatSession, msg.Text)
	}

	result, err := s.agent.GenActions(ctx, chatSession, msgs)
	if err != nil {
		log.WithError(err).Error("gen action error")
		s.replyMsg(ctx, chatSession, fmt.Sprintf("gen action error: %s", err.Error()))

		if chatSession.HasRole(ttypes.RoleAdmin) {
			s.emergencyClosePosition(ctx, chatSession, "agent error")
		}

		return
	}

	log.WithField("result", result).Info("gen actions result")

	actions := make([]*ttypes.Action, 0)

	if len(result.Texts) > 0 {
		text := strings.Join(result.Texts, "")
		s.replyMsg(ctx, chatSession, text)

		for _, actionDef := range s.world.Actions() {
			if strings.Contains(strings.ToLower(text), fmt.Sprintf("/%s", actionDef.Name)) {
				log.WithField("action", actionDef.Name).Info("match action")

				actions = append(actions, &ttypes.Action{
					Target: "exchange",
					Name:   actionDef.Name,
					Args:   []string{},
				})
			}
		}
	}

	if len(actions) > 0 {
		if chatSession.HasRole(ttypes.RoleAdmin) {
			if len(actions) > 1 {
				log.Info("skip handle actions for too many actions")
				return
			}

			for _, action := range actions {
				err := s.world.SendCommand(ctx, action.Target, action.Name, action.Args)
				if err != nil {
					log.WithError(err).Error("env send cmd error")
					s.feedbackCmdExecuteResult(ctx, chatSession, fmt.Sprintf("Command: /%s [%s] failed to execute by entity, reason: %s", action.Name, strings.Join(action.Args, ","), err.Error()))
				} else {
					s.feedbackCmdExecuteResult(ctx, chatSession, fmt.Sprintf("Command: /%s [%s] executed successfully by entity.", action.Name, strings.Join(action.Args, ",")))
				}
			}
		} else {
			log.Info("skip handle actions for not have RoleAdmin")
		}
	}
}

func (s *Strategy) handleChatMessage(ctx context.Context, chatSession *chat.ChatSession, msg *ttypes.Message) {
	log.WithField("msg", msg).Info("new message")
	s.agentAction(ctx, chatSession, []*ttypes.Message{msg})
}

func (s *Strategy) handleEnvEvent(ctx context.Context, session ttypes.ISession, evt *ttypes.Event) {
	log.WithField("event", evt).Info("handle env event")

	switch evt.Type {
	case "position_changed":
		position, ok := evt.Data.(*types.Position)
		if ok {
			s.handlePositionChanged(ctx, session, position)
		} else {
			log.Warn("event data Type not match")
		}
	case "kline_changed":
		klineWindow, ok := evt.Data.(*types.KLineWindow)
		if ok {
			s.handleKlineChanged(ctx, session, klineWindow)
		} else {
			log.Warn("event data Type not match")
		}
	case "boll_changed":
		boll, ok := evt.Data.(*indicator.BOLL)
		if ok {
			s.handleBOLLChanged(ctx, session, boll)
		} else {
			log.Warn("event data Type not match")
		}
	case "rsi_changed":
		rsi, ok := evt.Data.(*indicator.RSI)
		if ok {
			s.handleRSIChanged(ctx, session, rsi)
		} else {
			log.Warn("event data Type not match")
		}
	case "fng_changed":
		fng, ok := evt.Data.(*string)
		if ok {
			s.handleFngChanged(ctx, session, fng)
		} else {
			log.Warn("event data Type not match")
		}
	case "update_finish":
		s.handleUpdateFinish(ctx, session)
	default:
		log.WithField("eventType", evt.Type).Warn("no match event type")
	}
}

func (s *Strategy) handlePositionChanged(ctx context.Context, session ttypes.ISession, position *types.Position) {
	log.WithField("position", position).Info("handle boll values changed")

	msg := ""

	if position.IsLong() {
		msg = fmt.Sprintf("The current position is long, average cost: %.3f, and accumulated profit: %.3f%s", position.AverageCost.Float64(), position.AccumulatedProfit.Float64(), "%")
	} else if position.IsShort() {
		msg = fmt.Sprintf("The current position is short, average cost: %.3f, and accumulated profit: %.3f%s", position.AverageCost.Float64(), position.AccumulatedProfit.Mul(fixedpoint.NewFromInt(-1)).Float64(), "%")
	} else {
		msg = "There are currently no open positions"
	}

	s.stashMsg(ctx, session, msg)
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

	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleBOLLChanged(ctx context.Context, session ttypes.ISession, boll *indicator.BOLL) {
	log.WithField("boll", boll).Info("handle boll values changed")

	upVals := boll.UpBand
	if len(upVals) > s.MaxWindowSize {
		upVals = upVals[len(upVals)-s.MaxWindowSize:]
	}

	midVals := boll.SMA.Values
	if len(midVals) > s.MaxWindowSize {
		midVals = midVals[len(midVals)-s.MaxWindowSize:]
	}

	downVals := boll.DownBand
	if len(downVals) > s.MaxWindowSize {
		downVals = downVals[len(downVals)-s.MaxWindowSize:]
	}

	msg := fmt.Sprintf("BOLL data changed: UpBand:[%s], SMA:[%s], DownBand:[%s]",
		utils.JoinFloatSlice([]float64(upVals), " "),
		utils.JoinFloatSlice([]float64(midVals), " "),
		utils.JoinFloatSlice([]float64(downVals), " "),
	)

	s.stashMsg(ctx, session, msg)

	msg = fmt.Sprintf("The current UpBand is %.3f, and the current SMA is %.3f, and the current DownBand is %.3f",
		boll.UpBand.Last(),
		boll.SMA.Last(),
		boll.DownBand.Last(),
	)

	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleRSIChanged(ctx context.Context, session ttypes.ISession, rsi *indicator.RSI) {
	log.WithField("rsi", rsi).Info("handle RSI values changed")

	vals := rsi.Values
	if len(vals) > s.MaxWindowSize {
		vals = vals[len(vals)-s.MaxWindowSize:]
	}

	msg := fmt.Sprintf("RSI data changed: [%s], and the current RSI value is: %.3f",
		utils.JoinFloatSlice([]float64(vals), " "),
		rsi.Last(),
	)

	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleFngChanged(ctx context.Context, session ttypes.ISession, fng *string) {
	log.WithField("fng", fng).Info("handle FNG values changed")

	msg := fmt.Sprintf("The current Fear and Greed Index value is: %s", *fng)
	session.SetAttribute("fng_msg", &ttypes.Message{
		Text: msg,
	})
}

func (s *Strategy) handleUpdateFinish(ctx context.Context, session ttypes.ISession) {
	tempMsgs, ok := s.popMsgs(ctx, session)
	log.WithField("tempMsgs", tempMsgs).Info("session tmp msgs")

	if ok {
		fngMsg, ok := session.GetAttribute("fng_msg")
		if ok {
			tempMsgs = append(tempMsgs, fngMsg.(*ttypes.Message))
		}

		tempMsgs = append(tempMsgs, &ttypes.Message{
			Text: "Trading strategy: Trading on the right side, trailing stop loss 3%, trailing stop profit 10%.",
		})

		actionTips := make([]string, 0)
		for _, ac := range s.world.Actions() {
			actionTips = append(actionTips, fmt.Sprintf("/%s", ac.Name))
		}

		tempMsgs = append(tempMsgs, &ttypes.Message{
			Text: fmt.Sprintf("Analyze the data and generate only one trading command: %s, the entity will execute the command and give you feedback.", strings.Join(actionTips, ",")),
		})

		s.agentAction(ctx, session, tempMsgs)
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
