package pg

import (
	"context"
	"testing"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xgorm"
)

func Test_It(t *testing.T) {
	service := "test"
	reg, err := es.NewRegistry(service)
	if err != nil {
		t.Fatal(err)
		return
	}

	ctx := context.Background()
	cfg := &xgorm.DbConfig{
		Host:         "localhost",
		Port:         5432,
		Username:     "es",
		Password:     "es",
		Database:     "eventstore",
		MaxIdleConns: 10,
		MaxOpenConns: 10,
	}

	db, err := xgorm.NewDb(ctx, cfg)
	if err != nil {
		t.Fatal(err)
		return
	}

	notifications := make(chan *Notification)

	data := newData(service, db, reg, notifications)
	notifier, err := data.NewScheduledCommandNotifier(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}

	// TODO publish something

	one := <-notifier.C
	t.Log(one)
}
