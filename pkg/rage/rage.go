package rage

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edwinwalela/rage/pkg/config"
	"github.com/enescakir/emoji"
	"github.com/schollz/progressbar/v3"
)

type Rage struct {
	URL         string
	Method      string
	userCount   int
	Attempts    int
	wg          sync.WaitGroup
	progressBar *progressbar.ProgressBar
	client      http.Client
	result      chan Result
	startTime   time.Time
	request     request
}

type request struct {
	contentType string
	headers     map[string]string
	payload     map[string]interface{}
	jsonPayload []byte
}

type Result struct {
	StatusCode    int
	ContentType   string
	ContentLength int64
	Error         error
	RequestTime   time.Duration
}

func (r *Rage) LoadConfig() {
	fmt.Printf("\n%v Initializing rage\n\n", emoji.Fire)
	urlPtr := flag.String("url", "", "Target URL")
	methodPtr := flag.String("method", "", "HTTP request method")
	userCountPtr := flag.Int("users", 1, "Number of users to spawn")
	attemptsPtr := flag.Int("attempts", 1, "Number of requests to make per user")
	filePtr := flag.String("f", "", "Configuration file path")

	flag.Parse()

	if *filePtr != "" {
		cfg, err := config.Parse(*filePtr)
		if err != nil {
			fmt.Printf("failed to parse config file (%s): %v", *filePtr, err)
			os.Exit(2)
		}
		r.loadConfigFile(cfg)
	} else {
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
		r.userCount = *userCountPtr
		r.Attempts = *attemptsPtr
	}
	fmt.Printf("User Count...............: %d\n", r.userCount)
	fmt.Printf("Attempts.................: %d\n", r.Attempts)
	fmt.Printf("Endpoint.................: [%s] %s\n\n\n", r.Method, r.URL)
}

func (r *Rage) Run() {
	r.startTime = time.Now()
	jsonBytes, err := json.Marshal(&r.request.payload)
	if err != nil {
		fmt.Printf("failed to parse request body: %v", err)
		os.Exit(1)
	}
	r.request.jsonPayload = jsonBytes
	r.result = make(chan Result, r.userCount*r.Attempts)
	r.progressBar = progressbar.NewOptions(
		r.userCount*r.Attempts,
		progressbar.OptionSetWidth(30),
	)
	for i := 0; i < r.userCount; i++ {
		r.wg.Add(r.Attempts)
		go func() {
			for k := 0; k < r.Attempts; k++ {
				defer r.wg.Done()
				r.makeRequest()
			}
		}()
	}
	r.wg.Wait()
	close(r.result)
	r.summary()
}

func (r *Rage) loadConfigFile(cfg config.Config) {
	r.URL = cfg.Target.Url
	r.Method = cfg.Target.Method
	r.userCount = cfg.Load.Users
	r.Attempts = cfg.Load.Attempts
	r.request = request{
		headers:     cfg.Headers,
		contentType: cfg.Body["content-type"].(string),
		payload:     cfg.Body["payload"].(map[string]interface{}),
	}
}

func (r *Rage) makeRequest() {
	if r.request.contentType != "" && r.request.contentType != "application/json" {
		fmt.Printf("unsupported request content-type (%s)", r.request.contentType)
		os.Exit(1)
	}

	bodyReader := bytes.NewReader(r.request.jsonPayload)
	req, err := http.NewRequest(r.Method, r.URL, bodyReader)
	req.Header.Add("content-type", r.request.contentType)
	if err != nil {
		return
	}
	startTime := time.Now()
	resp, err := r.client.Do(req)
	requestTime := time.Since(startTime)
	if err != nil {
		r.result <- Result{
			Error:       err,
			RequestTime: requestTime,
		}
		return
	}
	r.result <- Result{
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
		RequestTime:   requestTime,
	}
	r.progressBar.Add(1)
}

func (r *Rage) getExecutionDuration() time.Duration {
	executionDuration := time.Since(r.startTime)

	if executionDuration > time.Microsecond {
		return executionDuration.Truncate(time.Millisecond)
	} else if executionDuration > time.Millisecond {
		return executionDuration.Truncate(time.Second)
	} else {
		return executionDuration.Truncate(time.Minute)
	}
}

func (r *Rage) summary() {
	fmt.Printf("\n\n%v Test complete in %s. Result summary:\n\n", emoji.CheckMarkButton, r.getExecutionDuration())
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

	avgResponseTime := float64(totalResponseTime / int64(r.userCount*r.Attempts))
	successRate := float32(successCount/(r.userCount*r.Attempts)) * 100
	failRate := float32(failCount/r.userCount*r.Attempts) * 100
	fmt.Printf("Success Rate................: %.1f%% (%d/%d)\n", successRate, successCount, r.userCount*r.Attempts)
	fmt.Printf("Failure Rate................: %.1f%% (%d/%d)\n", failRate, failCount, r.userCount*r.Attempts)
	if totalDataReceived > 0 {
		fmt.Printf("Data Received...............: %db\n", totalDataReceived)
	}
	fmt.Printf("Average Latency.............: %.2fms\n", avgResponseTime)
	fmt.Printf("Maximum Latency.............: %s\n", maxResponseTime)
	fmt.Printf("Minimum Latency.............: %s\n\n", minResponseTime)
}
