package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/exposer"
	"github.com/rectcircle/expose-edge-service-demo/helper"
)

// 本例中均为演示，不可以用于生产。

func main() {
	// 配置
	edgeDeviceID := demo.TCPProtoConvDeviceID
	edgeServiceID := demo.TCPProtoConvServiceID
	port := demo.TCPProtoConvPort
	redisAddr := demo.DemoRedisAddr

	// 全局路由表
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	IPPort, err := rdb.Get(helper.RouteKey(edgeServiceID, edgeDeviceID)).Result()
	if err != nil {
		panic(err) // 应该有完善的错误处理
	}
	if IPPort == "" {
		panic("not route found") // 应该有完善的错误处理
	}
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listen.Close()
	log.Printf("[tcp proto conv][device %s, service %s] server listening :%d", edgeDeviceID, edgeServiceID, port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err) // 应该有完善的错误处理
		}
		log.Printf("[tcp proto conv][device %s, service %s] accept success", edgeDeviceID, edgeServiceID)
		go proxy(conn, IPPort, edgeDeviceID, edgeServiceID)
	}
}

func proxy(conn net.Conn, IPPort string, edgeDeviceID, edgeServiceID string) {
	defer conn.Close()
	// 构造 http 路由需要的 header
	header := http.Header{}
	header.Add(exposer.EdgeDeviceIDHeaderKey, edgeDeviceID)
	header.Add(exposer.EdgeServiceIDHeaderKey, edgeServiceID)
	header.Add(exposer.EdgeFlowTypeHeaderKey, string(exposer.EdgeFlowTypeAccess))
	// 打开 websocket 连接
	exposerServerURL := "ws://" + IPPort
	c, _, err := websocket.DefaultDialer.Dial(exposerServerURL, header)
	if err != nil {
		log.Printf("[tcp proto conv][device %s, service %s] proxy connect to %s error: %s", edgeDeviceID, edgeDeviceID, exposerServerURL, err.Error())
		return
	}
	log.Printf("[tcp proto conv][device %s, service %s] proxy connect to %s success", edgeDeviceID, edgeDeviceID, exposerServerURL)
	// 包装成 tcp 连接
	nextConn := &helper.WebsocketConnWrapper{WsConn: c}
	defer nextConn.Close()
	err = helper.IORelay(nextConn, conn)
	if err != nil {
		log.Printf("[tcp proto conv][device %s, service %s] proxy IORelay to %s error: %s", edgeDeviceID, edgeDeviceID, exposerServerURL, err.Error())
		return
	}
	log.Printf("[tcp proto conv][device %s, service %s] proxy to %s finish", edgeDeviceID, edgeDeviceID, exposerServerURL)
}
