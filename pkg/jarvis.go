package pkg

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/chat/feishu"
	"github.com/yubing744/trading-bot/pkg/config"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
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

// State is a struct contains the information that we want to keep in the persistence layer,
// for example, redis or json file.
type State struct {
	Counter       int             `json:"counter,omitempty"`
	Delay         int             `json:"delay,omitempty"`
	LastTrend     types.Direction `json:"last_trend,omitempty"`
	LastIndex     int             `json:"last_index,omitempty"`
	LastOpenIndex int             `json:"last_open_index,omitempty"`
	LastHighIndex int             `json:"last_high_index,omitempty"`
	LastLowIndex  int             `json:"last_low_index,omitempty"`
}

// Strategy is a struct that contains the settings of your strategy.
// These settings will be loaded from the BBGO YAML config file "bbgo.yaml" automatically.
type Strategy struct {
	config.Config
	Environment *bbgo.Environment
	Market      types.Market

	Symbol   string         `json:"symbol"`
	Interval types.Interval `json:"interval"`
	// Leverage uses the account net value to calculate the order qty
	Leverage fixedpoint.Value `json:"leverage"`
	// Quantity sets the fixed order qty, takes precedence over Leverage
	Quantity fixedpoint.Value `json:"quantity"`
	// The tread threshold
	TrendThreshold fixedpoint.Value `json:"trend_threshold"`

	// State is a state of your strategy
	// When BBGO shuts down, everything in the memory will be dropped
	// If you need to store something and restore this information back,
	// Simply define the "persistence" tag
	State *State `persistence:"state"`
	// persistence fields
	Position *types.Position `persistence:"position"`

	BOLL *indicator.BOLL

	session                *bbgo.ExchangeSession
	orderExecutor          *bbgo.GeneralOrderExecutor
	currentTakeProfitPrice fixedpoint.Value
	currentStopLossPrice   fixedpoint.Value

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

	s.currentStopLossPrice = fixedpoint.Zero
	s.currentTakeProfitPrice = fixedpoint.Zero

	// calculate group id for orders
	instanceID := s.InstanceID()

	// Initialize the default value for state
	if s.State == nil {
		s.State = &State{
			Counter:       1,
			Delay:         0,
			LastTrend:     types.DirectionNone,
			LastIndex:     0,
			LastOpenIndex: 0,
			LastHighIndex: 0,
			LastLowIndex:  0,
		}
	}

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

	s.Position.OnModify(func(baseQty, quoteQty, price fixedpoint.Value) {
		log.WithField("price", price).Info("update currentStopLossPrice")
		s.currentStopLossPrice = price
	})

	// StrategyController
	s.Status = types.StrategyStatusRunning
	s.OnSuspend(func() {
		_ = s.orderExecutor.GracefulCancel(ctx)
		bbgo.Sync(ctx, s)
	})
	s.OnEmergencyStop(func() {
		_ = s.orderExecutor.GracefulCancel(ctx)
		// Close 100% position
		//_ = s.ClosePosition(ctx, fixedpoint.One)
	})

	s.setupBot()
	s.setupIndicators()

	// if you need to do something when the user data stream is ready
	// note that you only receive order update, trade update, balance update when the user data stream is connect.
	session.UserDataStream.OnStart(func() {
		log.Infof("connected")
	})

	session.MarketDataStream.OnKLineClosed(types.KLineWith(s.Symbol, s.Interval, func(kline types.KLine) {
		// StrategyController
		if s.Status != types.StrategyStatusRunning {
			log.Info("strategy status not running")
			return
		}

		bollDown := s.BOLL.LastDownBand()
		bollUp := s.BOLL.LastUpBand()
		bollMid := (bollDown + bollUp) / 2
		section := (bollUp - bollDown) / 10
		sma := s.BOLL.SMA

		log.Debug("")
		log.WithField("bollDown", bollDown).
			WithField("bollMid", bollMid).
			WithField("bollUp", bollUp).
			WithField("time", kline.GetStartTime()).
			Debug("boll values")

		if kline.Symbol != s.Symbol {
			log.Panic("kline symbol not equal strategy symbol")
			return
		}

		bbgo.Sync(ctx, s)

		// Call LastPrice(symbol) If you need to get the latest price
		// Note this last price is updated by the closed kline
		closePrice := kline.Close
		index := int((kline.Close.Float64() - bollMid) / section)
		openIndex := int((kline.Open.Float64() - bollMid) / section)
		lowIndex := int((kline.Low.Float64() - bollMid) / section)
		highIndex := int((kline.High.Float64() - bollMid) / section)
		trend := s.getTrend(sma)
		state := s.State

		log.WithField("index", index).
			WithField("open_index", openIndex).
			WithField("low_index", lowIndex).
			WithField("high_index", highIndex).
			WithField("delay", s.State.Delay).
			WithField("counter", s.State.Counter).
			WithField("trend", trend).
			WithField("time", kline.GetStartTime()).
			Infof("current state")

		log.WithField("last_index", s.State.LastIndex).
			WithField("last_open", s.State.LastOpenIndex).
			WithField("last_low", s.State.LastLowIndex).
			WithField("last_high", s.State.LastHighIndex).
			WithField("last_trend", s.State.LastTrend).
			WithField("time", kline.GetStartTime()).
			Infof("last state")

		log.WithField("position", s.Position).
			Infof("current position")

		// TP/SL if there's non-dust position and meets the criteria
		if !s.Market.IsDustQuantity(s.Position.GetBase().Abs(), closePrice) {
			shouldStop, delay, reason := s.shouldStop(kline, state, trend, lowIndex, openIndex, index, highIndex)
			if shouldStop {
				log.WithField("reason", reason).Info("close position")

				err := s.orderExecutor.ClosePosition(ctx, fixedpoint.One, "stop")
				if err != nil {
					log.WithError(err).Error("close position error")
				} else {
					s.State.Delay = delay
					s.currentStopLossPrice = fixedpoint.Zero
					s.currentTakeProfitPrice = fixedpoint.Zero
				}
			} else {
				log.Info("skip close position")
			}
		} else {
			log.Debug("market dust quantity")
		}
	}))

	// Graceful shutdown
	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		_ = s.orderExecutor.GracefulCancel(ctx)
	})

	return nil
}

