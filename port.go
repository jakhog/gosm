package gosm

import "time"

type Port uint

type portMessage struct {
	Port    Port
	Message interface{}
}

type listenerPort struct {
	From     Port
	To       Port
	Listener *portMultiplexer
}

type internalPort struct {
	Port      Port
	Listeners []*portMultiplexer
}

type portMultiplexer struct {
	In       chan portMessage
	Outgoing []listenerPort
	Internal []*internalPort
}

func portConnector(client, server *portMultiplexer, required, provided Port) {
	// Connect client -> server
	client.Outgoing = append(client.Outgoing, listenerPort{required, provided, server})
	// Also connect server -> client
	server.Outgoing = append(server.Outgoing, listenerPort{provided, required, client})
	// TODO: Is connecting a port to itself a problem?
}

func portInternal(client *portMultiplexer, port Port) {
	client.Internal = append(client.Internal, &internalPort{
		Port:      port,
		Listeners: []*portMultiplexer{client},
	})
}

func (pm *portMultiplexer) send(port Port, message interface{}) {
	// Send this message to all channels that are connected to given port
	for _, lp := range pm.Outgoing {
		if lp.From == port {
			lp.Listener.In <- portMessage{lp.To, message}
		}
	}
	for _, ip := range pm.Internal {
		if ip.Port == port {
			for _, pm := range ip.Listeners {
				pm.In <- portMessage{port, message}
			}
		}
	}
}

func (pm *portMultiplexer) receive(timeout *time.Timer) (Port, interface{}) {
	select {
	case portMsg := <-pm.In:
		return portMsg.Port, portMsg.Message
	case <-timeout.C:
		return 0, nil
	}
}
