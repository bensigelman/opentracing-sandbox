package main

import (
	"fmt"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

type SmartLock struct {
	realLock sync.Mutex

	waiters     []string
	waitersLock sync.Mutex
}

func (sl *SmartLock) Lock(activeSpan opentracing.Span) time.Duration {
	sl.waitersLock.Lock()
	waitersLen := len(sl.waiters)
	if waitersLen > 0 {
		activeSpan.LogEventWithPayload(
			fmt.Sprintf("Waiting for lock behind %d transactions", waitersLen),
			sl.waiters)
	}
	sl.waiters = append(sl.waiters, activeSpan.BaggageItem(donutOriginKey))
	sl.waitersLock.Unlock()

	before := time.Now()
	sl.realLock.Lock()

	sl.waitersLock.Lock()
	behindLen := len(sl.waiters) - 1
	sl.waitersLock.Unlock()
	activeSpan.LogEvent(
		fmt.Sprintf("Acquired lock with %d transactions waiting behind", behindLen))
	return time.Now().Sub(before)
}

func (sl *SmartLock) Unlock() {
	sl.waitersLock.Lock()
	sl.waiters = sl.waiters[0 : len(sl.waiters)-1]
	sl.waitersLock.Unlock()

	sl.realLock.Unlock()
}
