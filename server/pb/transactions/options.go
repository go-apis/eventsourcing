package transactions

import (
	"math"
	"time"
)

const infinity = time.Duration(math.MaxInt64)

type Options struct {
	SessionGuardCheckInterval time.Duration
	// MaxSessionInactivityTime is a duration for the amount of time after which an idle session would be closed by the server
	MaxSessionInactivityTime time.Duration
	// MaxSessionAgeTime is a duration for the maximum amount of time a session may exist before it will be closed by the server
	MaxSessionAgeTime time.Duration
	// Timeout the server waits for a duration of Timeout and if no activity is seen even after that the session is closed
	Timeout time.Duration
	// Max number of simultaneous sessions
	MaxSessions int
}

func DefaultOptions() *Options {
	return &Options{
		SessionGuardCheckInterval: time.Minute * 1,
		MaxSessionInactivityTime:  time.Minute * 3,
		MaxSessionAgeTime:         infinity,
		Timeout:                   time.Minute * 2,
		MaxSessions:               100,
	}
}
