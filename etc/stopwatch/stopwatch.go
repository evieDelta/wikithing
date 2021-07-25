package stopwatch

import "time"

// New returns a fresh new stopwatch
func New() *Stopwatch {
	return &Stopwatch{}
}

// NewAndStart creates a new stopwatch and starts it
// mainly exists just to reduce a line of boilerplate
func NewAndStart() *Stopwatch {
	s := New()
	defer s.Start()
	return s
}

// Stopwatch calculates
type Stopwatch struct {
	durTotal  time.Duration
	active    bool
	startTime time.Time
}

// Start starts the counting
func (s *Stopwatch) Start() {
	s.startTime = time.Now()
	s.active = true
}

// Resume unpauses the stopwatch
func (s *Stopwatch) Resume() {
	s.startTime = time.Now()
	s.active = true
}

// Pause pauses the stopwatch and adds the current time to the total
func (s *Stopwatch) Pause() time.Duration {
	if !s.active {
		return 0
	}
	s.durTotal += time.Now().Sub(s.startTime)
	return s.durTotal
}

// Stop stops and resets the stopwatch for future use
func (s *Stopwatch) Stop() time.Duration {
	dr := s.Pause()
	s.Reset()
	return dr
}

// Reset resets the stopwatch for future use
func (s *Stopwatch) Reset() {
	s.active = false
	s.durTotal = 0
	s.startTime = time.Time{}
}
