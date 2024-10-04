/*
A single flight module that prevents redundant or duplicate calls.
It provides the guarantee that even when multiple threads request
the same key, only one of those goroutines actually goes to fetch the data.
*/
package singleflight

import "sync"

// A task that is current ongoing
// or is finished
type promise struct {
	// WaitGroup provides synchronization mechanism
	wg sync.WaitGroup
	// interface {} is the `any` type
	val interface{}
	// if the task results in an error
	err error
}

type Group struct {
	mu             sync.Mutex
	keyPromiseDict map[string]*promise
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()

	if g.keyPromiseDict == nil {
		g.keyPromiseDict = make(map[string]*promise)
	}

	if p, ok := g.keyPromiseDict[key]; ok {
		g.mu.Unlock()
		// if there is ready a task for this key, wait until it finishes
		// this is the main idea for the singleflight, very simple
		// However, we havn't consider to limit the overall request number
		// external data source. Therefore, we still cannot guarantee that
		// there is no huge amount of requests being sent in a short period
		// of time.
		// TODO: apply overall limitation
		p.wg.Wait()
		// then we returns whatever has been fetched by that goroutine
		return p.val, p.err
	}

	p := new(promise)
	// start block all same tasks
	p.wg.Add(1)
	g.keyPromiseDict[key] = p
	// allow other tasks for other keys to begin
	g.mu.Unlock()
	// this is a very useful shortcut
	p.val, p.err = fn()

	// eq to wg.Add(-1)
	p.wg.Done()

	g.mu.Lock()
	// we can do this because other tasks already get the ref to p
	delete(g.keyPromiseDict, key)
	g.mu.Unlock()

	return p.val, p.err
}
