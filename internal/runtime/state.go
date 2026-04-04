package runtime

import "time"

type State struct {
	StartedAt           time.Time
	LastSuccessAt       time.Time
	SuccessfulSends     int
	ConsecutiveFailures int
	LastError           string
}

func NewState(now time.Time) *State {
	return &State{
		StartedAt: now,
	}
}

func (s *State) MarkSuccess(now time.Time) {
	s.LastSuccessAt = now
	s.SuccessfulSends++
	s.ConsecutiveFailures = 0
	s.LastError = ""
}

func (s *State) MarkFailure(err error) {
	s.ConsecutiveFailures++
	if err != nil {
		s.LastError = err.Error()
	}
}
