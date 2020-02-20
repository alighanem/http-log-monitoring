package metric

import (
	"sync"
)

// Simple metric counter
type Counter struct {
	sync.RWMutex
	value int64
}

func NewCounter() *Counter {
	return &Counter{}
}

// Increments the counter with the value
// by default pass 1
func (e *Counter) Inc(value int64) {
	e.Lock()
	defer e.Unlock()

	e.value += value
}

// Value reads and return the counter value
func (e *Counter) Value() int64 {
	e.RLock()
	defer e.RUnlock()

	return e.value
}

// CounterVec is a collection of counters
type CounterVec struct {
	sync.RWMutex
	counters map[string]int64
}

func NewCounterVec() *CounterVec {
	return &CounterVec{
		counters: make(map[string]int64),
	}
}

// Increments the counter value specified by its name
func (e *CounterVec) Inc(label string, value int64) {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.counters[label]; !ok {
		e.counters[label] = value
		return
	}
	e.counters[label] = e.counters[label] + value
}

// Returns the counter value from its name
func (e *CounterVec) Value(label string) int64 {
	e.RLock()
	defer e.RUnlock()

	if _, ok := e.counters[label]; !ok {
		return 0
	}
	return e.counters[label]
}

// Returns all the counters values
func (e *CounterVec) AllValues() map[string]int64 {
	e.RLock()
	defer e.RUnlock()

	values := make(map[string]int64)
	for label, value := range e.counters {
		values[label] = value
	}

	return values
}
