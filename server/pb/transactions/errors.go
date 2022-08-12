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

var ErrInvalidOptionsProvided = errors.New("invalid options provided")
var ErrTransactionAlreadyPresent = errors.New("transaction already present").WithCode(errors.CodeInternalError)
var ErrCantCreateTransaction = errors.New("can not create new transaction")
var ErrMaxTransactionsReached = fmt.Errorf("%w: max transactions number reached", ErrCantCreateTransaction)
var ErrTransactionNotFound = errors.New("no transaction found").WithCode(errors.CodeInvalidParameterValue)
var ErrGuardAlreadyRunning = errors.New("transaction guard already launched")
var ErrGuardNotRunning = errors.New("transaction guard not running")
