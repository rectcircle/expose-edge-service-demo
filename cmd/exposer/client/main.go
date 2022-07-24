package main

import (
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/exposer"
)

// 每个设备应该有一个唯一的设备 ID，这一块应该在出厂时，固定到设备里。在此使用测试值
const DeviceID = demo.DemoEdgeDeviceID

// 每个设备需要声明，需要暴露的服务的信息，一般从设备的配置文件中读取。目前模拟暴露 demo1 和 demo2 两个服务。
var ExposeServiceInstances = []struct {
	EdgeServiceID        string
	EdgeServiceLocalPort int
}{
	{
		EdgeServiceID:        demo.DemoEdgeService1ID,
		EdgeServiceLocalPort: demo.DemoEdgeService1Port,
	},
	{
		EdgeServiceID:        demo.DemoEdgeService2ID,
		EdgeServiceLocalPort: demo.DemoEdgeService2Port,
	},
}

func main() {
	// 创建一个 exposer 客户端
	c := exposer.NewExposerClient(DeviceID, demo.ExposerServerURL)
	// 将服务暴露到 exposer server 中
	for _, instance := range ExposeServiceInstances {
		c.Expose(instance.EdgeServiceID, instance.EdgeServiceLocalPort)
	}
	// 等待信号
	c.WaitSignal()
}
