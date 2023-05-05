package exchange

import (
	"context"
	"fmt"
	"strings"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/utils"

	ttypes "github.com/yubing744/trading-gpt/pkg/types"
)

var log = logrus.WithField("entity", "exchange")

type ExchangeEntity struct {
	symbol   string
	interval types.Interval
	leverage fixedpoint.Value

	cfg *config.EnvExchangeConfig

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor
	position      *types.Position

	Status      types.StrategyStatus
	BOLL        *indicator.BOLL
	RSI         *indicator.RSI
	KLineWindow *types.KLineWindow

	vm *goja.Runtime
}

func NewExchangeEntity(
	symbol string,
	interval types.Interval,
	leverage fixedpoint.Value,
	cfg *config.EnvExchangeConfig,
	session *bbgo.ExchangeSession,
	orderExecutor *bbgo.GeneralOrderExecutor,
	position *types.Position,
) *ExchangeEntity {
	return &ExchangeEntity{
		symbol:        symbol,
		interval:      interval,
		leverage:      leverage,
		cfg:           cfg,
		session:       session,
		orderExecutor: orderExecutor,
		position:      position,
		vm:            goja.New(),
	}
}

func (ent *ExchangeEntity) GetID() string {
	return "exchange"
}

func (ent *ExchangeEntity) Actions() []*ttypes.ActionDesc {
	return []*ttypes.ActionDesc{
		{
			Name:        "open_long_position",
			Description: "open long position",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "stop_loss",
					Description: "stop-loss trigger price",
				},
				{
					Name:        "take_profit",
					Description: "take-profit trigger price",
				},
			},
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90 2.92], SMA:[2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85 2.86], DownBand:[2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81 2.80]",
						"RSI data changed: [73.454 41.980 25.516 17.727 32.413 18.679 8.576 42.228 29.611 36.948 57.658 46.181 61.506 77.894 76.378 44.059 35.556 50.472 56.603 60.012]",
						"There are currently no open position",
						"Analyze data, generate trading cmd",
					},
					Output: []string{
						"Execute cmd: /open_long_position",
					},
				},
			},
		},
		{
			Name:        "open_short_position",
			Description: "open short position",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "stop_loss",
					Description: "stop-loss trigger price",
				},
				{
					Name:        "take_profit",
					Description: "take-profit trigger price",
				},
			},
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90 2.92], SMA:[2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85 2.86], DownBand:[2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81 2.80]",
						"RSI data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"The current position is short, and average cost: 2.84",
						"Analyze data, generate trading cmd",
					},
					Output: []string{
						"Execute cmd: /open_short_position",
					},
				},
			},
		},
		{
			Name:        "close_position",
			Description: "close position",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"BOLL data changed: UpBand:[2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.92 2.91 2.91 2.90 2.90 2.89 2.89 2.89 2.89 2.89 2.89 2.90 2.92], SMA:[2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.87 2.86 2.86 2.86 2.85 2.85 2.85 2.85 2.85 2.85 2.85 2.86], DownBand:[2.81 2.81 2.82 2.82 2.82 2.82 2.83 2.83 2.82 2.82 2.82 2.81 2.81 2.82 2.82 2.82 2.82 2.82 2.82 2.81 2.80]",
						"RSI data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"The current position is long, average cost: 2.736, and accumulated profit: 15.324",
						"Analyze data, generate trading cmd",
					},
					Output: []string{
						"Execute cmd: /close_position",
					},
				},
			},
		},
		{
			Name:        "no_action",
			Description: "No action to be taken",
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"RSI data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"The current position is long, and average cost: 2.80",
						"Analyze data, generate trading cmd",
					},
					Output: []string{
						"Execute cmd: /no_action",
					},
				},
			},
		},
	}
}

func (ent *ExchangeEntity) cmdToSide(cmd string) types.SideType {
	switch cmd {
	case "open_long_position":
		return types.SideTypeBuy
	case "open_short_position":
		return types.SideTypeSell
	default:
		return types.SideTypeSelf
	}
}

