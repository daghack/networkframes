package main

import (
	"demogame"
	"gamenet"
	"math/rand"
	"time"
)

func main() {
	state := demogame.NewState()
	client, err := gamenet.NewGameClient(
		"TestClient",
		"localhost:7777",
		"localhost:7788",
		state,
	)
	if err != nil {
		panic(err)
	}
	err = client.Join()
	if err != nil {
		panic(err)
	}
	go func() {
		panic(client.RunGameClient())
	}()
	tick := time.Tick(50 * time.Millisecond)
	for range tick {
		r := rand.Int63() % 10
		n := rand.Int63() % 2
		if n == 1 {
			r *= -1
		}
		delta := &demogame.Delta{
			Diff: r,
		}
		client.Input(delta)
	}
}
