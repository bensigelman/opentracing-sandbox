package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
)

var (
	seededRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type SmartLock struct {
	realLock sync.Mutex

	lockID string
	cont   bool
}

func NewSmartLock(cont bool) *SmartLock {
	return &SmartLock{
		realLock: sync.Mutex{},
		lockID:   fmt.Sprintf("smart_lock-%v", seededRNG.Int63()),
		cont:     cont,
	}
}

func (sl *SmartLock) Lock(activeSpan opentracing.Span) {
	if sl.cont {
		fmt.Println("BHS60", sl.lockID)
		activeSpan.SetTag("c:", sl.lockID)
	}
	sl.realLock.Lock()
}

func (sl *SmartLock) Unlock() {
	sl.realLock.Unlock()
}