func (ent *ExchangeEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	log.
		WithField("cmd", cmd).
		WithField("args", args).
		Infof("entity exchange handle command")

	if ent.KLineWindow == nil {
		log.Warn("skip for current kline nil")
		return errors.New("current kline nil")
	}

	closePrice := ent.KLineWindow.GetClose()

	// close position if need
	if cmd == "close_position" {
		// TP/SL if there's non-dust position and meets the criteria
		if !ent.position.IsDust(closePrice) {
			if ent.position.IsShort() || ent.position.IsLong() {
				log.Infof("close existing %s position", ent.symbol)

				err := ent.ClosePosition(ctx, fixedpoint.One, closePrice)
				if err != nil {
					return errors.Wrap(err, "close position error")
				}
			} else {
				return errors.New("no existing open position")
			}
		} else {
			log.Debug("market dust quantity")
		}

		return nil
	}

	// open position
	if cmd == "open_long_position" || cmd == "open_short_position" {
		side := ent.cmdToSide(cmd)

		// Close opposite position if any
		if !ent.position.IsDust(closePrice) {
			if (side == types.SideTypeSell && ent.position.IsLong()) || (side == types.SideTypeBuy && ent.position.IsShort()) {
				log.Infof("close existing %s position before open a new position", ent.symbol)
				err := ent.ClosePosition(ctx, fixedpoint.One, closePrice)
				if err != nil {
					return errors.Wrap(err, "close existing position error")
				}
			} else {
				return errors.Errorf("existing %s position has the same direction with the signal", ent.symbol)
			}
		}

		log.Infof("open %s position for signal %v, reason: %s", ent.symbol, side, "")

		opts := make([]interface{}, 0)

		// config stop losss
		if stopLoss, ok := args["stop_loss"]; ok {
			stopLoss, err := utils.ParseStopLoss(ent.vm, side, closePrice, stopLoss)
			if err != nil {
				return errors.Wrapf(err, "the stop loss invalid: %s", stopLoss)
			}

			if stopLoss != nil {
				opts = append(opts, &StopLossPrice{
					Value: *stopLoss,
				})
			}
		}

		// config take profix
		if takeProfix, ok := args["take_profix"]; ok {
			takeProfix, err := utils.ParseTakeProfit(ent.vm, side, closePrice, takeProfix)
			if err != nil {
				return errors.Wrapf(err, "the take profit invalid: %s", takeProfix)
			}

			if takeProfix != nil {
				opts = append(opts, &TakeProfitPrice{
					Value: *takeProfix,
				})
			}
		}

		err := ent.OpenPosition(ctx, side, closePrice, opts...)
		if err != nil {
			return errors.Wrap(err, "open position error")
		}

		return nil
	}

	log.Info("no signal")

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

		// Update Kline
		if ent.KLineWindow != nil {
			ent.KLineWindow.Add(kline)

			if ent.KLineWindow.Len() > ent.cfg.WindowSize {
				ent.KLineWindow.Truncate(ent.cfg.WindowSize)
			}
		}

		// Update postion accumulated Profit
		if ent.position != nil {
			ent.position.AccumulatedProfit = kline.GetClose().Sub(ent.position.AverageCost).Div(ent.position.AverageCost).Mul(fixedpoint.NewFromFloat(100.0))
		}

		log.WithField("kline", kline).Info("kline closed")

		ent.emitEvent(ch, &ttypes.Event{
			Type: "kline_changed",
			Data: ent.KLineWindow,
		})

		ent.emitEvent(ch, &ttypes.Event{
			Type: "boll_changed",
			Data: ent.BOLL,
		})

		ent.emitEvent(ch, &ttypes.Event{
			Type: "rsi_changed",
			Data: ent.RSI,
		})

		ent.emitEvent(ch, &ttypes.Event{
			Type: "position_changed",
			Data: ent.position,
		})

		ent.emitEvent(ch, &ttypes.Event{
			Type: "update_finish",
		})
	}))
}

// setupIndicators initializes indicators
func (ent *ExchangeEntity) setupIndicators() {
	log.WithField("WindowSize", ent.cfg.WindowSize).Infof("setup indicators")

	// set kline window
	inc := &types.KLineWindow{}
	dataStore, ok := ent.session.MarketDataStore(ent.symbol)
	if ok {
		if klines, ok := dataStore.KLinesOfInterval(ent.interval); ok {
			for _, k := range *klines {
				inc.Add(k)
			}
		}
	}
	ent.KLineWindow = inc

	// setup BOLL
	indicators := ent.session.StandardIndicatorSet(ent.symbol)
	ent.BOLL = indicators.BOLL(types.IntervalWindow{
		Interval: ent.interval,
		Window:   ent.cfg.WindowSize,
	}, 2)

	// setup RSI
	ent.RSI = indicators.RSI(types.IntervalWindow{
		Interval: ent.interval,
		Window:   ent.cfg.WindowSize,
	})
}

func (ent *ExchangeEntity) emitEvent(ch chan *ttypes.Event, evt *ttypes.Event) {
	ch <- evt
}

type StopLossPrice struct {
	Value fixedpoint.Value
}

type TakeProfitPrice struct {
	Value fixedpoint.Value
}

func (s *ExchangeEntity) OpenPosition(ctx context.Context, side types.SideType, closePrice fixedpoint.Value, args ...interface{}) error {
	quantity := s.calculateQuantity(ctx, closePrice, side)

	for {
		if quantity.Compare(s.position.Market.MinQuantity) < 0 {
			return fmt.Errorf("%s order quantity %v is too small, less than %v", s.symbol, quantity, s.position.Market.MinQuantity)
		}

		orderForm := s.generateOrderForm(side, quantity, types.SideEffectTypeMarginBuy)
		for _, arg := range args {
			switch val := arg.(type) {
			case *StopLossPrice:
				orderForm.StopPrice = val.Value
			case *TakeProfitPrice:
				orderForm.TakePrice = val.Value
			}
		}

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
