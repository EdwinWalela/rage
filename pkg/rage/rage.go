package rage

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Rage struct {
	URL         string
	Method      string
	BotCount    int
	Attempts    int
	wg          sync.WaitGroup
	progressBar *progressbar.ProgressBar
	client      http.Client
	result      chan Result
}

type Result struct {
	StatusCode  int
	ContentType string
	Error       error
	RequestTime time.Duration
}

func (r *Rage) LoadConfig() {
	fmt.Printf("\nInitializing rage...\n\n")
	urlPtr := flag.String("url", "", "Target URL")
	methodPtr := flag.String("method", "", "HTTP request method")
	botCountPtr := flag.Int("bots", 1, "Number of bots to spawn")
	attemptsPtr := flag.Int("attempts", 1, "Number of requests to make per user")
	flag.Parse()

	if *urlPtr == "" {
		fmt.Printf("missing required -url flag\n")
		os.Exit(2)
	}
	if *methodPtr == "" {
		fmt.Printf("missing required -method flag\n")
		os.Exit(2)
	}

	r.URL = *urlPtr
	r.Method = strings.ToUpper(*methodPtr)
	r.BotCount = *botCountPtr
	r.Attempts = *attemptsPtr

	fmt.Printf("Bot Count -> %d\n", r.BotCount)
	fmt.Printf("Endpoint  -> [%s] %s\n\n\n", r.Method, r.URL)

}

func (r *Rage) Run() {
	r.result = make(chan Result, r.BotCount)
	r.progressBar = progressbar.Default(int64(r.BotCount))

	for i := 1; i <= r.BotCount; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			defer r.progressBar.Add(1)
			startTime := time.Now()
			req, err := http.NewRequest(r.Method, r.URL, nil)
			requestTime := time.Since(startTime)
			if err != nil {
				return
			}
			resp, err := r.client.Do(req)
			if err != nil {
				r.result <- Result{
					Error:       err,
					RequestTime: requestTime,
				}
				return
			}
			r.result <- Result{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
				RequestTime: requestTime,
			}
		}()
	}

	r.wg.Wait()
	close(r.result)
	r.summary()
}

func (r *Rage) summary() {
	fmt.Printf("\n\nTest complete. Result summary:\n\n")
	successCount := 0
	failCount := 0
	totalResponseTime := int64(0)
	var maxResponseTime time.Duration

	for val := range r.result {
		if val.StatusCode == http.StatusOK {
			successCount++
		} else {
			failCount++
		}
		totalResponseTime += val.RequestTime.Microseconds()
		responseTime := val.RequestTime
		if responseTime > maxResponseTime {
			maxResponseTime = responseTime
		}
	}

	avgResponseTime := float64(totalResponseTime / int64(r.BotCount))
	successRate := float32(successCount/r.BotCount) * 100
	failRate := float32(failCount/r.BotCount) * 100

	fmt.Printf("Success Rate    = %.1f%%\n", successRate)
	fmt.Printf("Failure Rate    = %.1f%%\n", failRate)
	fmt.Printf("Average Latency = %.3fµs\n", avgResponseTime)
	fmt.Printf("Maximum Latency = %s\n\n", maxResponseTime)
}
