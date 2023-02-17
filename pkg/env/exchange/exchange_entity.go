package exchange

import (
	"context"
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"

	ttypes "github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("entity", "exchange")

type ExchangeEntity struct {
	id       string
	symbol   string
	interval types.Interval
	leverage fixedpoint.Value

	cfg *config.EnvExchangeConfig

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor
	position      *types.Position

	Status       types.StrategyStatus
	BOLL         *indicator.BOLL
	CurrentKline *types.KLine
}

func NewExchangeEntity(
	id string,
	symbol string,
	interval types.Interval,
	leverage fixedpoint.Value,
	cfg *config.EnvExchangeConfig,
	session *bbgo.ExchangeSession,
	orderExecutor *bbgo.GeneralOrderExecutor,
	position *types.Position,
) *ExchangeEntity {
	return &ExchangeEntity{
		id:            id,
		symbol:        symbol,
		interval:      interval,
		leverage:      leverage,
		cfg:           cfg,
		session:       session,
		orderExecutor: orderExecutor,
		position:      position,
	}
}

func (ent *ExchangeEntity) GetID() string {
	return ent.id
}

func (ent *ExchangeEntity) cmdToSide(cmd string) types.SideType {
	switch cmd {
	case "buy":
		return types.SideTypeBuy
	case "sell":
		return types.SideTypeSell
	default:
		return types.SideTypeSelf
	}
}

func (ent *ExchangeEntity) HandleCommand(ctx context.Context, cmd string, args []string) error {
	log.
		WithField("cmd", cmd).
		WithField("args", args).
		Infof("entity exchange handle command")

	if ent.CurrentKline == nil {
		log.Warn("skip for current kline nil")
		return errors.New("current kline nil")
	}

	side := ent.cmdToSide(cmd)
	closePrice := ent.CurrentKline.GetClose()

	// close position if need
	// TP/SL if there's non-dust position and meets the criteria
	if !ent.position.IsDust(closePrice) {
		if (side == types.SideTypeSell && ent.position.IsLong()) || (side == types.SideTypeBuy && ent.position.IsShort()) {
			log.Infof("close existing %s position before open a new position", ent.symbol)

			err := ent.ClosePosition(ctx, fixedpoint.One, closePrice)
			if err != nil {
				return errors.Wrap(err, "close position error")
			}
		} else {
			log.Infof("existing %s position has the same direction with the signal(%s)", ent.symbol, side)
			return nil
		}
	}

	// open position
	if side == types.SideTypeSell || side == types.SideTypeBuy {
		log.Infof("open %s position for signal %v, reason: %s", ent.symbol, side, "")

		err := ent.OpenPosition(ctx, side, closePrice)
		if err != nil {
			return errors.Wrap(err, "open position error")
		}
	} else {
		log.Info("no signal")
	}

	return nil
}

func (ent *ExchangeEntity) Run(ctx context.Context, ch chan *ttypes.Event) {
	session := ent.session

	ent.Status = types.StrategyStatusRunning

	ent.setupIndicators()

	// if you need to do something when the user data stream is ready
	// note that you only receive order update, trade update, balance update when the user data stream is connect.
	session.UserDataStream.OnStart(func() {
		log.Infof("connected")
	})

	log.
		WithField("symbol", ent.symbol).
		WithField("interval", ent.interval).
		Info("exchange entity run")

	session.MarketDataStream.OnKLineClosed(types.KLineWith(ent.symbol, ent.interval, func(kline types.KLine) {
		// StrategyController
		if ent.Status != types.StrategyStatusRunning {
			log.Info("strategy status not running")
			return
		}

		ent.CurrentKline = &kline
		log.WithField("kline", kline).Info("kline closed")

		bollVal := *ent.BOLL

		ent.emitEvent(ch, &ttypes.Event{
			Type: "boll_changed",
			Data: &bollVal,
		})
	}))
}

// setupIndicators initializes indicators
func (ent *ExchangeEntity) setupIndicators() {
	log.Infof("setup indicators")

	indicators := ent.session.StandardIndicatorSet(ent.symbol)
	ent.BOLL = indicators.BOLL(types.IntervalWindow{
		Interval: ent.interval,
		Window:   20,
	}, 2)
}

func (ent *ExchangeEntity) emitEvent(ch chan *ttypes.Event, evt *ttypes.Event) {
	log.WithField("event", evt).Info("emit event")

	ch <- evt
}

func (s *ExchangeEntity) OpenPosition(ctx context.Context, side types.SideType, closePrice fixedpoint.Value) error {
	quantity := s.calculateQuantity(ctx, closePrice, side)

	for {
		if quantity.Compare(s.position.Market.MinQuantity) < 0 {
			return fmt.Errorf("%s order quantity %v is too small, less than %v", s.symbol, quantity, s.position.Market.MinQuantity)
		}

		orderForm := s.generateOrderForm(side, quantity, types.SideEffectTypeMarginBuy)
		log.Infof("submit open position order %v", orderForm)
		_, err := s.orderExecutor.SubmitOrders(ctx, orderForm)
		if err != nil {
			if strings.Contains(err.Error(), "Insufficient USDT") {
				log.WithField("quantity", quantity.Float64()).Error("Insufficient USDT, try reduce order quantity")
				quantity = quantity.Mul(fixedpoint.NewFromFloat(0.99))
				continue
			}

			log.WithError(err).Errorf("can not place %s open position order", s.symbol)
			return err
		}

		break
	}

	return nil
}

func (s *ExchangeEntity) ClosePosition(ctx context.Context, percentage fixedpoint.Value, closePrice fixedpoint.Value) error {
	if s.position.IsClosed() {
		return fmt.Errorf("no opened %s position", s.position.Symbol)
	}

	// make it negative
	quantity := s.position.GetBase().Mul(percentage).Abs()
	side := types.SideTypeBuy

	if s.position.IsLong() {
		side = types.SideTypeSell

		if quantity.Compare(s.position.Market.MinQuantity) < 0 {
			return fmt.Errorf("%s order quantity %v is too small, less than %v", s.symbol, quantity, s.position.Market.MinQuantity)
		}
	} else {
		quantity = quantity.Mul(closePrice)
	}

	orderForm := s.generateOrderForm(side, quantity, types.SideEffectTypeAutoRepay)
	if percentage.Compare(fixedpoint.One) == 0 {
		orderForm.ClosePosition = true // Full close position
	}

	bbgo.Notify("submitting %s %s order to close position by %v", s.symbol, side.String(), percentage, orderForm)

	_, err := s.orderExecutor.SubmitOrders(ctx, orderForm)
	if err != nil {
		log.WithError(err).Errorf("can not place %s position close order", s.symbol)
		bbgo.Notify("can not place %s position close order", s.symbol)
	}

	return err
}

func (s *ExchangeEntity) generateOrderForm(side types.SideType, quantity fixedpoint.Value, marginOrderSideEffect types.MarginOrderSideEffectType) types.SubmitOrder {
	orderForm := types.SubmitOrder{
		Symbol:           s.symbol,
		Market:           s.position.Market,
		Side:             side,
		Type:             types.OrderTypeMarket,
		Quantity:         quantity,
		MarginSideEffect: marginOrderSideEffect,
	}

	return orderForm
}

// calculateQuantity returns leveraged quantity
func (s *ExchangeEntity) calculateQuantity(ctx context.Context, currentPrice fixedpoint.Value, side types.SideType) fixedpoint.Value {
	quoteQty, err := bbgo.CalculateQuoteQuantity(ctx, s.session, s.position.Market.QuoteCurrency, s.leverage)
	if err != nil {
		log.WithError(err).Errorf("can not update %s quote balance from exchange", s.symbol)
		return fixedpoint.Zero
	}

	if side == types.SideTypeSell {
		return quoteQty.Div(currentPrice).
			Mul(fixedpoint.NewFromFloat(0.99))
	} else {
		return quoteQty
	}
}
