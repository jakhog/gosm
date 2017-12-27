package gosm

import "time"

const receiveTimeout time.Duration = time.Millisecond * 100

type Component struct {
	ports       portMultiplexer
	stateCharts []State
	timer       *time.Timer
	active      bool
	stopped     bool
	Runtime     *Runtime
}

func (c *Component) loop() {
	// The actual event loop that handles incoming messages
	for sc := range c.stateCharts {
		c.stateCharts[sc].OnEntry()
	}
	for {
		// The activity flag indicates if anything has happened since the last receive timeout
		active := false
		// Process empty events until nothing happens
		handled := true
		for handled {
			handled = false
			for sc := range c.stateCharts {
				scHandled, _, _, action := c.stateCharts[sc].Handle(0, nil)
				if scHandled && action != nil {
					action()
				}
				handled = handled || scHandled
				active = active || scHandled
			}
		}
		// Handle incoming messages
		c.timer.Reset(receiveTimeout)
		p, m := c.ports.receive(c.timer)
		for sc := range c.stateCharts {
			scHandled, _, _, action := c.stateCharts[sc].Handle(p, m)
			if scHandled && action != nil {
				action()
			}
			active = active || scHandled
		}
		c.active = active
		// Check if we have been stopped
		if c.stopped {
			break
		}
	}
	for sc := range c.stateCharts {
		c.stateCharts[sc].OnExit()
	}
	c.active = false
	c.Runtime.wg.Done()
}

func MakeComponent(portsBufferLen int, stateCharts ...State) *Component {
	return &Component{
		ports: portMultiplexer{
			In:       make(chan portMessage, portsBufferLen),
			Outgoing: make([]listenerPort, 0),
			Internal: make([]*internalPort, 0),
		},
		stateCharts: stateCharts,
		timer:       time.NewTimer(receiveTimeout),
		stopped:     false,
		active:      true,
	}
}

func (c *Component) Send(port Port, message interface{}) {
	c.ports.send(port, message)
}

func (c *Component) Stop() {
	c.stopped = true
}

func (c *Component) Active() bool {
	return c.active || len(c.ports.In) > 0
}

func Connector(client, server *Component, required, provided Port) {
	portConnector(&client.ports, &server.ports, required, provided)
}

func InternalPort(client *Component, port Port) {
	portInternal(&client.ports, port)
}
