package exchange

import (
	"context"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	ttypes "github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("entity", "exchange")

type ExchangeEntity struct {
	id  string
	cfg *config.EnvExchangeConfig

	session       *bbgo.ExchangeSession
	orderExecutor *bbgo.GeneralOrderExecutor
	position      *types.Position

	Status types.StrategyStatus
}

func NewExchangeEntity(
	id string,
	cfg *config.EnvExchangeConfig,
	session *bbgo.ExchangeSession,
	orderExecutor *bbgo.GeneralOrderExecutor,
	position *types.Position,
) *ExchangeEntity {
	return &ExchangeEntity{
		id:            id,
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

	// if you need to do something when the user data stream is ready
	// note that you only receive order update, trade update, balance update when the user data stream is connect.
	session.UserDataStream.OnStart(func() {
		log.Infof("connected")
	})

	log.WithField("session", session).
		Infof("run")
	/*
		session.MarketDataStream.OnKLineClosed(types.KLineWith(ent.cfg.Symbol, ent.cfg.Interval, func(kline types.KLine) {
			// StrategyController
			if ent.Status != types.StrategyStatusRunning {
				log.Info("strategy status not running")
				return
			}

			log.WithField("position", ent.position).
				Infof("current position")

		}))
	*/
}
