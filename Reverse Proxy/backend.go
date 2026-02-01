package main

import (
	"net/url"
	"sync"
	"net/http/httputil"
)

type Backend struct {
	URL 			*url.URL		`json:"url"`
	Alive 			bool			`json:"alive"`
	CurrentConns 	int64			`json:"current_connections"`
	RevProxy 		*httputil.ReverseProxy
	mux 			sync.RWMutex
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}