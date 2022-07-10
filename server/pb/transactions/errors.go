package transactions

import (
	"fmt"

	"github.com/contextcloud/eventstore/pkg/errors"
)

// var ErrNoSessionIDPresent = errors.New("no sessionID provided").WithCode(errors.CodInvalidAuthorizationSpecification)
// var ErrNoSessionAuthDataProvided = errors.New("no session auth data provided").WithCode(errors.CodInvalidAuthorizationSpecification)

// var ErrOngoingReadWriteTx = errors.New("only 1 read write transaction supported at once").WithCode(errors.CodSqlserverRejectedEstablishmentOfSqlSession)
// var ErrNoTransactionIDPresent = errors.New("no transactionID provided").WithCode(errors.CodInvalidAuthorizationSpecification)
// var ErrNoTransactionAuthDataProvided = errors.New("no transaction auth data provided").WithCode(errors.CodInvalidAuthorizationSpecification)

// var ErrTransactionNotFound = transactions.ErrTransactionNotFound
// var ErrGuardAlreadyRunning = errors.New("session guard already launched")
// var ErrGuardNotRunning = errors.New("session guard not running")

var ErrInvalidOptionsProvided = errors.New("invalid options provided")
var ErrSessionAlreadyPresent = errors.New("session already present").WithCode(errors.CodeInternalError)
var ErrCantCreateSession = errors.New("can not create new session")
var ErrMaxSessionsReached = fmt.Errorf("%w: max sessions number reached", ErrCantCreateSession)
var ErrCantCreateSessionID = fmt.Errorf("%w: generation of session id failed", ErrCantCreateSession)
var ErrSessionNotFound = errors.New("no session found").WithCode(errors.CodeInvalidParameterValue)