func (s *Strategy) setupBot() {
	feishuCfg := s.Chat.Feishu
	if feishuCfg != nil && os.Getenv("CHAT_FEISHU_APP_ID") != "" {
		feishuCfg.AppId = os.Getenv("CHAT_FEISHU_APP_ID")
		feishuCfg.AppSecret = os.Getenv("CHAT_FEISHU_APP_SECRET")
		feishuCfg.EventEncryptKey = os.Getenv("CHAT_FEISHU_EVENT_ENCRYPT_KEY")
		feishuCfg.VerificationToken = os.Getenv("CHAT_FEISHU_VERIFICATION_TOKEN")
	}

	chat := feishu.NewFeishuChatProvider(feishuCfg)

	go func() {
		err := chat.Start()
		if err != nil {
			log.WithError(err).Error("feishu chat start error")
		}
	}()
}

// setupIndicators initializes indicators
func (s *Strategy) setupIndicators() {
	log.Infof("setupIndicators")

	indicators := s.session.StandardIndicatorSet(s.Symbol)
	s.BOLL = indicators.BOLL(types.IntervalWindow{
		Interval: s.Interval,
		Window:   20,
	}, 2)
}

func (s *Strategy) getTrend(line types.SeriesExtend) types.Direction {
	if line.Length() > 2 {
		last := line.Index(0)
		first := line.Index(2)
		rate := (last - first) * 100 / first
		log.WithField("rate", rate).Debug("get trend")

		if rate > s.TrendThreshold.Float64() {
			return types.DirectionUp
		} else if rate < s.TrendThreshold.Neg().Float64() {
			return types.DirectionDown
		}
	}

	return types.DirectionNone
}

func (s *Strategy) shouldStop(kline types.KLine, state *State, trend types.Direction, lowIndex int, openIndex int, closeIndex int, highIndex int) (bool, int, string) {
	stopNow := false
	delay := 0
	stopReason := ""

	if !s.currentStopLossPrice.IsZero() {
		if s.Position.IsShort() && kline.GetClose().Compare(s.currentStopLossPrice.Mul(fixedpoint.NewFromFloat(1.01))) > 0 {
			stopNow = true
			delay = 15
			stopReason = fmt.Sprintf("%s stop loss by triggering the kline high", s.Symbol)
			return stopNow, delay, stopReason
		} else if s.Position.IsLong() && kline.GetClose().Compare(s.currentStopLossPrice.Mul(fixedpoint.NewFromFloat(0.99))) < 0 {
			stopNow = true
			delay = 15
			stopReason = fmt.Sprintf("%s stop loss by triggering the kline low", s.Symbol)
			return stopNow, delay, stopReason
		}
	}

	if s.Position.IsLong() {
		// Long
		if highIndex >= 5 {
			stopNow = true
			delay = 0
			stopReason = "high price rebound to upper boundary"
		}
	} else if s.Position.IsShort() {
		// short
		if lowIndex <= -5 {
			stopNow = true
			delay = 0
			stopReason = "low price rebound to lower boundary"
		}
	}

	return stopNow, delay, stopReason
}
