package gosm

import "time"

func (original *Component) ForkComponent(session State) *Component {
	component := &Component{
		ports:       original.ports,
		stateCharts: []State{session},
		timer:       time.NewTimer(receiveTimeout),
		stopped:     false,
		active:      true,
	}
	original.ports.addListeningComponent(original, component)
	/*
		// Clone all the ports from the original component
		for lp := range original.ports.Outgoing {
			portConnector(&component.ports, original.ports.Outgoing[lp].Listener, original.ports.Outgoing[lp].From, original.ports.Outgoing[lp].To)
		}
		for _, ip := range original.ports.Internal {
			component.ports.Internal = append(component.ports.Internal, ip)
			ip.Listeners = append(ip.Listeners, &component.ports)
		}
	*/
	return component
}

func (original *Component) LaunchSession(session *Component) {
	original.Runtime.startComponent(session)
}
