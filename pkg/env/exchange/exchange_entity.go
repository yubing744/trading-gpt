package exchange

import (
	"context"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/indicator"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	ttypes "github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("entity", "exchange")

type ExchangeEntity struct {
	id       string
	symbol   string
	interval types.Interval

	cfg *config.EnvExchangeConfig

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor
	position      *types.Position

	Status types.StrategyStatus
	BOLL   *indicator.BOLL
}

func NewExchangeEntity(
	id string,
	symbol string,
	interval types.Interval,
	cfg *config.EnvExchangeConfig,
	session *bbgo.ExchangeSession,
	orderExecutor *bbgo.GeneralOrderExecutor,
	position *types.Position,
) *ExchangeEntity {
	return &ExchangeEntity{
		id:            id,
		symbol:        symbol,
		interval:      interval,
		cfg:           cfg,
		session:       session,
		orderExecutor: orderExecutor,
		position:      position,
	}
}

func (ent *ExchangeEntity) GetID() string {
	return ent.id
}

func (ent *ExchangeEntity) HandleCommand(ctx context.Context, cmd string, args []string) {
	log.
		WithField("cmd", cmd).
		WithField("args", args).
		Infof("handle command")
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

		log.WithField("kline", kline).Info("kline closed")

		ent.emitEvent(ch, &ttypes.Event{
			Type: "sma_changed",
			Data: ent.BOLL.SMA.Values,
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
