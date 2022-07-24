package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/exposer"
	"github.com/rectcircle/expose-edge-service-demo/helper"
)

func main() {
	// 配置项
	redisAddr := demo.DemoRedisAddr
	port := demo.HTTPProtoConvPort

	// 全局路由表
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 读取路由表，获取 exposer 的 ip port
		edgeDeviceID := r.Header.Get(exposer.EdgeDeviceIDHeaderKey)
		edgeServiceID := r.Header.Get(exposer.EdgeServiceIDHeaderKey)
		if edgeDeviceID == "" || edgeServiceID == "" {
			helper.RespString(w, 400, fmt.Sprintf("bad request: %s or %s header not exist", exposer.EdgeDeviceIDHeaderKey, exposer.EdgeServiceIDHeaderKey))
			return
		}
		log.Printf("[http proto conv][device %s, service %s] request", edgeDeviceID, edgeServiceID)
		// 路由信息不需要透传到边缘 service
		r.Header.Del(exposer.EdgeDeviceIDHeaderKey)
		r.Header.Del(exposer.EdgeServiceIDHeaderKey)
		IPPort, err := rdb.Get(helper.RouteKey(edgeServiceID, edgeDeviceID)).Result()
		if err != nil {
			log.Printf("[http proto conv][device %s, service %s] query route table error: %s", edgeDeviceID, edgeServiceID, err.Error())
			helper.RespString(w, 502, "bad gateway: "+err.Error())
			return
		}
		if IPPort == "" {
			log.Printf("[http proto conv][device %s, service %s] route table not found", edgeDeviceID, edgeServiceID)
			helper.RespString(w, 502, "bad gateway: route table not found")
			return
		}
		log.Printf("[http proto conv][device %s, service %s] query route table success: %s", edgeDeviceID, edgeServiceID, IPPort)
		// 使用反向代理库访问 exposer 的 access 服务
		u, _ := url.Parse("http://" + IPPort)
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.Transport = &http.Transport{
			// TCP over websocket
			DialContext: func(_ context.Context, _ string, _ string) (net.Conn, error) {
				// 构造 http 路由需要的 header
				header := http.Header{}
				header.Add(exposer.EdgeDeviceIDHeaderKey, edgeDeviceID)
				header.Add(exposer.EdgeServiceIDHeaderKey, edgeServiceID)
				header.Add(exposer.EdgeFlowTypeHeaderKey, string(exposer.EdgeFlowTypeAccess))
				// 打开 websocket 连接
				exposerServerURL := "ws://" + IPPort
				c, _, err := websocket.DefaultDialer.Dial(exposerServerURL, header)
				if err != nil {
					log.Printf("[http proto conv] connect to %s error: %s", exposerServerURL, err.Error())
					return nil, err
				}
				log.Printf("[http proto conv] connect to %s success", exposerServerURL)
				// 包装成 tcp 连接
				return &helper.WebsocketConnWrapper{WsConn: c}, nil
			},
		}
		proxy.ServeHTTP(w, r)
		log.Printf("[http proto conv][device %s, service %s] finish", edgeDeviceID, edgeServiceID)
	})

	log.Printf("[http proto conv] listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
