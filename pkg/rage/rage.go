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
	result := make(chan Result, r.BotCount)
	client := http.Client{}
	bar := progressbar.Default(int64(r.BotCount))
	for i := 1; i <= r.BotCount; i++ {
		r.Wg.Add(1)
		go func() {
			defer r.Wg.Done()
			defer bar.Add(1)
			req, err := http.NewRequest(r.Method, r.URL, nil)
			if err != nil {
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				result <- Result{
					Error: err,
				}
				return
			}
			result <- Result{
				StatusCode:  resp.StatusCode,
				ContentType: resp.Header.Get("Content-Type"),
			}
		}()
	}

	r.exit()
	close(result)
	// for val := range result {
	// 	fmt.Println(val)
	// }
}

func (r *Rage) exit() {
	r.Wg.Wait()
}
