package pkg

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/agent/openai"
	"github.com/yubing744/trading-bot/pkg/chat"
	"github.com/yubing744/trading-bot/pkg/chat/feishu"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/env/exchange"
	ttypes "github.com/yubing744/trading-bot/pkg/types"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/types"
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

	// setup env
	env := env.NewEnvironment()
	env.RegisterEntity(exchange.NewExchangeEntity(
		"exchange",
		s.Env.ExchangeConfig,
		s.session,
		s.orderExecutor,
		s.Position,
	))
	err := env.Start(ctx)
	if err != nil {
		log.WithError(err).Error("error in start env")
	}

	// setup agent
	openaiCfg := &s.Agent.OpenAI
	if openaiCfg != nil {
		openaiCfg.Token = os.Getenv("AGENT_OPENAI_TOKEN")
	}

	agent := openai.NewOpenAIAgent(&s.Agent.OpenAI)
	agent.SetBackgroup("以下是和股票交易助手的对话，股票交易助手支持注册实体，支持输出指令控制实体，支持根据股价数据生成K线形态。")
	agent.RegisterActions(ctx, []*ttypes.ActionDesc{
		{
			Name:        "buy",
			Description: "购买指令",
			Samples: []string{
				"1.0 1.1 1.2 1.3 1.4 1.5 1.6",
				"1.0 1.1 1.0 1.1 1.2 1.3 1.4",
			},
		},
		{
			Name:        "sell",
			Description: "卖出指令",
			Samples: []string{
				"1.6 1.5 1.4 1.3 1.2 1.1 1.0",
				"1.0 1.1 1.2 1.3 1.4 1.3 1.2",
			},
		},
		{
			Name:        "hold",
			Description: "持仓",
			Samples: []string{
				"1.2 1.3 1.4 1.5 1.6 1.7 1.8",
			},
		},
	})

	// set chat provider
	feishuCfg := s.Chat.Feishu
	if feishuCfg != nil && os.Getenv("CHAT_FEISHU_APP_ID") != "" {
		feishuCfg.AppId = os.Getenv("CHAT_FEISHU_APP_ID")
		feishuCfg.AppSecret = os.Getenv("CHAT_FEISHU_APP_SECRET")
		feishuCfg.EventEncryptKey = os.Getenv("CHAT_FEISHU_EVENT_ENCRYPT_KEY")
		feishuCfg.VerificationToken = os.Getenv("CHAT_FEISHU_VERIFICATION_TOKEN")
	}

	chatProvider := feishu.NewFeishuChatProvider(feishuCfg)

	// init chat provider and start session
	err = chatProvider.Listen(func(ch ttypes.Channel) {
		log.WithField("channel", ch).Info("new channel")

		session := chat.NewChatSession(ch, agent, env)

		err := session.Start()
		if err != nil {
			log.
				WithError(err).
				WithField("session", session).
				Error("start session error")
		}
	})
	if err != nil {
		log.WithError(err).Error("listen chat error")
	}

	return nil
}
