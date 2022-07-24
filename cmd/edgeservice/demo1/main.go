package main

import (
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/edgeservice"
)

func main() {
	edgeservice.Run(demo.DemoEdgeService1ID, demo.DemoEdgeService1Port)
}
