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
	"github.com/yubing744/trading-gpt/pkg/utils/xtemplate"

	"github.com/yubing744/trading-gpt/pkg/agents"
	"github.com/yubing744/trading-gpt/pkg/agents/keeper"
	"github.com/yubing744/trading-gpt/pkg/agents/trading"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/env"
	"github.com/yubing744/trading-gpt/pkg/env/coze"
	"github.com/yubing744/trading-gpt/pkg/env/exchange"
	"github.com/yubing744/trading-gpt/pkg/env/fng"
	"github.com/yubing744/trading-gpt/pkg/env/twitterapi"
	"github.com/yubing744/trading-gpt/pkg/memory"
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

	// memory system
	memoryManager *memory.MemoryManager
	memoryEnabled bool
	currentMemory string

	// command system
	commandMemory *memory.CommandMemory
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

	s.SubscribeIntervals = append(s.SubscribeIntervals, s.Interval)
	for _, interval := range s.SubscribeIntervals {
		session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{Interval: interval})
	}
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

	// Setup Memory
	err = s.setupMemory(ctx)
	if err != nil {
		return err
	}

	// Setup Commands
	err = s.setupCommands(ctx)
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

	if s.Env.FNG != nil && s.Env.FNG.Enabled {
		log.Info("fng_enabled")

		world.RegisterEntity(fng.NewFearAndGreedEntity())
	}

	if s.Env.Coze != nil && s.Env.Coze.Enabled {
		log.Info("coze_enabled")

		cozeAPIKey := os.Getenv("COZE_API_KEY")
		if cozeAPIKey == "" {
			return errors.New("COZE_API_KEY not set in .env.local")
		}
		s.Env.Coze.APIKey = cozeAPIKey

		world.RegisterEntity(coze.NewCozeEntity(s.Env.Coze))
	}

	if s.Env.TwitterAPI != nil && s.Env.TwitterAPI.Enabled {
		log.Info("twitterapi_enabled")

		twitterAPIKey := os.Getenv("TWITTER_API_KEY")
		if twitterAPIKey == "" {
			return errors.New("TWITTER_API_KEY not set in .env.local")
		}
		s.Env.TwitterAPI.APIKey = twitterAPIKey

		world.RegisterEntity(twitterapi.NewTwitterAPIEntity(s.Env.TwitterAPI))
	}

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

func (s *Strategy) setupMemory(ctx context.Context) error {
	// Initialize memory manager if memory is enabled
	if s.Memory.Enabled {
		// Set default values if not configured
		if s.Memory.MemoryPath == "" {
			s.Memory.MemoryPath = "memory-bank/trading-memory.md"
		}
		if s.Memory.MaxWords == 0 {
			s.Memory.MaxWords = 1000
		}

		s.memoryManager = memory.NewMemoryManager(s.Memory.MemoryPath, s.Memory.MaxWords)
		s.memoryEnabled = true

		// Load existing memory
		memoryContent, err := s.memoryManager.LoadMemory()
		if err != nil {
			log.WithError(err).Warn("Failed to load existing memory")
		} else {
			s.currentMemory = memoryContent
		}

		log.Info("Memory system enabled")
	} else {
		s.memoryEnabled = false
		log.Info("Memory system disabled")
	}

	return nil
}

