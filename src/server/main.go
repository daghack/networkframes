package main

import (
	"demogame"
	"gamenet"
)

func main() {
	state := demogame.NewState()
	g, err := gamenet.NewGameServer(state)
	if err != nil {
		panic(err)
	}
	panic(g.RunGameServer())
}
