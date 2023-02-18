package pkg

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/agent/openai"
	"github.com/yubing744/trading-bot/pkg/chat"
	"github.com/yubing744/trading-bot/pkg/chat/feishu"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/env/exchange"
	"github.com/yubing744/trading-bot/pkg/utils"

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

	// setup world
	world := env.NewEnvironment()
	world.RegisterEntity(exchange.NewExchangeEntity(
		"exchange",
		s.Symbol,
		s.Interval,
		s.Leverage,
		s.Env.ExchangeConfig,
		s.session,
		s.orderExecutor,
		s.Position,
	))
	err := world.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "Error in start env")
	}
	s.world = world

	// setup agent
	openaiCfg := &s.Agent.OpenAI
	if openaiCfg != nil {
		openaiCfg.Token = os.Getenv("AGENT_OPENAI_TOKEN")
	}

	agent := openai.NewOpenAIAgent(&s.Agent.OpenAI)
	agent.SetBackgroup("以下是和交易助手的对话，交易助手支持注册实体，支持输出命令控制实体、支持分析实体产生的行情数据并生成交易信号来操作实体，实体会反馈操作结果。交易助手的目标是：最大化投资者收益。")
	agent.RegisterActions(ctx, "exchange", []*ttypes.ActionDesc{
		{
			Name:        "open_long_position",
			Description: "开启做多仓位",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"There are currently no open position",
						"KLine data changed: Open:[2.83 2.83], Close:[2.81 2.83], High:[2.83 2.83], Low:[2.81 2.83], Volume:[27097.45 19859.13]",
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90 2.92], SMA:[2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85 2.86], DownBand:[2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81 2.80]",
						"VWMA data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"Analyze data, generate trading signals and explain why.",
					},
					Output: []string{
						"/open_long_position [] #原因：BOLL上轨和中轨开始上升",
					},
				},
			},
		},
		{
			Name:        "open_short_position",
			Description: "开启做空仓位",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"The current position is short, and average cost: 2.84",
						"KLine data changed: Open:[2.83 2.83], Close:[2.81 2.83], High:[2.83 2.83], Low:[2.81 2.83], Volume:[27097.45 19859.13]",
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.90 2.92 2.94 2.94 2.94 2.95 2.95 2.96], SMA:[2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.86 2.86 2.86 2.87 2.87 2.87 2.88 2.88], DownBand:[2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.81 2.80 2.79 2.79 2.79 2.80 2.80 2.80]}",
						"VWMA data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"Analyze data, generate trading signals and explain why.",
					},
					Output: []string{
						"/open_short_position [] #原因：BOLL上轨和中轨开始下降",
					},
				},
			},
		},
		{
			Name:        "close_position",
			Description: "关闭多空仓位",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"The current position is long, and average cost: 2.80",
						"KLine data changed: Open:[2.83 2.83], Close:[2.81 2.83], High:[2.83 2.83], Low:[2.81 2.83], Volume:[27097.45 19859.13]",
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90], SMA:[2.86 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85], DownBand:[2.80 2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81]",
						"VWMA data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"Analyze data, generate trading signals and explain why.",
					},
					Output: []string{
						"/close_position [] #原因：趋势反转",
					},
				},
			},
		},
		{
			Name:        "hold",
			Description: "不操作",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"The current position is long, and average cost: 2.80",
						"KLine data changed: Open:[2.83 2.83], Close:[2.81 2.83], High:[2.83 2.83], Low:[2.81 2.83], Volume:[27097.45 19859.13]",
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90], SMA:[2.86 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85], DownBand:[2.80 2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81]",
						"VWMA data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"Analyze data, generate trading signals and explain why.",
					},
					Output: []string{
						"/hold [] #原因：当前趋势和当前仓位一致",
					},
				},
			},
		},
	})
	s.agent = agent

	// set chats
	feishuCfg := s.Chat.Feishu
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
		err = chatProvider.Listen(func(ch ttypes.IChannel) {
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

func (s *Strategy) notifyMsg(ctx context.Context, msg string) {
	err := s.chatSessions.Notify(ctx, &ttypes.Message{
		ID:   uuid.NewString(),
		Text: msg,
	})
	if err != nil {
		log.WithError(err).Error("notify message error")
	}
}

func (s *Strategy) agentAction(ctx context.Context, chatSession ttypes.ISession, msgs []*ttypes.Message) {
	result, err := s.agent.GenActions(ctx, chatSession, msgs)
	if err != nil {
		log.WithError(err).Error("gen action error")
		s.replyMsg(ctx, chatSession, fmt.Sprintf("gen action error: %s", err.Error()))
		return
	}

	log.WithField("result", result).Info("gen actions result")

	if len(result.Texts) > 0 {
		s.replyMsg(ctx, chatSession, strings.Join(result.Texts, ""))
	}

	if len(result.Actions) > 0 {
		if chatSession.HasRole(ttypes.RoleAdmin) {
			for _, action := range result.Actions {
				err := s.world.SendCommand(ctx, action.Target, action.Name, action.Args)
				if err != nil {
					log.WithError(err).Error("env send cmd error")
					s.replyMsg(ctx, chatSession, fmt.Sprintf("cmd /%s [%s] handle fail: %s", action.Name, strings.Join(action.Args, ","), err.Error()))
				} else {
					s.replyMsg(ctx, chatSession, fmt.Sprintf("cmd /%s [%s] handle succes", action.Name, strings.Join(action.Args, ",")))
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
		msg = fmt.Sprintf("The current position is long, and average cost: %.4f", position.AverageCost.Float64())
	} else if position.IsShort() {
		msg = fmt.Sprintf("The current position is short, and average cost: %.4f", position.AverageCost.Float64())
	} else {
		msg = "There are currently no open positions"
	}

	s.replyMsg(ctx, session, msg)
	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleKlineChanged(ctx context.Context, session ttypes.ISession, klineWindow *types.KLineWindow) {
	log.WithField("kline", klineWindow).Info("handle klineWindow values changed")

	msg := fmt.Sprintf("KLine data changed: Open:[%s], Close:[%s], High:[%s], Low:[%s], Volume:[%s]",
		utils.JoinFloatSeries(klineWindow.Open(), " "),
		utils.JoinFloatSeries(klineWindow.Close(), " "),
		utils.JoinFloatSeries(klineWindow.High(), " "),
		utils.JoinFloatSeries(klineWindow.Low(), " "),
		utils.JoinFloatSeries(klineWindow.Volume(), " "),
	)

	s.replyMsg(ctx, session, msg)
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

	s.replyMsg(ctx, session, msg)
	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleRSIChanged(ctx context.Context, session ttypes.ISession, rsi *indicator.RSI) {
	log.WithField("rsi", rsi).Info("handle vwma values changed")

	vals := rsi.Values
	if len(vals) > s.MaxWindowSize {
		vals = vals[len(vals)-s.MaxWindowSize:]
	}

	msg := fmt.Sprintf("RSI data changed: [%s]",
		utils.JoinFloatSlice([]float64(vals), " "),
	)

	s.replyMsg(ctx, session, msg)
	s.stashMsg(ctx, session, msg)
}

func (s *Strategy) handleUpdateFinish(ctx context.Context, session ttypes.ISession) {
	tempMsgs, ok := s.popMsgs(ctx, session)
	log.WithField("tempMsgs", tempMsgs).Info("session tmp msgs")

	if ok {
		msg := "Analyze the data and generate a trading signal."
		s.replyMsg(ctx, session, msg)

		tempMsgs = append(tempMsgs, &ttypes.Message{
			Text: msg,
		})

		s.agentAction(ctx, session, tempMsgs)
	}

	session.SetState(nil)
}

func (s *Strategy) stashMsg(ctx context.Context, session ttypes.ISession, msg string) {
	tempMsgs, _ := session.GetState().([]*ttypes.Message)
	tempMsgs = append(tempMsgs, &ttypes.Message{
		Text: msg,
	})

	log.WithField("tempMsgs", tempMsgs).Info("session tmp msgs")
	session.SetState(tempMsgs)
}

func (s *Strategy) popMsgs(ctx context.Context, session ttypes.ISession) ([]*ttypes.Message, bool) {
	tempMsgs, ok := session.GetState().([]*ttypes.Message)
	if ok {
		return tempMsgs, ok
	}

	return []*ttypes.Message{}, false
}
