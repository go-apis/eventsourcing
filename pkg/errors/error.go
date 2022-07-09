package errors

type Code string

const (
	CodeSuccessCompletion Code = "00000"
	CodeInternalError     Code = "XX000"
)

type MyError interface {
	//Error return the message.
	Error() string
	//Cause is the inner error cause.
	Cause() string
	//Stack is present if immudb is running with LEVEL_INFO=debug
	Stack() string
	//Code is the immudb error code
	Code() Code
	//RetryDelay if present the error is retryable after N milliseconds
	RetryDelay() int32
}

func New(message string) *myError {
	return &myError{
		msg:  message,
		code: CodeInternalError,
	}
}

type myError struct {
	cause      string
	code       Code
	msg        string
	retryDelay int32
	stack      string
}

func (f *myError) Error() string {
	return f.msg
}

func (f *myError) Cause() string {
	return f.cause
}

func (f *myError) Stack() string {
	return f.stack
}

func (f *myError) Code() Code {
	return f.code
}

func (f *myError) RetryDelay() int32 {
	return f.retryDelay
}

func (e *myError) WithMessage(message string) *myError {
	e.msg = message
	return e
}

func (e *myError) WithCause(cause string) *myError {
	e.cause = cause
	return e
}

func (e *myError) WithCode(code Code) *myError {
	e.code = code
	return e
}

func (e *myError) WithStack(stack string) *myError {
	e.stack = stack
	return e
}

func (e *myError) WithRetryDelay(retry int32) *myError {
	e.retryDelay = retry
	return e
}

func (e *myError) Is(target error) bool {
	if target == nil {
		return false
	}
	t, ok := target.(MyError)
	if !ok {
		return e.Error() == target.Error()
	}
	return compare(e, t)
}

func compare(e MyError, t MyError) bool {
	if e.Code() != CodeInternalError || t.Code() != CodeInternalError {
		return e.Code() == t.Code()
	}
	return e.Cause() == t.Cause() && e.Error() == t.Error()
}
