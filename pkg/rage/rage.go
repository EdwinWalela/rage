package rage

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Rage struct {
	URL      string
	Method   string
	BotCount int
	Attempts int
	Wg       sync.WaitGroup
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
	result := make(map[int]Result)
	client := http.Client{}
	for i := 1; i <= r.BotCount; i++ {
		r.Wg.Add(1)
		go func(i int) {
			defer r.Wg.Done()
			req, err := http.NewRequest(r.Method, r.URL, nil)
			if err != nil {
				result[i] = Result{
					Error: err,
				}
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				result[i] = Result{
					Error: err,
				}
				return
			}
			result[i] = Result{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
			}
		}(i)
	}
	r.exit()
	fmt.Println(result)
}

func (r *Rage) exit() {
	r.Wg.Wait()
}
