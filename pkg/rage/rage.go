package rage

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

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
}

func (r *Rage) LoadConfig() {
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
}

func (r *Rage) Run() {
	r.result = make(chan Result, r.BotCount)
	r.progressBar = progressbar.Default(int64(r.BotCount))

	for i := 1; i <= r.BotCount; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			defer r.progressBar.Add(1)
			req, err := http.NewRequest(r.Method, r.URL, nil)
			if err != nil {
				return
			}
			resp, err := r.client.Do(req)
			if err != nil {
				r.result <- Result{
					Error: err,
				}
				return
			}
			r.result <- Result{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
			}
		}()
	}

	r.exit()
	close(r.result)
	// for val := range result {
	// 	fmt.Println(val)
	// }
}

func (r *Rage) exit() {
	r.wg.Wait()
}
