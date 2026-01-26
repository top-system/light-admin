package rpc

import "sync"

// ResponseProcFn is a function that processes a response
type ResponseProcFn func(resp ClientResponse) error

// ResponseProcessor processes responses by ID
type ResponseProcessor struct {
	cbs map[uint64]ResponseProcFn
	mu  *sync.RWMutex
}

// NewResponseProcessor creates a new ResponseProcessor
func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{
		make(map[uint64]ResponseProcFn),
		&sync.RWMutex{},
	}
}

// Add adds a callback for a specific request ID
func (r *ResponseProcessor) Add(id uint64, fn ResponseProcFn) {
	r.mu.Lock()
	r.cbs[id] = fn
	r.mu.Unlock()
}

func (r *ResponseProcessor) remove(id uint64) {
	r.mu.Lock()
	delete(r.cbs, id)
	r.mu.Unlock()
}

// Process processes a response by calling the registered callback
func (r *ResponseProcessor) Process(resp ClientResponse) error {
	id := *resp.Id
	r.mu.RLock()
	fn, ok := r.cbs[id]
	r.mu.RUnlock()
	if ok && fn != nil {
		defer r.remove(id)
		return fn(resp)
	}
	return nil
}
