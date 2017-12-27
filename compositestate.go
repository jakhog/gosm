package gosm

type CompositeState struct {
	Regions []Region
}

func (c *CompositeState) OnEntry() {
	for i := range c.Regions {
		c.Regions[i].Enter()
	}
}

func (c *CompositeState) Handle(p Port, m interface{}) (bool, bool, State, func()) {
	handled := false
	for i := range c.Regions {
		handled = c.Regions[i].Handle(p, m) || handled
	}
	// More return values than we need really, but it makes it cleaner to call this
	// function from the actual composite.handle methods
	return handled, true, c, nil
}

func (c *CompositeState) OnExit() {
	for i := range c.Regions {
		c.Regions[i].Exit()
	}
}
