package rage

import (
	"flag"
	"fmt"
	"os"
)

type Rage struct {
	URL      string
	Method   string
	BotCount int
	Attempts int
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
		fmt.Printf("missing required -url flag\n")
		os.Exit(2)
	}

	r.URL = *urlPtr
	r.Method = *methodPtr
	r.BotCount = *botCountPtr
	r.Attempts = *attemptsPtr

}
