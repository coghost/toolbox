package sleep

import (
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// Sleeper implements exponential backoff with optional jitter for retry mechanisms.
// It provides configurable base and maximum delay durations, as well as the ability
// to enable/disable random jitter to prevent thundering herd problems in distributed systems.
type Sleeper struct {
	logger    *zap.Logger
	attempts  int
	baseDelay time.Duration
	maxDelay  time.Duration
	useJitter bool
}

// NewSleeper creates a new Sleeper for implementing exponential backoff delays.
// By default, it:
//   - Uses a base delay of 5 seconds
//   - Uses a maximum delay of 30 minutes
//   - Enables random jitter
//   - Uses a no-op logger if none is provided
//
// The returned Sleeper can be further configured using:
//   - WithDelays() to customize the base and max delay durations
//   - WithJitter() to enable/disable random jitter
//
// Example usage:
//
//	sleeper := NewSleeper(logger).WithDelays(1*time.Second, 1*time.Minute)
//	for {
//	    err := doSomething()
//	    if err == nil {
//	        break
//	    }
//	    sleeper.Sleep()
//	}
func NewSleeper(logger *zap.Logger) *Sleeper {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Sleeper{
		logger:    logger,
		baseDelay: 5 * time.Second,
		maxDelay:  30 * time.Minute,
		useJitter: true, // enable by default
	}
}

// WithDelays allows customizing the base and max delays
func (s *Sleeper) WithDelays(base, max time.Duration) *Sleeper {
	s.baseDelay = base
	s.maxDelay = max
	return s
}

// WithJitter enables or disables random jitter in sleep duration.
// Jitter helps prevent thundering herd problems by randomizing actual sleep time.
func (s *Sleeper) WithJitter(enable bool) *Sleeper {
	s.useJitter = enable
	return s
}

// Sleep performs exponential backoff sleep with optional jitter and logging.
// When jitter is enabled, the actual sleep time will be between the calculated
// delay and up to 2x that value.
// Returns actual sleep duration for information purposes.
func (s *Sleeper) Sleep() time.Duration {
	// Calculate base exponential delay
	baseDelay := time.Duration(math.Min(
		float64(s.baseDelay)*math.Pow(2, float64(s.attempts)),
		float64(s.maxDelay),
	))

	actualDelay := baseDelay
	if s.useJitter {
		// Add random jitter between 0% to 100% of calculated delay
		jitter := time.Duration(rand.Float64() * float64(baseDelay))
		actualDelay = baseDelay + jitter

		s.logger.Info("backing off with jitter",
			zap.Duration("base_delay", baseDelay),
			zap.Duration("jittered_delay", actualDelay),
			zap.Int("attempt", s.attempts+1))
	} else {
		s.logger.Info("backing off",
			zap.Duration("delay", actualDelay),
			zap.Int("attempt", s.attempts+1))
	}

	time.Sleep(actualDelay)
	s.attempts++
	return actualDelay
}

// Reset resets the attempt counter to 0
func (s *Sleeper) Reset() {
	s.attempts = 0
}
