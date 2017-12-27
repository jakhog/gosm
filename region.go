package gosm

type Region struct {
	//States       []State
	Initial      State
	KeepsHistory bool
	current      State
}

func (r *Region) Enter() {
	if !r.KeepsHistory || r.current == nil {
		r.current = r.Initial
	}
	r.current.OnEntry()
}

func (r *Region) Handle(p Port, m interface{}) bool {
	handled, internal, next, action := r.current.Handle(p, m)
	if handled {
		if internal {
			if action != nil {
				action()
			}
		} else {
			r.current.OnExit()
			if action != nil {
				action()
			}
			next.OnEntry()
			r.current = next
		}
	}
	return handled
}

func (r *Region) Exit() {
	r.current.OnExit()
}
