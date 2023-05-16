package rage

import (
	"flag"
)

type Rage struct {
	URL      string
	Method   string
	BotCount uint
	Attempts uint
}

func (r *Rage) LoadConfig() {
	urlPtr := flag.String("url", "", "Target URL")
	methodPtr := flag.String("method", "", "HTTP request method")
	botCountPtr := flag.Uint("bots", 1, "Number of bots to spawn")
	attemptsPtr := flag.Uint("attempts", 1, "Number of requests to make per user")
	flag.Parse()

	r.URL = *urlPtr
	r.Method = *methodPtr
	r.BotCount = *botCountPtr
	r.Attempts = *attemptsPtr
}
