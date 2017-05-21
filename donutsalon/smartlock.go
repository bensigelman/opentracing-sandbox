package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

var (
	seededRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type SmartLock struct {
	realLock    sync.Mutex
	queueLength int64
	lockID      string
	cont        bool
	acquired    time.Time
	activeSpan  opentracing.Span
}

func NewSmartLock(cont bool) *SmartLock {
	return &SmartLock{
		realLock: sync.Mutex{},
		lockID:   fmt.Sprintf("smart_lock-%v", seededRNG.Int63()),
		cont:     cont,
	}
}

func (sl *SmartLock) Lock(activeSpan opentracing.Span) {
	sl.activeSpan = activeSpan
	if sl.cont {
		sl.activeSpan.SetTag("c:", sl.lockID)
	}
	atomic.AddInt64(&sl.queueLength, 1)
	sl.realLock.Lock()
	atomic.AddInt64(&sl.queueLength, -1)
	sl.acquired = time.Now()
}

func (sl *SmartLock) Unlock() {
	sl.realLock.Unlock()
	released := time.Now()
	sl.activeSpan.SetTag("weight", int(released.Sub(sl.acquired).Seconds()*1000.0+1))
}

func (sl *SmartLock) QueueLength() float64 {
	return float64(sl.queueLength)
}
