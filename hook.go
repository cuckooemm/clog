package clog

// Hook defines an interface to a log hook.
type Hook interface {
	// Run runs the hook with the event.
	Run(e *Event, level Level, message string)
}

var th = timestampHook{}

type timestampHook struct{}

func (ts timestampHook) Run(e *Event, level Level, msg string) {
	e.Timestamp()
}
