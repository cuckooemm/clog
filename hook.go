package clog

// Hook defines an interface to a log hook.
type Hook interface {
	// Run runs the hook with the event.
	Run(e *Event, level Level, message string)
}

var stp = timestampHook{}

type timestampHook struct{}

func (ts timestampHook) Run(e *Event, _ Level, _ string) {
	e.Timestamp()
}

type TrackingHook struct {
	Key string
	Val string
}

func (t TrackingHook) Run(e *Event, _ Level, _ string) {
	e.Str(t.Key, t.Val)
}
