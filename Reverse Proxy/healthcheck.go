package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func StartHealthChecker(freqMillis int) {
	freq := time.Duration(freqMillis) * time.Millisecond
	if freq <= 0 {
		freq = 5 * time.Second
	}

	client := &http.Client{Timeout: 2 * time.Second}
	go func() {
		for {
			// snapshot of backends
			var wg sync.WaitGroup
			for _, b := range mainPool.Backends {
				wg.Add(1)
				backend := b
				go func() {
					defer wg.Done()
					alive := false
					// Perform a GET request
					resp, err := client.Get(backend.URL.String())
					if err == nil {
						resp.Body.Close()
						if resp.StatusCode >= 200 && resp.StatusCode < 400 {
							alive = true
						}
					}
					mainPool.SetBackendStatus(backend.URL, alive)
				}()
			}
			wg.Wait()
			time.Sleep(freq)
		}
	}()
	log.Printf("health checker started (interval=%s)", freq)
}
