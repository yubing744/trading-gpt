package exchange

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
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
	position      *PositionX

	Status      types.StrategyStatus
	Indicators  []*ExchangeIndicator
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
		position:      NewPositionX(position),
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
			Description: "Open long position (supports market and limit orders; unfilled limit orders auto-cancel at next cycle)",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "order_type",
					Description: "Order type: market|limit (default: market)",
				},
				{
					Name:        "limit_price",
					Description: "Limit price or expression (required if order_type=limit, e.g., 'last_close * 0.995')",
				},
				{
					Name:        "time_in_force",
					Description: "Time in force: GTC|IOC|FOK (default: GTC)",
				},
				{
					Name:        "post_only",
					Description: "Post only: true|false (default: false, maker only)",
				},
				{
					Name:        "quote_ratio",
					Description: "Optional ratio (0-1] of available quote balance to use before leverage",
				},
				{
					Name:        "stop_loss_trigger_price",
					Description: "Stop-loss trigger price",
				},
				{
					Name:        "take_profit_trigger_price",
					Description: "Take-profit trigger price",
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
						"Execute cmd: /open_long_position quote_ratio=0.3",
					},
				},
			},
		},
		{
			Name:        "open_short_position",
			Description: "Open short position (supports market and limit orders; unfilled limit orders auto-cancel at next cycle)",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "order_type",
					Description: "Order type: market|limit (default: market)",
				},
				{
					Name:        "limit_price",
					Description: "Limit price or expression (required if order_type=limit, e.g., 'last_close * 1.005')",
				},
				{
					Name:        "time_in_force",
					Description: "Time in force: GTC|IOC|FOK (default: GTC)",
				},
				{
					Name:        "post_only",
					Description: "Post only: true|false (default: false, maker only)",
				},
				{
					Name:        "quote_ratio",
					Description: "Optional ratio (0-1] of available quote balance to use before leverage",
				},
				{
					Name:        "stop_loss_trigger_price",
					Description: "Stop-loss trigger price",
				},
				{
					Name:        "take_profit_trigger_price",
					Description: "Take-profit trigger price",
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
			Name:        "update_position",
			Description: "Update position stop-loss/take-profit",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "stop_loss_trigger_price",
					Description: "Stop-loss trigger price",
				},
				{
					Name:        "take_profit_trigger_price",
					Description: "Take-profit trigger price",
				},
			},
			Samples: []ttypes.Sample{
				{
					Input: []string{
						"RSI data changed: [2.66 2.65 2.65 2.64 2.64 2.63 2.63 2.63 2.63 2.63 2.63 2.64 2.65 2.66 2.67 2.67 2.68 2.68 2.68 2.68 2.69]",
						"The current position is long, and average cost: 2.80",
						"Analyze data, generate trading cmd",
					},
					Output: []string{
						"Execute cmd: /update_position",
					},
				},
			},
		},
		{
			Name:        "close_position",
			Description: "close position",
			Args: []ttypes.ArgmentDesc{
				{
					Name:        "percentage",
					Description: "Optional close ratio (0-1 or percentage like 50%)",
				},
				{
					Name:        "quantity",
					Description: "Optional base quantity to close",
				},
				{
					Name:        "profit_amount",
					Description: "Optional profit amount in quote currency to realize",
				},
			},
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
				{
					Input: []string{
						"The current position is long, average cost: 2.736, base quantity: 10, accumulated profit value: 120.5",
						"Analyze data, generate trading cmd to lock part of the profit",
					},
					Output: []string{
						"Execute cmd: /close_position percentage=50%",
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

func (ent *ExchangeEntity) getPositionSide(pos *PositionX) types.SideType {
	if pos.IsLong() {
		return types.SideTypeBuy
	} else {
		return types.SideTypeSell
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

				closePercentage := fixedpoint.One

				if percentageArg, ok := args["percentage"]; ok && strings.TrimSpace(percentageArg) != "" {
					val, err := fixedpoint.NewFromString(strings.TrimSuffix(strings.TrimSpace(percentageArg), "%"))
					if err != nil {
						return errors.Wrap(err, "invalid close percentage")
					}

					if strings.Contains(percentageArg, "%") || val.Compare(fixedpoint.One) > 0 {
						val = val.Div(fixedpoint.NewFromInt(100))
					}

					if val.Compare(fixedpoint.Zero) <= 0 {
						return errors.New("close percentage must be greater than zero")
					}

					closePercentage = val
				} else if quantityArg, ok := args["quantity"]; ok && strings.TrimSpace(quantityArg) != "" {
					if ent.position == nil {
						return errors.New("no position data available")
					}

					targetQty, err := fixedpoint.NewFromString(strings.TrimSpace(quantityArg))
					if err != nil {
						return errors.Wrap(err, "invalid close quantity")
					}

					baseQty := ent.position.GetBase().Abs()
					if baseQty.IsZero() {
						return errors.New("no base quantity available for partial close")
					}

					if targetQty.Compare(fixedpoint.Zero) <= 0 {
						return errors.New("close quantity must be greater than zero")
					}

					if targetQty.Compare(baseQty) >= 0 {
						closePercentage = fixedpoint.One
					} else {
						closePercentage = targetQty.Div(baseQty)
					}
				} else if profitArg, ok := args["profit_amount"]; ok && strings.TrimSpace(profitArg) != "" {
					if ent.position == nil {
						return errors.New("no position data available")
					}

					targetProfit, err := fixedpoint.NewFromString(strings.TrimSpace(profitArg))
					if err != nil {
						return errors.Wrap(err, "invalid profit amount")
					}

					if targetProfit.Compare(fixedpoint.Zero) <= 0 {
						return errors.New("profit amount must be greater than zero")
					}

					currentProfit := ent.position.AccumulatedProfitValue.Abs()
					if currentProfit.IsZero() {
						return errors.New("current profit amount unavailable for partial close")
					}

					if targetProfit.Compare(currentProfit) >= 0 {
						closePercentage = fixedpoint.One
					} else {
						closePercentage = targetProfit.Div(currentProfit)
					}
				}

				if closePercentage.Compare(fixedpoint.One) > 0 {
					closePercentage = fixedpoint.One
				}

				log.WithField("percentage", closePercentage.Float64()).
					WithField("modeArgs", args).
					Info("executing close_position command")

				err := ent.ClosePosition(ctx, closePercentage, closePrice)
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
	if cmd == "open_long_position" || cmd == "open_short_position" || cmd == "update_position" {
		side := ent.cmdToSide(cmd)

		var quoteRatio *fixedpoint.Value
		if cmd == "open_long_position" || cmd == "open_short_position" {
			if ratioArg, ok := args["quote_ratio"]; ok {
				raw := strings.TrimSpace(ratioArg)
				if raw != "" {
					isPercent := strings.HasSuffix(raw, "%")
					if isPercent {
						raw = strings.TrimSuffix(raw, "%")
					}

					val, err := fixedpoint.NewFromString(strings.TrimSpace(raw))
					if err != nil {
						return errors.Wrap(err, "invalid quote_ratio")
					}

					if isPercent {
						val = val.Div(fixedpoint.NewFromInt(100))
					}

					if val.Compare(fixedpoint.Zero) <= 0 {
						return errors.New("quote_ratio must be greater than zero")
					}

					if val.Compare(fixedpoint.One) > 0 {
						return errors.New("quote_ratio must be less than or equal to one")
					}

					ratio := val
					quoteRatio = &ratio
				}
			}
		}

		// Close opposite position if any
		if !ent.position.IsDust(closePrice) {
			if (side == types.SideTypeSell && ent.position.IsLong()) || (side == types.SideTypeBuy && ent.position.IsShort()) {
				log.Infof("close existing %s position before open a new position", ent.symbol)
				err := ent.ClosePosition(ctx, fixedpoint.One, closePrice)
				if err != nil {
					return errors.Wrap(err, "close existing position error")
				}
			} else {
				if (cmd == "open_long_position" || cmd == "open_short_position") && quoteRatio == nil {
					return errors.Errorf("existing %s position has the same direction with the signal", ent.symbol)
				}
			}
		}

		opts := make([]interface{}, 0)

		// config stop losss
		if stopLoss, ok := args["stop_loss_trigger_price"]; ok && stopLoss != "" {
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
		if takeProfix, ok := args["take_profit_trigger_price"]; ok && takeProfix != "" {
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

		// config order type
		if orderType, ok := args["order_type"]; ok && orderType != "" {
			opts = append(opts, &OrderTypeOpt{
				Type: types.OrderType(strings.ToUpper(orderType)),
			})
		}

		// config limit price
		if limitPrice, ok := args["limit_price"]; ok && limitPrice != "" {
			price, err := utils.ParsePrice(ent.vm, ent.KLineWindow, closePrice, limitPrice)
			if err != nil {
				return errors.Wrapf(err, "invalid limit_price: %s", limitPrice)
			}

			if price != nil {
				opts = append(opts, &LimitPriceOpt{
					Value: *price,
				})
			}
		}

		// config time in force
		if timeInForce, ok := args["time_in_force"]; ok && timeInForce != "" {
			opts = append(opts, &TimeInForceOpt{
				Value: types.TimeInForce(strings.ToUpper(timeInForce)),
			})
		}

		// config post only
		if postOnly, ok := args["post_only"]; ok && postOnly != "" {
			opts = append(opts, &PostOnlyOpt{
				Enabled: strings.EqualFold(postOnly, "true"),
			})
		}

		if quoteRatio != nil {
			opts = append(opts, &QuoteRatioOpt{
				Value: *quoteRatio,
			})

			log.
				WithField("ratio", quoteRatio.Float64()).
				WithField("symbol", ent.symbol).
				Debug("apply quote_ratio sizing")
		}

		// Validation: order_type=limit requires limit_price
		if ot, ok := args["order_type"]; ok && strings.ToUpper(ot) == "LIMIT" {
			if lp, ok := args["limit_price"]; !ok || lp == "" {
				return errors.New("limit_price is required when order_type=limit")
			}
		}

		log.Infof("open %s position for signal %v, options: %v", ent.symbol, side, opts)

		if cmd == "open_long_position" || cmd == "open_short_position" {
			err := ent.OpenPosition(ctx, side, closePrice, opts...)
			if err != nil {
				return errors.Wrap(err, "open position error")
			}
		} else if cmd == "update_position" {
			side := ent.getPositionSide(ent.position)
			err := ent.UpdatePositionV2(ctx, side, closePrice, opts...)
			if err != nil {
				return errors.Wrap(err, "open position error")
			}
		}

		return nil
	}

	log.Info("no signal")

	return nil
}

func (ent *ExchangeEntity) Run(ctx context.Context, ch chan ttypes.IEvent) {
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

			if ent.KLineWindow.Len() > ent.cfg.KlineNum {
				ent.KLineWindow.Truncate(ent.cfg.KlineNum)
			}
		}

		// Update position accumulated profit metrics
		if ent.position != nil {
			log.WithField("position", ent.position).Info("update_position")

			accumulatedProfit := kline.GetClose().Sub(ent.position.AverageCost).Div(ent.position.AverageCost).Mul(fixedpoint.NewFromFloat(100.0)).Mul(ent.leverage)
			if ent.position.IsShort() {
				accumulatedProfit = accumulatedProfit.Mul(fixedpoint.NewFromInt(-1))
			}

			baseQty := ent.position.GetBase().Abs()
			profitValue := fixedpoint.Zero

			if !baseQty.IsZero() {
				if ent.position.IsLong() {
					profitValue = kline.GetClose().Sub(ent.position.AverageCost).Mul(baseQty)
				} else if ent.position.IsShort() {
					profitValue = ent.position.AverageCost.Sub(kline.GetClose()).Mul(baseQty)
				}

				profitValue = profitValue.Mul(ent.leverage)
			}

			ent.position.UpdateProfit(accumulatedProfit, profitValue)
			ent.position.Dust = ent.position.IsDust(kline.GetClose())
		}

		ent.updatePositionFundRatios(ctx, kline.GetClose())

		log.WithField("kline", kline).Info("kline closed")

		// Auto cleanup unfilled limit orders before new decision cycle
		ent.cleanupLimitOrders(ctx)

		ent.emitEvent(ch, ttypes.NewEvent("kline_changed", ent.KLineWindow))

		for _, indicator := range ent.Indicators {
			ent.emitEvent(ch, ttypes.NewEvent("indicator_changed", indicator))
		}

		ent.emitEvent(ch, ttypes.NewEvent("position_changed", ent.position))

		ent.emitEvent(ch, ttypes.NewEvent("update_finish", nil))
	}))

	// Handle position update
	ent.orderExecutor.TradeCollector().OnPositionUpdate(func(position *types.Position) {
		log.WithField("position", position).Info("ExchangeEntity_OnPositionUpdate")

		if position.IsClosed() {
			log.WithField("position", position).Info("ExchangeEntity_PositionClose")

			// Get the latest close price from KLineWindow
			var exitPrice float64
			if ent.KLineWindow != nil && ent.KLineWindow.Len() > 0 {
				lastIdx := ent.KLineWindow.Len() - 1
				exitPrice = (*ent.KLineWindow)[lastIdx].Close.Float64()
			} else {
				exitPrice = position.AverageCost.Float64() // Fallback if no kline data
			}

			// Determine position data for closed position event
			positionData := PositionClosedEventData{
				StrategyID:           position.StrategyInstanceID,
				Symbol:               ent.symbol,
				EntryPrice:           position.AverageCost.Float64(),
				ExitPrice:            exitPrice,
				Quantity:             position.Base.Float64(),
				ProfitAndLoss:        ent.position.AccumulatedProfitValue.Float64(),
				ProfitAndLossPercent: ent.position.AccumulatedProfit.Float64(),
				CloseReason:          CloseReasonManual, // Default to Manual (will be overridden by the context in ClosePosition if available)
				Timestamp:            time.Now(),
			}

			// Get recent market data as context if available
			if ent.KLineWindow != nil && ent.KLineWindow.Len() > 0 {
				lastIdx := ent.KLineWindow.Len() - 1
				kline := (*ent.KLineWindow)[lastIdx]
				positionData.RelatedMarketData = map[string]interface{}{
					"lastKline": map[string]interface{}{
						"open":      kline.Open.Float64(),
						"high":      kline.High.Float64(),
						"low":       kline.Low.Float64(),
						"close":     kline.Close.Float64(),
						"volume":    kline.Volume.Float64(),
						"startTime": kline.StartTime.Time(),
						"endTime":   kline.EndTime.Time(),
					},
				}
			}

			ent.updatePositionFundRatios(ctx, fixedpoint.NewFromFloat(exitPrice))

			// Emit the position closed event
			go func() {
				log.WithField("positionData", positionData).Info("Emitting position_closed event")
				ent.emitEvent(ch, NewPositionClosedEvent(positionData))
			}()

			if ent.cfg.HandlePositionClose {
				go func() {
					time.Sleep(time.Second * 5)
					log.WithField("position", position).Info("ExchangeEntity_Handle_PositionClose")

					refPrice := fixedpoint.Zero
					if ent.KLineWindow != nil && ent.KLineWindow.Len() > 0 {
						refPrice = ent.KLineWindow.GetClose()
					} else {
						refPrice = fixedpoint.NewFromFloat(exitPrice)
					}

					ent.updatePositionFundRatios(ctx, refPrice)

					ent.emitEvent(ch, ttypes.NewEvent("kline_changed", ent.KLineWindow))

					for _, indicator := range ent.Indicators {
						ent.emitEvent(ch, ttypes.NewEvent("indicator_changed", indicator))
					}

					ent.emitEvent(ch, ttypes.NewEvent("position_changed", ent.position))
					ent.emitEvent(ch, ttypes.NewEvent("update_finish", nil))
				}()
			}
		}
	})

	cleanPostionCfg := ent.cfg.CleanPosition
	if cleanPostionCfg.Enabled {
		log.WithField("config", cleanPostionCfg).Info("clean position enabled")

		session.MarketDataStream.OnKLineClosed(types.KLineWith(ent.symbol, cleanPostionCfg.Interval, func(kline types.KLine) {
			log.WithField("kline", kline).Info("clean position triggered")
			ent.handleCleanPosition(ctx, kline)
		}))
	}
}

func (ent *ExchangeEntity) handleCleanPosition(ctx context.Context, kline types.KLine) {
	exchange := ent.session.Exchange
	service, implemented := exchange.(types.ExchangePositionUpdateService)
	if implemented {
		log.Info("handleCleanPosition_start")

		if ent.position.IsClosed() {
			log.Info("handleCleanPosition_skip_for_no_postion")
			return
		}

		var err error
		var posInfo *types.PositionInfo

		for i := 0; i < 3; i++ {
			duration := time.Duration(rand.Intn(10000)) * time.Millisecond
			log.WithField("duration", duration).Info("handleCleanPosition_delay")
			time.Sleep(duration)

			newCtx, cancel := context.WithTimeout(ctx, time.Second*20)
			defer cancel()

			posInfo, err = service.QueryPositionInfo(newCtx, kline.Symbol)
			if err == nil {
				break
			}

			log.WithField("kline", kline).
				WithField("postion", ent.position).
				WithField("queryPositionInfo", posInfo).
				WithError(err).
				Infof("handleCleanPosition_QueryPositionInfo_fail_retrying attempt %d", i+1)
		}

		if err != nil {
			log.WithField("kline", kline).
				WithField("postion", ent.position).
				WithField("queryPositionInfo", posInfo).
				WithError(err).
				Error("handleCleanPosition_QueryPositionInfo_fail")
			return
		}

		log.WithField("kline", kline).
			WithField("postion", ent.position).
			WithField("queryPositionInfo", posInfo).
			Infof("handleCleanPosition_QueryPositionInfo_success")

		// Check for take profit trigger
		if posInfo.TpTriggerPx != nil {
			currentPrice := kline.Close

			// For long positions: close if price >= take profit
			// For short positions: close if price <= take profit
			if (ent.position.IsLong() && currentPrice.Compare(*posInfo.TpTriggerPx) >= 0) ||
				(ent.position.IsShort() && currentPrice.Compare(*posInfo.TpTriggerPx) <= 0) {

				log.WithField("kline", kline).
					WithField("postion", ent.position).
					WithField("tpTriggerPx", *posInfo.TpTriggerPx).
					WithField("currentPrice", currentPrice).
					Infof("handleCleanPosition_take_profit_triggered")

				// Create context with take profit close reason
				tpCtx := context.WithValue(ctx, "closeReason", CloseReasonTakeProfit)

				err := ent.ClosePosition(tpCtx, fixedpoint.One, currentPrice)
				if err != nil {
					log.WithError(err).Error("handleCleanPosition_take_profit_ClosePosition_fail")
					return
				}

				log.Info("handleCleanPosition_take_profit_ClosePosition_success")
				return
			}
		}

		// Check for stop loss trigger
		if posInfo.SlTriggerPx != nil {
			currentPrice := kline.Close

			// For long positions: close if price <= stop loss
			// For short positions: close if price >= stop loss
			if (ent.position.IsLong() && currentPrice.Compare(*posInfo.SlTriggerPx) <= 0) ||
				(ent.position.IsShort() && currentPrice.Compare(*posInfo.SlTriggerPx) >= 0) {

				log.WithField("kline", kline).
					WithField("postion", ent.position).
					WithField("slTriggerPx", *posInfo.SlTriggerPx).
					WithField("currentPrice", currentPrice).
					Infof("handleCleanPosition_stop_loss_triggered")

				// Create context with stop loss close reason
				slCtx := context.WithValue(ctx, "closeReason", CloseReasonStopLoss)

				err := ent.ClosePosition(slCtx, fixedpoint.One, currentPrice)
				if err != nil {
					log.WithError(err).Error("handleCleanPosition_stop_loss_ClosePosition_fail")
					return
				}

				log.Info("handleCleanPosition_stop_loss_ClosePosition_success")
				return
			}
		}

		// If we get here and there's no stop loss configured, this might be a manual position without SL/TP
		if posInfo.SlTriggerPx == nil && posInfo.TpTriggerPx == nil {
			log.WithField("kline", kline).
				WithField("postion", ent.position).
				WithField("queryPositionInfo", posInfo).
				Infof("handleCleanPosition_found_no_stop_loss_or_take_profit")

			// Positions without stop loss or take profit might need other handling
			// For now just logging, but you could add other logic here
		}
	}
}

// setupIndicators initializes indicators
func (ent *ExchangeEntity) setupIndicators() {
	log.Infof("setup indicators")

	// set kline window
	inc := &types.KLineWindow{}
	dataStore, ok := ent.session.MarketDataStore(ent.symbol)
	if ok {
		if klines, ok := dataStore.KLinesOfInterval(ent.interval); ok {
			log.WithField("klines_length", len(*klines)).Warn("MarketDataStore_klines")

			for _, k := range *klines {
				inc.Add(k)
			}
		} else {
			log.Warn("MarketDataStore_klines_not_found")
		}
	} else {
		log.Warn("MarketDataStore_not_found")
	}

	ent.KLineWindow = inc

	// setup indicators
	indicators := ent.session.StandardIndicatorSet(ent.symbol)

	for name, cfg := range ent.cfg.Indicators {
		log.WithField("name", name).WithField("cfg", cfg).Info("setupIndicators")
		ent.Indicators = append(ent.Indicators, NewExchangeIndicator(name, cfg, indicators))
	}

	sort.Slice(ent.Indicators, func(i int, j int) bool {
		return strings.Compare(string(ent.Indicators[i].Type), string(ent.Indicators[j].Type)) < 0
	})
}

func (ent *ExchangeEntity) emitEvent(ch chan ttypes.IEvent, evt ttypes.IEvent) {
	ch <- evt
}

func clampRatio(val fixedpoint.Value) fixedpoint.Value {
	if val.Compare(fixedpoint.One) > 0 {
		return fixedpoint.One
	}

	if val.Compare(fixedpoint.Zero) < 0 {
		return fixedpoint.Zero
	}

	return val
}

func (ent *ExchangeEntity) updatePositionFundRatios(ctx context.Context, referencePrice fixedpoint.Value) {
	if ent.position == nil {
		return
	}

	quoteCurrency := ent.position.QuoteCurrency
	if quoteCurrency == "" {
		ent.position.UpdateFundRatios(fixedpoint.Zero, fixedpoint.Zero)
		return
	}

	calculator := bbgo.NewAccountValueCalculator(ent.session, quoteCurrency)

	netValue, err := calculator.NetValue(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to calculate account net value for ratios")
		ent.position.UpdateFundRatios(fixedpoint.Zero, fixedpoint.Zero)
		return
	}

	if netValue.IsZero() {
		ent.position.UpdateFundRatios(fixedpoint.Zero, fixedpoint.Zero)
		return
	}

	availableQuote, err := calculator.AvailableQuote(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to calculate available quote for ratios")
	}

	remainingRatio := fixedpoint.Zero
	if err == nil {
		remainingRatio = clampRatio(availableQuote.Div(netValue))
	}

	baseQty := ent.position.GetBase().Abs()
	exposure := fixedpoint.Zero
	if !baseQty.IsZero() && !referencePrice.IsZero() {
		exposure = baseQty.Mul(referencePrice)
	}

	positionRatio := fixedpoint.Zero
	if !exposure.IsZero() {
		positionRatio = clampRatio(exposure.Div(netValue))
	}

	ent.position.UpdateFundRatios(remainingRatio, positionRatio)
}

// cleanupLimitOrders clears all unfilled limit orders
// Called automatically at the start of each decision cycle to ensure AI starts with a clean state
func (ent *ExchangeEntity) cleanupLimitOrders(ctx context.Context) {
	orders, err := ent.session.Exchange.QueryOpenOrders(ctx, ent.symbol)
	if err != nil {
		log.WithError(err).Warn("query open orders for cleanup failed")
		return
	}

	// Only cancel limit orders, keep stop-loss/take-profit orders
	limitOrders := make([]types.Order, 0)
	for _, order := range orders {
		if order.Type == types.OrderTypeLimit || order.Type == types.OrderTypeLimitMaker {
			limitOrders = append(limitOrders, order)
		}
	}

	if len(limitOrders) == 0 {
		return // No limit orders to clean up
	}

	err = ent.session.Exchange.CancelOrders(ctx, limitOrders...)
	if err != nil {
		log.WithError(err).
			WithField("order_count", len(limitOrders)).
			Warn("cancel limit orders failed, continuing anyway")
	} else {
		log.WithField("cancelled_count", len(limitOrders)).
			Info("auto cleanup limit orders before new decision cycle")
	}
}

type StopLossPrice struct {
	Value fixedpoint.Value
}

type TakeProfitPrice struct {
	Value fixedpoint.Value
}

type QuoteRatioOpt struct {
	Value fixedpoint.Value
}

type OrderTypeOpt struct {
	Type types.OrderType
}

type LimitPriceOpt struct {
	Value fixedpoint.Value
}

type TimeInForceOpt struct {
	Value types.TimeInForce
}

type PostOnlyOpt struct {
	Enabled bool
}

func (s *ExchangeEntity) OpenPosition(ctx context.Context, side types.SideType, closePrice fixedpoint.Value, args ...interface{}) error {
	var quoteRatio *fixedpoint.Value
	filteredArgs := make([]interface{}, 0, len(args))

	for _, arg := range args {
		switch val := arg.(type) {
		case *QuoteRatioOpt:
			r := val.Value
			quoteRatio = &r
		default:
			filteredArgs = append(filteredArgs, arg)
		}
	}

	quantity := s.calculateQuantity(ctx, closePrice, side, quoteRatio)

	for {
		if quantity.Compare(s.position.Market.MinQuantity) < 0 {
			return fmt.Errorf("%s order quantity %v is too small, less than %v", s.symbol, quantity, s.position.Market.MinQuantity)
		}

		orderForm := s.generateOrderForm(side, quantity, types.SideEffectTypeMarginBuy)

		for _, arg := range filteredArgs {
			switch val := arg.(type) {
			case *StopLossPrice:
				orderForm.StopPrice = val.Value
			case *TakeProfitPrice:
				orderForm.TakePrice = val.Value
			case *OrderTypeOpt:
				orderForm.Type = val.Type
			case *LimitPriceOpt:
				orderForm.Price = val.Value
			case *TimeInForceOpt:
				orderForm.TimeInForce = val.Value
			case *PostOnlyOpt:
				// Note: PostOnly is not directly supported in bbgo's SubmitOrder
				// Some exchanges may support it through TimeInForce=POST_ONLY
				// For now, we accept the parameter but don't apply it
				log.WithField("post_only", val.Enabled).Debug("post_only parameter received but not applied (not supported by bbgo SubmitOrder)")
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

	// Capture position info before closing for reflection event
	posBeforeClose := *s.position
	isFullClose := percentage.Compare(fixedpoint.One) == 0

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
	if isFullClose {
		orderForm.ClosePosition = true // Full close position
	}

	bbgo.Notify("submitting %s %s order to close position by %v, orderForm:%v", s.symbol, side.String(), percentage, orderForm)

	_, err := s.orderExecutor.SubmitOrders(ctx, orderForm)
	if err != nil {
		log.WithError(err).Errorf("can not place %s position close order", s.symbol)
		bbgo.Notify("can not place %s position close order", s.symbol)
		return err
	}

	// Only emit position closed event for full closures
	if isFullClose {
		// Get the strategy ID from context
		strategyID := "unknown"
		if val, exists := ctx.Value("strategyID").(string); exists {
			strategyID = val
		}

		// Determine close reason based on available context
		closeReason := CloseReasonManual // Default to Manual
		if val, exists := ctx.Value("closeReason").(string); exists {
			closeReason = val
		}
		if val, exists := ctx.Value("closeReason").(string); exists {
			closeReason = val
		}

		// Create the position closed event data
		positionData := PositionClosedEventData{
			StrategyID:           strategyID,
			Symbol:               s.symbol,
			EntryPrice:           posBeforeClose.AverageCost.Float64(),
			ExitPrice:            closePrice.Float64(),
			Quantity:             posBeforeClose.GetBase().Float64(),
			ProfitAndLoss:        posBeforeClose.AccumulatedProfitValue.Float64(),
			ProfitAndLossPercent: posBeforeClose.AccumulatedProfit.Float64(),
			CloseReason:          closeReason,
			Timestamp:            time.Now(),
		}

		// Get recent market data as context if available
		if s.KLineWindow != nil && s.KLineWindow.Len() > 0 {
			lastIdx := s.KLineWindow.Len() - 1
			kline := (*s.KLineWindow)[lastIdx]
			positionData.RelatedMarketData = map[string]interface{}{
				"lastKline": map[string]interface{}{
					"open":      kline.Open.Float64(),
					"high":      kline.High.Float64(),
					"low":       kline.Low.Float64(),
					"close":     kline.Close.Float64(),
					"volume":    kline.Volume.Float64(),
					"startTime": kline.StartTime.Time(),
					"endTime":   kline.EndTime.Time(),
				},
			}
		}

		// Emit the position closed event if we can access a channel
		log.WithField("positionData", positionData).Info("Position fully closed, emitting PositionClosedEvent")

		// Extract event channel from context if available
		if eventCh, exists := ctx.Value("eventChannel").(chan ttypes.IEvent); exists {
			eventCh <- NewPositionClosedEvent(positionData)
		}
	}

	return err
}

func (s *ExchangeEntity) UpdatePosition(ctx context.Context, side types.SideType, closePrice fixedpoint.Value, args ...interface{}) error {
	err := s.ClosePosition(ctx, fixedpoint.NewFromFloat(1.0), closePrice)
	if err != nil {
		return errors.Wrap(err, "UpdatePosition_ClosePosition_error")
	}

	// Create a context with a 20-second timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Create a ticker that ticks every second
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Loop until the position is closed or the context times out
	for {
		select {
		case <-timeoutCtx.Done():
			// Context has reached its deadline
			return errors.Wrap(timeoutCtx.Err(), "UpdatePosition_timeout")
		case <-ticker.C:
			if s.position.IsClosed() {
				// If the position is closed, break out of the loop
				goto POSITION_CLOSED
			}
			// Otherwise, continue looping
		}
	}

POSITION_CLOSED:
	// Once the position is confirmed closed, open a new position
	err = s.OpenPosition(ctx, side, closePrice, args...)
	if err != nil {
		return errors.Wrap(err, "UpdatePosition_OpenPosition_error")
	}

	return nil
}

func (s *ExchangeEntity) UpdatePositionV2(ctx context.Context, side types.SideType, closePrice fixedpoint.Value, args ...interface{}) error {
	exchange := s.session.Exchange
	service, implemented := exchange.(types.ExchangePositionUpdateService)
	if implemented {
		log.Info("UpdatePositionV2_start")

		tmpPos := s.position.Position

		for _, arg := range args {
			switch val := arg.(type) {
			case *StopLossPrice:
				tmpPos.SlTriggerPx = &val.Value
			case *TakeProfitPrice:
				tmpPos.TpTriggerPx = &val.Value
			}
		}

		err := service.UpdatePosition(ctx, tmpPos)
		if err != nil {
			log.WithError(err).Error("UpdatePositionV2_fail")
			return errors.Wrap(err, "UpdatePositionV2_UpdatePosition_error")
		}

		log.Info("UpdatePositionV2_ok")
		return nil
	} else {
		log.Info("Exchange not impl types.ExchangePositionUpdateService")
		return s.UpdatePosition(ctx, side, closePrice, args...)
	}
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
func (s *ExchangeEntity) calculateQuantity(ctx context.Context, currentPrice fixedpoint.Value, side types.SideType, quoteRatio *fixedpoint.Value) fixedpoint.Value {
	quoteQty, err := bbgo.CalculateQuoteQuantity(ctx, s.session, s.position.Market.QuoteCurrency, s.leverage)
	if err != nil {
		log.WithError(err).Errorf("can not update %s quote balance from exchange", s.symbol)
		return fixedpoint.Zero
	}

	if quoteRatio != nil && quoteRatio.Compare(fixedpoint.Zero) > 0 {
		ratio := *quoteRatio
		if ratio.Compare(fixedpoint.One) > 0 {
			log.WithField("ratio", ratio.Float64()).Warn("quote_ratio greater than one detected; clamping to 1")
			ratio = fixedpoint.One
		}

		quoteQty = quoteQty.Mul(ratio)
	}

	if side == types.SideTypeSell {
		return quoteQty.Div(currentPrice).
			Mul(fixedpoint.NewFromFloat(0.99))
	} else {
		return quoteQty
	}
}
