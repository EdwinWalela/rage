package main

import "github.com/edwinwalela/rage/pkg/rage"

func main() {
	r := rage.Rage{}
	r.LoadConfig()
	r.Run()
}
