package pg

import (
	"context"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xgorm"
	"github.com/contextcloud/goutils/xlog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func initialize(ctx context.Context, db *gorm.DB, service string, reg es.Registry) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Initialize")
	defer pspan.End()

	if err := db.AutoMigrate(&Event{}, &Snapshot{}, &PersistedCommand{}); err != nil {
		return err
	}

	entities := reg.GetEntities()
	for _, opt := range entities {
		obj, err := opt.Factory()
		if err != nil {
			return err
		}

		table := TableName(service, opt.Name)
		if err := db.Table(table).AutoMigrate(&Entity{}); err != nil {
			return err
		}
		if err := db.Table(table).AutoMigrate(obj); err != nil {
			return err
		}
	}

	return nil
}

type conn struct {
	service       string
	registry      es.Registry
	db            *gorm.DB
	notifications chan *Notification
}

func (c *conn) NewData(ctx context.Context) (es.Data, error) {
	pctx, pspan := otel.Tracer("local").Start(ctx, "NewData")
	defer pspan.End()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case notification := <-c.notifications:
				log := xlog.Logger(ctx)
				log.Debug("notification", zap.Any("notification", notification))
			}
		}
	}()

	db := c.db.WithContext(pctx)
	return newData(c.service, db, c.registry, c.notifications), nil
}

func (c *conn) Close(ctx context.Context) error {
	_, pspan := otel.Tracer("local").Start(ctx, "Close")
	defer pspan.End()

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func NewConn(ctx context.Context, service string, cfg *xgorm.DbConfig, reset bool, registry es.Registry) (es.Conn, error) {
	log := xlog.Logger(ctx)

	notifications := make(chan *Notification)
	afterConnect := stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "LISTEN "+service)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					notification, err := conn.WaitForNotification(ctx)
					if err != nil {
						log.Error("wait-for-notification", zap.Error(err))
						return
					}
					notifications <- &Notification{
						PID:     notification.PID,
						Channel: notification.Channel,
						Payload: notification.Payload,
					}
				}
			}
		}()

		return nil
	})

	dbops := []xgorm.Option{
		xgorm.WithLogger(log.ZapLogger(), gormlogger.Info),
		xgorm.WithTracing(),
		xgorm.WithDisableNestedTransaction(),
		xgorm.WithSkipDefaultTransaction(),
		xgorm.WithOptionOpenDB(afterConnect),
	}
	if reset {
		dbops = append(dbops, xgorm.WithRecreate())
	}

	db, err := xgorm.NewDb(ctx, cfg, dbops...)
	if err != nil {
		return nil, err
	}

	if err := initialize(ctx, db, service, registry); err != nil {
		return nil, err
	}

	return &conn{
		service:       service,
		db:            db,
		registry:      registry,
		notifications: notifications,
	}, nil
}
