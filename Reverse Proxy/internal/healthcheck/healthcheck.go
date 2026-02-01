package healthcheck

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MouadBensafir/SafeNode/internal/pool"
)

func StartHealthChecker(freqMillis int, healthPath string, mainPool *pool.ServerPool) {
	freq := time.Duration(freqMillis) * time.Millisecond
	if freq <= 0 {
		freq = 5 * time.Second
	}

	client := &http.Client{Timeout: 10 * time.Second}
	go func() {
		for {
			var wg sync.WaitGroup

			currentBackends := mainPool.BackendsSnapshot()

			for _, b := range currentBackends {
				wg.Add(1)
				backend := b
				go func() {
					defer wg.Done()
					alive := false
					healthURL := backend.URL.JoinPath(healthPath)

					resp, err := client.Get(healthURL.String())
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
	log.Printf("Health Checker started (interval=%s)", freq)
}
