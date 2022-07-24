package main

import (
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/exposer"
)

func main() {
	s, err := exposer.NewExposerServer(demo.ExposerServerPort)
	if err != nil {
		panic(err)
	}
	s.Run()
}
