package error

type ListenerRunning struct{}

func (l ListenerRunning) Error() string {
	return "this listener is already started"
}
