package main

import (
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/edgeservice"
)

func main() {
	edgeservice.Run(demo.DemoEdgeService2ID, demo.DemoEdgeService2Port)
}