func (s *Strategy) setupCommands(ctx context.Context) error {
	// Initialize command memory manager if commands are enabled
	if s.Commands.Enabled {
		// Set default values if not configured
		if s.Commands.CommandPath == "" {
			s.Commands.CommandPath = "memory-bank/commands.json"
		}

		s.commandMemory = memory.NewCommandMemory(s.Commands.CommandPath)
		log.Info("Command system enabled")
	} else {
		log.Info("Command system disabled")
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
			Text: "Please wait a moment while I prepare the market data. ",
		}}, MaxRetryTime)

		log.Info("init feishu notify channel ok!")
	}

	hookNotifyCfg := s.Notify.FeishuHook
	if hookNotifyCfg != nil && hookNotifyCfg.Enabled {
		feishuHookNotifyChannel := feishu_hook.NewFeishuHookNotifyChannel(hookNotifyCfg)
		chatSession := chat.NewChatSession(feishuHookNotifyChannel)
		s.setupAdminSession(ctx, chatSession)
		s.agentAction(ctx, chatSession, []*ttypes.Message{{
			Text: "Please wait a moment while I prepare the market data. ",
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

	s.world.OnEvent(func(evt ttypes.IEvent) {
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

		hasThinking, thinkingText, resultText := utils.ExtractThinkingFull(resultText)
		if hasThinking {
			s.replyMsg(ctx, chatSession, fmt.Sprintf("Thinking: %s", thinkingText))
		}

		if strings.HasPrefix(resultText, "{") || strings.Contains(resultText, "```json") {
			result, err := utils.ParseResult(resultText)
			if err != nil {
				log.WithError(err).WithField("resultText", resultText).Error("parse resp error")

				errMsg := fmt.Sprintf("parse resp error, resultText: %s", resultText)
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

				return
			}

			if result.Thoughts != nil {
				s.replyMsg(ctx, chatSession, result.Thoughts.ToHumanText())
			}

			if result.Action != nil {
				s.replyMsg(ctx, chatSession, fmt.Sprintf("Action: %s", result.Action.JSON()))

				if result.Action.Name != "" {
					actions = append(actions, result.Action)
				}
			}

			// Process memory output if memory is enabled
			if s.memoryEnabled && s.memoryManager != nil && result.Memory != nil {
				s.processMemoryOutput(ctx, chatSession, result.Memory)
			}

			// Process next_commands if command system is enabled
			if s.commandMemory != nil && result.NextCommands != nil && len(result.NextCommands) > 0 {
				s.processNextCommands(ctx, chatSession, result.NextCommands)
			}
		} else {
			s.replyMsg(ctx, chatSession, resultText)
		}
	}

	if resp.Model != "" {
		s.replyMsg(ctx, chatSession, fmt.Sprintf("Generated by LLM model: %s", resp.Model))
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

func (s *Strategy) handleEnvEvent(ctx context.Context, session ttypes.ISession, evt ttypes.IEvent) {
	log.WithField("event", evt).Info("handle env event")

	switch evt.GetType() {
	case "position_changed":
		position, ok := evt.GetData().(*exchange.PositionX)
		if ok {
			s.handlePositionChanged(ctx, session, position)
		} else {
			log.WithField("eventType", evt.GetType()).Warn("event data Type not match")
		}
	case "kline_changed":
		klineWindow, ok := evt.GetData().(*types.KLineWindow)
		if ok {
			s.handleKlineChanged(ctx, session, klineWindow)
		} else {
			log.WithField("eventType", evt.GetType()).Warn("event data Type not match")
		}
	case "indicator_changed":
		indicator, ok := evt.GetData().(*exchange.ExchangeIndicator)
		if ok {
			s.handleExchangeIndicatorChanged(ctx, session, indicator)
		} else {
			log.WithField("eventType", evt.GetType()).Warn("event data Type not match")
		}
	case "fng_changed":
		fng, ok := evt.GetData().(*string)
		if ok {
			s.handleFngChanged(ctx, session, fng)
		} else {
			log.WithField("eventType", evt.GetType()).Warn("event data Type not match")
		}
	case exchange.EventPositionClosed:
		s.replyMsg(ctx, session, "Position closed event received")

		positionData, ok := evt.GetData().(exchange.PositionClosedEventData)
		if ok {
			s.handlePositionClosed(ctx, session, positionData)
		} else {
			log.WithField("eventType", evt.GetType()).Warn("event data Type not match")
		}
	case "update_finish":
		s.handleUpdateFinish(ctx, session)
	default:
		s.handleDefaultEvent(ctx, session, evt)
	}
}

func (s *Strategy) handleKlineChanged(ctx context.Context, session ttypes.ISession, klineWindow *types.KLineWindow) {
	log.WithField("kline", klineWindow).Info("handle klineWindow values changed")

	msg := fmt.Sprintf("KLine data changed:\n%s", utils.FormatKLineWindow(*klineWindow, s.MaxNum))

	session.SetAttribute("kline", klineWindow)
	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleExchangeIndicatorChanged(ctx context.Context, session ttypes.ISession, indicator *exchange.ExchangeIndicator) {
	log.WithField("indicator", indicator).Info("handle indicator changed")

	messages := indicator.ToPrompts(s.MaxNum)

	for _, msg := range messages {
		s.stashMsg(ctx, session, msg)
	}
}

func (s *Strategy) handleDefaultEvent(ctx context.Context, session ttypes.ISession, evt ttypes.IEvent) {
	messages := evt.ToPrompts()
	log.WithField("event", evt.GetType()).WithField("messages", messages).Info("handle_default_event")

	for _, msg := range messages {
		s.stashMsg(ctx, session, msg)
	}
}

func (s *Strategy) handleFngChanged(ctx context.Context, session ttypes.ISession, fng *string) {
	log.WithField("fng", fng).Info("handle FNG values changed")

	msg := fmt.Sprintf(
		"The current Crypto Market Fear & Greed Index (global, not asset-specific) is: %s. "+
			"This reflects overall sentiment across the entire cryptocurrency market: "+
			"0–24=Extreme Fear, 25–49=Fear, 50=Neutral, 51–74=Greed, 75–100=Extreme Greed. "+
			"Use this as a macro risk filter, not as a signal for the specific asset.",
		*fng,
	)
	session.SetAttribute("fng_msg", &ttypes.Message{
		Text: msg,
	})
}

func (s *Strategy) handlePositionChanged(_ctx context.Context, session ttypes.ISession, position *exchange.PositionX) {
	log.WithField("position", position).Info("handle position changed")

	msg := "There are currently no open positions"

	remainingPercent := position.RemainingFundsRatio.Float64() * 100
	positionPercent := position.PositionFundsRatio.Float64() * 100

	kline, ok := s.getKline(session)
	if ok {
		if position.IsOpened(kline.GetClose()) {
			side := "short"
			if position.IsLong() {
				side = "long"
			}

			msg = fmt.Sprintf("The current position is %s with %dx leverage, average cost: %.3f, and accumulated profit: %.3f%% (%.3f %s).",
				side,
				s.Leverage.Int(),
				position.AverageCost.Float64(),
				position.AccumulatedProfit.Float64(),
				position.AccumulatedProfitValue.Float64(),
				position.Market.QuoteCurrency)

			if position.TpTriggerPx != nil {
				msg += fmt.Sprintf("\nThe current position's take-profit trigger price is %s.", position.Market.FormatPrice(*position.TpTriggerPx))
			}

			if position.SlTriggerPx != nil {
				msg += fmt.Sprintf("\nThe current position's stop-loss trigger price is %s.", position.Market.FormatPrice(*position.SlTriggerPx))
			}

			profits := position.GetProfitValues()
			if len(profits) > s.MaxNum {
				profits = profits[len(profits)-s.MaxNum:]
			}

			msg = msg + fmt.Sprintf("\nThe profits of the recent %d periods: [%s], and the holding period: %d.",
				s.MaxNum,
				utils.JoinFloatSlicePercentage([]float64(profits), " "),
				position.GetHoldingPeriod())
		}

		msg = msg + fmt.Sprintf("\nAvailable quote capital: %.2f%% of total equity; current position exposure: %.2f%%.",
			remainingPercent,
			positionPercent)

		session.SetAttribute("position_msg", &ttypes.Message{
			Text: msg,
		})
	}
}

func (s *Strategy) handleUpdateFinish(ctx context.Context, session ttypes.ISession) {
	// Execute pending commands from previous cycle before collecting new data
	if s.commandMemory != nil {
		s.executeNextCycleCommands(ctx, session)
	}

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
		for _, ac := range s.world.Actions() {
			actionTips = append(actionTips, ac.String())
		}

		// Build template data
		templateData := map[string]interface{}{
			"ActionTips":              actionTips,
			"Strategy":                s.Strategy,
			"StrategyAttentionPoints": s.StrategyAttentionPoints,
		}

		// Add memory data if memory is enabled
		if s.memoryEnabled && s.memoryManager != nil {
			templateData["MemoryEnabled"] = true
			maxWords := s.memoryManager.GetMaxWords()
			templateData["MaxWords"] = maxWords

			memory, err := s.memoryManager.LoadMemory()
			if err != nil {
				log.WithError(err).Warn("Failed to load memory")
				templateData["Memory"] = ""
				templateData["CurrentWords"] = 0
				templateData["MemoryUsagePercent"] = 0
			} else {
				templateData["Memory"] = memory
				// Calculate memory usage
				currentWords := len(strings.Fields(memory))
				usagePercent := 0
				if maxWords > 0 {
					usagePercent = (currentWords * 100) / maxWords
				}
				templateData["CurrentWords"] = currentWords
				templateData["MemoryUsagePercent"] = usagePercent
			}
		} else {
			templateData["MemoryEnabled"] = false
		}

		prompt, err := xtemplate.Render(prompt.ThoughtTpl, templateData)
		if err != nil {
			s.replyMsg(ctx, session, fmt.Sprintf("Render prompt error: %s", err.Error()))
			return
		}

		tempMsgs = append(tempMsgs, &ttypes.Message{
			Text: prompt,
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

// handlePositionClosed processes position closed events and generates trading reflections
func (s *Strategy) handlePositionClosed(ctx context.Context, session ttypes.ISession, posData exchange.PositionClosedEventData) {
	log.WithField("positionData", posData).Info("Handling position closed event")

	// Notify users about position closure
	pnlStr := "loss"
	if posData.ProfitAndLoss >= 0 {
		pnlStr = "profit"
	}

	// Format notification message
	message := fmt.Sprintf("Position closed for %s:\n"+
		"Strategy: %s\n"+
		"Symbol: %s\n"+
		"Entry Price: %.2f\n"+
		"Exit Price: %.2f\n"+
		"Quantity: %.6f\n"+
		"%s: %.2f (%.2f%%)\n"+
		"Close Reason: %s\n"+
		"Close Time: %s",
		posData.Symbol,
		posData.StrategyID,
		posData.Symbol,
		posData.EntryPrice,
		posData.ExitPrice,
		posData.Quantity,
		pnlStr,
		posData.ProfitAndLoss,
		posData.ProfitAndLossPercent,
		posData.CloseReason,
		posData.Timestamp.Format(time.RFC3339))

	// Use Strategy's own reply mechanism for notification
	s.replyMsg(ctx, session, message)

	// Store this in session for later use
	session.SetAttribute("last_closed_position", posData)

	// Add a message to the chat
	s.stashMsg(ctx, session, fmt.Sprintf("📊 Position closed for %s with %s: %.2f (%.2f%%)",
		posData.Symbol, pnlStr, posData.ProfitAndLoss, posData.ProfitAndLossPercent))
}

// processMemoryOutput processes memory output from AI and saves it
func (s *Strategy) processMemoryOutput(ctx context.Context, chatSession ttypes.ISession, memory *ttypes.Memory) {
	if memory == nil || memory.Content == "" {
		return
	}

	// AI outputs complete memory content, so we replace the entire memory
	// Save memory and get truncation information
	savedMemory, wasTruncated, err := s.memoryManager.SaveMemory(memory.Content)
	if err != nil {
		log.WithError(err).Error("Failed to save memory")
		s.replyMsg(ctx, chatSession, fmt.Sprintf("Memory save failed: %s", err.Error()))
		return
	}

	// Update current memory
	s.currentMemory = savedMemory

	// Provide different feedback based on whether content was truncated
	if wasTruncated {
		// Truncated case: provide warning and word limit information
		wordCount := len(strings.Fields(memory.Content))
		limitInfo := s.memoryManager.GetWordLimitInfo()

		warningMsg := fmt.Sprintf("⚠️ Memory saved but truncated!\n"+
			"Current memory word count: %d words\n"+
			"%s\n"+
			"Please keep memory content concise in future outputs.\n"+
			"Memory content: %s",
			wordCount, limitInfo, memory.Content)

		s.replyMsg(ctx, chatSession, warningMsg)

		// Store word limit info in session for next AI decision reference
		s.stashMsg(ctx, chatSession, fmt.Sprintf("Memory word limit reminder: %s", limitInfo))

	} else {
		// Normal save case
		s.replyMsg(ctx, chatSession, fmt.Sprintf("💾 Memory saved: %s", memory.Content))
	}
}

// executeNextCycleCommands executes pending commands from previous cycle
func (s *Strategy) executeNextCycleCommands(ctx context.Context, session ttypes.ISession) {
	commands, err := s.commandMemory.LoadPendingCommands()
	if err != nil {
		log.WithError(err).Warn("Failed to load pending commands")
		return
	}

	if len(commands) == 0 {
		return
	}

	s.replyMsg(ctx, session, fmt.Sprintf("📋 Executing %d pending commands from previous cycle...", len(commands)))

	for _, cmd := range commands {
		// Check for context cancellation between iterations
		select {
		case <-ctx.Done():
			log.Warn("Context cancelled, stopping command execution")
			return
		default:
		}

		if cmd.Status != "pending" && !(cmd.Status == "failed" && cmd.RetryCount < cmd.MaxRetries) {
			continue
		}

		// Execute command with timeout
		err := s.executeCommand(ctx, session, cmd)
		if err != nil {
			log.WithError(err).WithField("command", cmd).Error("Command execution failed")
			cmd.Status = "failed"
			cmd.Error = err.Error()
			cmd.RetryCount++

			if cmd.RetryCount >= cmd.MaxRetries {
				s.replyMsg(ctx, session, fmt.Sprintf("❌ Command failed permanently: %s.%s - %s", cmd.EntityID, cmd.CommandName, err.Error()))
			} else {
				s.replyMsg(ctx, session, fmt.Sprintf("⚠️ Command failed (retry %d/%d): %s.%s", cmd.RetryCount, cmd.MaxRetries, cmd.EntityID, cmd.CommandName))
			}
		} else {
			cmd.Status = "completed"
			s.replyMsg(ctx, session, fmt.Sprintf("✅ Command executed successfully: %s.%s", cmd.EntityID, cmd.CommandName))
		}

		cmd.UpdatedAt = time.Now()
	}

	// Save updated command statuses
	if err := s.commandMemory.SaveCommands(commands); err != nil {
		log.WithError(err).Error("Failed to save command statuses")
	}

	// Archive old completed/failed commands
	if err := s.commandMemory.ArchiveCompletedCommands(); err != nil {
		log.WithError(err).Warn("Failed to archive commands")
	}
}

// executeCommand executes a single command with timeout
func (s *Strategy) executeCommand(ctx context.Context, session ttypes.ISession, cmd *memory.PendingCommand) error {
	// Set 30-second timeout for command execution
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build full command name
	fullCommandName := cmd.EntityID + "." + cmd.CommandName

	// Execute via world.SendCommand
	return s.world.SendCommand(cmdCtx, fullCommandName, cmd.Args)
}

const MaxCommandsPerCycle = 10

// processNextCommands processes next_commands from AI output and saves them for next cycle
func (s *Strategy) processNextCommands(ctx context.Context, session ttypes.ISession, nextCommands []*ttypes.NextCommand) {
	if len(nextCommands) == 0 {
		return
	}

	// Enforce command count limit to prevent resource exhaustion
	if len(nextCommands) > MaxCommandsPerCycle {
		s.replyMsg(ctx, session, fmt.Sprintf("⚠️ Too many commands scheduled (%d), limiting to %d",
			len(nextCommands), MaxCommandsPerCycle))
		nextCommands = nextCommands[:MaxCommandsPerCycle]
	}

	s.replyMsg(ctx, session, fmt.Sprintf("📝 Scheduling %d commands for next cycle...", len(nextCommands)))

	pendingCommands := make([]*memory.PendingCommand, 0)

	for _, nc := range nextCommands {
		// Validate entity exists
		entity := s.world.GetEntity(nc.EntityID)
		if entity == nil {
			log.WithField("entityID", nc.EntityID).Warn("Entity not found, skipping command")
			s.replyMsg(ctx, session, fmt.Sprintf("⚠️ Skipping command: entity '%s' not found", nc.EntityID))
			continue
		}

		// Optional: Validate command is supported by entity
		if !s.isCommandSupported(entity, nc.CommandName) {
			log.WithField("command", nc.CommandName).WithField("entityID", nc.EntityID).Warn("Command not supported by entity")
			s.replyMsg(ctx, session, fmt.Sprintf("⚠️ Skipping command: '%s' not supported by entity '%s'", nc.CommandName, nc.EntityID))
			continue
		}

		// Create pending command
		cmd := &memory.PendingCommand{
			ID:          uuid.NewString(),
			EntityID:    nc.EntityID,
			CommandName: nc.CommandName,
			Args:        nc.Args,
			Status:      "pending",
			RetryCount:  0,
			MaxRetries:  1, // Default retry once
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		pendingCommands = append(pendingCommands, cmd)
	}

	// Save commands to file
	if len(pendingCommands) > 0 {
		err := s.commandMemory.SaveCommands(pendingCommands)
		if err != nil {
			log.WithError(err).Error("Failed to save next commands")
			s.replyMsg(ctx, session, "⚠️ Failed to save commands for next cycle")
		} else {
			s.replyMsg(ctx, session, fmt.Sprintf("💾 Saved %d commands for next cycle", len(pendingCommands)))
		}
	}
}

// isCommandSupported checks if an entity supports a specific command
func (s *Strategy) isCommandSupported(entity env.IEntity, commandName string) bool {
	actions := entity.Actions()
	for _, action := range actions {
		if action.Name == commandName {
			return true
		}
	}
	return false
}
