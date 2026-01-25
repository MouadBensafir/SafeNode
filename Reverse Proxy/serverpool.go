package main

import "sync/atomic"

type ServerPool struct { 
	Backends []*Backend `json:"backends"` 
	Current  atomic.Uint64     
} 