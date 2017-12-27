package gosm

//type StateId uint

type State interface {
	OnEntry()
	Handle(Port, interface{}) (bool, bool, State, func())
	OnExit()
}
