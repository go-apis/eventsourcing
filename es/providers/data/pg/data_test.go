package pg

import (
	"context"
	"testing"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/utils/xgorm"
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

	data := newData(service, db, reg)
	notifier, err := data.NewScheduledCommandNotifier(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}

	// TODO publish something

	one := <-notifier.C
	t.Log(one)
}
