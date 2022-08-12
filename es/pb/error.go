package pb

import (
	"github.com/contextcloud/eventstore/es/pb/store"
	"github.com/contextcloud/eventstore/pkg/errors"
	"google.golang.org/grpc/status"
)

func FromError(err error) errors.MyError {
	if err == nil {
		return nil
	}

	if inner, ok := err.(errors.MyError); ok {
		// Already an ImmuError instance
		return inner
	}

	st, ok := status.FromError(err)
	if ok {
		ie := errors.New(st.Message())
		for _, det := range st.Details() {
			switch ele := det.(type) {
			case *store.ErrorInfo:
				ie.WithCode(errors.Code(ele.Code)).WithCause(ele.Cause)
			case *store.DebugInfo:
				ie.WithStack(ele.Stack)
			case *store.RetryInfo:
				ie.WithRetryDelay(ele.RetryDelay)
			}
		}
		return ie
	}
	return errors.New(err.Error())
}
