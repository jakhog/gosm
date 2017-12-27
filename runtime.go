package gosm

import "sync"
import "time"

type Runtime struct {
	components []*Component
	wg         *sync.WaitGroup
}

func (rt *Runtime) startComponent(c *Component) {
	//FIXME: This should probably be done in a thread-safe manner
	rt.components = append(rt.components, c)
	c.Runtime = rt
	rt.wg.Add(1)
	go c.loop()
}

func (rt *Runtime) Activity() bool {
	activity := false
	for _, c := range rt.components {
		activity = activity || c.Active()
	}
	return activity
}

func (rt *Runtime) WaitUntilInactive() {
	for {
		time.Sleep(receiveTimeout)
		if !rt.Activity() {
			break
		}
	}
}

func (rt *Runtime) StopWhenInactive() {
	go func() {
		rt.WaitUntilInactive()
		for _, c := range rt.components {
			c.Stop()
		}
	}()
}

func LaunchComponents(cs ...*Component) *Runtime {
	rt := &Runtime{
		components: make([]*Component, 0),
		wg:         &sync.WaitGroup{},
	}
	for _, c := range cs {
		rt.startComponent(c)
	}
	return rt
}

func RunComponents(cs ...*Component) {
	rt := LaunchComponents(cs...)
	rt.wg.Wait()
}
