package eshttp

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type Return struct {
	Id uuid.UUID `json:"id" format:"uuid" required:"true"`
}

func NewUsecaseInteractor[T es.Command]() usecase.Interactor {
	var cmd T
	u := usecase.NewIOI(cmd, new(Return), func(ctx context.Context, input interface{}, output interface{}) error {
		var (
			in  = input.(T)
			out = output.(*Return)
		)

		// do it!
		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}
		tx, err := unit.NewTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		if err := unit.Dispatch(ctx, in); err != nil {
			return err
		}

		if _, err := tx.Commit(ctx); err != nil {
			return err
		}

		out.Id = in.GetAggregateId()
		return nil
	})

	u.SetTitle(fmt.Sprintf("Command %T", cmd))
	u.SetExpectedErrors(status.InvalidArgument)
	return u
}
