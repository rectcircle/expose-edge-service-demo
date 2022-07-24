package exposer

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/rectcircle/expose-edge-service-demo/helper"
)

type ExposerClient struct {
	DeviceID  string
	ServerURL string

	wg           sync.WaitGroup
	exposedFlags sync.Map // <service-id> => chan(struct{})
}

func NewExposerClient(deviceID string, serviceURL string) *ExposerClient {
	return &ExposerClient{
		DeviceID:     deviceID,
		ServerURL:    serviceURL,
		wg:           sync.WaitGroup{},
		exposedFlags: sync.Map{},
	}
}

func (c *ExposerClient) WaitSignal() {
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
		sno := <-s
		log.Printf("[exposer client][device %s] receive signal: %s", c.DeviceID, sno.String())
		c.exposedFlags.Range(func(key, _ interface{}) bool {
			c.UnExpose(key.(string))
			return true
		})
	}()
	c.wg.Wait()
}

func (c *ExposerClient) Expose(ServiceID string, ServiceLocalPort int) {
	wantCloseChan := make(chan struct{})
	if _, ok := c.exposedFlags.LoadOrStore(ServiceID, wantCloseChan); ok {
		log.Printf("[exposer client][device %s, service %s] already exposed", c.DeviceID, ServiceID)
		return
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		header := http.Header{}
		header.Add(EdgeDeviceIDHeaderKey, c.DeviceID)
		header.Add(EdgeServiceIDHeaderKey, ServiceID)
		header.Add(EdgeFlowTypeHeaderKey, string(EdgeFlowTypeExpose))

		for tryNumber := 0; ; tryNumber++ {
			if tryNumber != 0 {
				select {
				case <-wantCloseChan:
					log.Printf("[exposer client][device %s, service %s] close expose success", c.DeviceID, ServiceID)
					return
				case <-time.After(time.Second):
					log.Printf("[exposer client][device %s, service %s] retry: %d ...", c.DeviceID, ServiceID, tryNumber)
				}
			}
			// 打开 websocket 连接
			wsConn, _, err := websocket.DefaultDialer.Dial(c.ServerURL, header)
			if err != nil {
				log.Printf("[exposer client][device %s, service %s] connect to exposer server %s error: %s", c.DeviceID, ServiceID, c.ServerURL, err.Error())
				continue // 重试
			}
			log.Printf("[exposer client][device %s, service %s] try connect to exposer server %s success", c.DeviceID, ServiceID, c.ServerURL)
			// 包装成 tcp 连接
			conn := &helper.WebsocketConnWrapper{WsConn: wsConn}
			// 构建 yamux server
			session, err := yamux.Server(conn, nil)
			if err != nil {
				log.Printf("[exposer client][device %s, service %s] make yamux server session error: %s", c.DeviceID, ServiceID, err.Error())
				continue // 重试
			}
			log.Printf("[exposer client][device %s, service %s] make yamux server session success", c.DeviceID, ServiceID)
			// 获取是否需要关闭该 session
			go func() {
				select {
				case <-session.CloseChan(): // 这个链接关闭了
					log.Printf("[exposer client][device %s, service %s] yamux server session has closed,", c.DeviceID, ServiceID)
				case <-wantCloseChan:
					_ = session.Close()
					log.Printf("[exposer client][device %s, service %s] close yamux server session", c.DeviceID, ServiceID)
				}
			}()
			for {
				conn, err1 := session.Accept()
				if err1 != nil {
					log.Printf("[exposer client][device %s, service %s] session accept error: %s", c.DeviceID, ServiceID, err1.Error())
					break
				}
				log.Printf("[exposer client][device %s, service %s] session accept success", c.DeviceID, ServiceID)
				go c.proxy(conn, ServiceLocalPort)
			}
		}
	}()
}

func (c *ExposerClient) UnExpose(ServiceID string) {
	if wantCloseChanI, ok := c.exposedFlags.Load(ServiceID); ok {
		log.Printf("[exposer client][device %s, service %s] want to close expose", c.DeviceID, ServiceID)
		c.exposedFlags.Delete(ServiceID)
		wantCloseChan := wantCloseChanI.(chan struct{})
		close(wantCloseChan)
		return
	}
}

func (c *ExposerClient) Wait() {
	c.wg.Wait()
}

func (c *ExposerClient) proxy(conn net.Conn, port int) {
	defer conn.Close()
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Printf("[exposer client][proxy] parse localhost:%d error: %s", port, err.Error())
		return
	}
	log.Printf("[exposer client][proxy] parse localhost:%d success", port)
	nextConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("[exposer client][proxy] open tcp connect to localhost:%d error: %s", port, err.Error())
		return
	}
	log.Printf("[exposer client][proxy] open tcp connect to localhost:%d success", port)
	defer nextConn.Close()
	err = helper.IORelay(nextConn, conn)
	if err != nil {
		log.Printf("[exposer client][proxy] ip copy error: %s", err.Error())
		return
	}
	log.Printf("[exposer client][proxy] proxy finish")
}
