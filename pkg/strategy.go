package demo

import (
	"context"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
)

// ID is the unique strategy ID, it needs to be in all lower case
// For example, grid strategy uses "grid"
const ID = "trading-bot"

// log is a logrus.Entry that will be reused.
// This line attaches the strategy field to the logger with our ID, so that the logs from this strategy will be tagged with our ID
var log = logrus.WithField("trading-bot", ID)

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
	Counter int `json:"counter,omitempty"`
}

// Strategy is a struct that contains the settings of your strategy.
// These settings will be loaded from the BBGO YAML config file "bbgo.yaml" automatically.
type Strategy struct {
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
	s.State = &State{
		Counter: 0,
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

		log.Debug("")
		log.WithField("counter", s.State.Counter).
			Debug("boll values")

		if kline.Symbol != s.Symbol {
			log.Panic("kline symbol not equal strategy symbol")
			return
		}

		bbgo.Sync(ctx, s)

		closePrice := kline.GetClose()

		if s.State.Counter%10 == 0 {
			side := types.SideTypeBuy
			orderQty := s.calculateQuantity(ctx, closePrice, side)

			for {
				if orderQty.Compare(fixedpoint.NewFromFloat(1)) < 0 {
					log.Errorf("can not place %s open position order, orderQty too small %f", s.Symbol, orderQty.Float64())
					break
				}

				orderForm := s.generateOrderForm(side, orderQty, types.SideEffectTypeMarginBuy)
				log.Infof("submit open position order %v", orderForm)
				_, err := s.orderExecutor.SubmitOrders(ctx, orderForm)
				if err != nil {
					if strings.Contains(err.Error(), "Insufficient USDT") {
						log.WithField("orderQty", orderQty.Float64()).Error("Insufficient USDT, try reduce orderQty")
						orderQty = orderQty.Mul(fixedpoint.NewFromFloat(0.99))
						continue
					}

					log.WithError(err).Errorf("can not place %s open position order", s.Symbol)
					log.Infof("can not place %s open position order", s.Symbol)
					return
				}

				break
			}

		} else if s.State.Counter%10 == 3 {
			perc := fixedpoint.One
			err := s.orderExecutor.ClosePosition(ctx, perc)
			if err != nil {
				log.WithError(err).Error("close position error")
			}
		}

		// Update our counter and sync the changes to the persistence layer on time
		// If you don't do this, BBGO will sync it automatically when BBGO shuts down.
		s.State.Counter++
	}))

	// Graceful shutdown
	bbgo.OnShutdown(ctx, func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		_ = s.orderExecutor.GracefulCancel(ctx)
	})

	return nil
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

func (s *Strategy) generateOrderForm(side types.SideType, quantity fixedpoint.Value, marginOrderSideEffect types.MarginOrderSideEffectType) types.SubmitOrder {
	orderForm := types.SubmitOrder{
		Symbol:           s.Symbol,
		Market:           s.Market,
		Side:             side,
		Type:             types.OrderTypeMarket,
		Quantity:         quantity,
		MarginSideEffect: marginOrderSideEffect,
	}

	return orderForm
}

// calculateQuantity returns leveraged quantity
func (s *Strategy) calculateQuantity(ctx context.Context, currentPrice fixedpoint.Value, side types.SideType) fixedpoint.Value {
	// Quantity takes precedence
	if !s.Quantity.IsZero() {
		return s.Quantity
	}

	quoteQty, err := bbgo.CalculateQuoteQuantity(ctx, s.session, s.Market.QuoteCurrency, s.Leverage)
	if err != nil {
		log.WithError(err).Errorf("can not update %s quote balance from exchange", s.Symbol)
		return fixedpoint.Zero
	}

	if side == types.SideTypeSell {
		return quoteQty.Div(currentPrice)
	} else {
		return quoteQty
	}
}
