package gosm

import (
	"time"
)

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
	In                  chan portMessage
	Outgoing            []listenerPort
	Internal            []*internalPort
	PortsBufferLen      int
	ListeningComponents map[*Component]chan portMessage
	PauseChan           chan chan struct{}
}

func newPortMultiplexer(portsBufferLen int, component *Component) *portMultiplexer {
	pm := &portMultiplexer{
		In:                  make(chan portMessage, portsBufferLen),
		Outgoing:            make([]listenerPort, 0),
		Internal:            make([]*internalPort, 0),
		PortsBufferLen:      portsBufferLen,
		ListeningComponents: make(map[*Component]chan portMessage),
		PauseChan:           make(chan chan struct{}),
	}
	pm.ListeningComponents[component] = make(chan portMessage, portsBufferLen)
	go pm.loop()
	return pm
}

func (pm *portMultiplexer) loop() {
	for {
		select {
		// Pause signals stops the processing of incoming messages
		case unpause := <-pm.PauseChan:
			// Wait for the un-pause
			<-unpause
		// Send a copy of all incoming messages to listening components (sessions)
		case portMsg := <-pm.In:
			for _, lc := range pm.ListeningComponents {
				lc <- portMsg
			}
		}
	}
}

func (pm *portMultiplexer) addListeningComponent(original, fork *Component) {
	unpause := make(chan struct{})
	// Wait for the loop to stop forwarding messages (will block)
	pm.PauseChan <- unpause
	// Add a new channel for the listening fork
	pm.ListeningComponents[fork] = make(chan portMessage, pm.PortsBufferLen)
	// Now, make a copy of all messages sent to the original, and send it to both the fork and the original (for later processing)
	originalChan := pm.ListeningComponents[original]
	forkChan := pm.ListeningComponents[fork]
	queued := len(originalChan)
	for i := 0; i < queued; i++ {
		portMsg := <-originalChan
		originalChan <- portMsg
		forkChan <- portMsg
	}
	// Re-start the forwarding
	unpause <- struct{}{}
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

func (pm *portMultiplexer) receive(component *Component, timeout *time.Timer) (Port, interface{}) {
	componentChan := pm.ListeningComponents[component]
	select {
	case portMsg := <-componentChan:
		if !timeout.Stop() {
			<-timeout.C
		}
		return portMsg.Port, portMsg.Message
	case <-timeout.C:
		return 0, nil
	}
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
