package main

import (
	"sync/atomic"
	"sync"
)

type ServerPool struct { 
	Backends []*Backend `json:"backends"` 
	Current  atomic.Uint64     
	mux sync.RWMutex
} 