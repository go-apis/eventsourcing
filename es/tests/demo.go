package tests

import (
	"context"
	"eventstore/es"
	"fmt"
)

type CreateTest struct {
	es.BaseCommand

	Name string
}

type TestCreated struct {
	Name string
}

type TestAggregate struct {
	es.BaseAggregate

	Name string
}

func (a TestAggregate) Handle(ctx context.Context, cmd *CreateTest) error {
	a.Apply(ctx, &TestCreated{Name: cmd.Name})
	return fmt.Errorf("Whoops")
}
