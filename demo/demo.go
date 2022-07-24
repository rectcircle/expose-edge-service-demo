package demo

const (
	DemoEdgeService1ID   = "demo1"
	DemoEdgeService1Port = 8081
	DemoEdgeService2ID   = "demo2"
	DemoEdgeService2Port = 8082

	DemoEdgeDeviceID = "DEVICE-0000"

	DemoRedisAddr = "localhost:6379"

	ExposerServerURL  = "ws://localhost:8080"
	ExposerServerPort = 8080

	HTTPProtoConvPort = 9000

	TCPProtoConvPort      = 9001
	TCPProtoConvServiceID = DemoEdgeService2ID
	TCPProtoConvDeviceID  = DemoEdgeDeviceID
)
