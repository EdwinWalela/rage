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
	StatusCode    int
	ContentType   string
	ContentLength int64
	Error         error
	RequestTime   time.Duration
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

	fmt.Printf("Bot Count.......: %d\n", r.BotCount)
	fmt.Printf("Attempts........: %d\n", r.Attempts)
	fmt.Printf("Endpoint........: [%s] %s\n\n\n", r.Method, r.URL)
}

func (r *Rage) Run() {
	r.result = make(chan Result, 100000)
	r.progressBar = progressbar.Default(int64(r.BotCount * (r.Attempts)))
	for i := 0; i < r.BotCount; i++ {
		r.wg.Add(r.Attempts)
		go func() {
			for k := 0; k < r.Attempts; k++ {
				r.progressBar.Add(1)
				defer r.wg.Done()
				req, err := http.NewRequest(r.Method, r.URL, nil)
				if err != nil {
					continue
				}

				startTime := time.Now()
				resp, err := r.client.Do(req)
				requestTime := time.Since(startTime)
				_ = resp
				_ = requestTime
				if err != nil {
					r.result <- Result{
						Error:       err,
						RequestTime: requestTime,
					}
					continue
				}
				r.result <- Result{
					StatusCode:    resp.StatusCode,
					ContentType:   resp.Header.Get("Content-Type"),
					ContentLength: resp.ContentLength,
					RequestTime:   requestTime,
				}
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
	totalDataReceived := int64(0)
	var maxResponseTime time.Duration
	minResponseTime := time.Hour * 24

	for val := range r.result {
		if val.StatusCode == http.StatusOK {
			successCount++
		} else {
			failCount++
		}

		if val.ContentLength > 0 {
			totalDataReceived += val.ContentLength
		}
		totalResponseTime += val.RequestTime.Milliseconds()
		responseTime := val.RequestTime

		if responseTime > maxResponseTime {
			maxResponseTime = responseTime
		}
		if responseTime < minResponseTime {
			minResponseTime = responseTime
		}
	}

	avgResponseTime := float64(totalResponseTime / int64(r.BotCount))
	successRate := float32(successCount/(r.BotCount*r.Attempts)) * 100
	failRate := float32(failCount/r.BotCount) * 100
	fmt.Printf("Success Rate........: %.1f%% (%d/%d)\n", successRate, successCount, r.BotCount*r.Attempts)
	fmt.Printf("Failure Rate........: %.1f%% (%d/%d)\n", failRate, failCount, r.BotCount*r.Attempts)
	if totalDataReceived > 0 {
		fmt.Printf("Data Received.......: %db\n", totalDataReceived)
	}
	fmt.Printf("Average Latency.....: %.2fms\n", avgResponseTime)
	fmt.Printf("Maximum Latency.....: %s\n", maxResponseTime)
	fmt.Printf("Minimum Latency.....: %s\n\n", minResponseTime)
}
