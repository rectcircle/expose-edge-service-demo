package exposer

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/rectcircle/expose-edge-service-demo/demo"
	"github.com/rectcircle/expose-edge-service-demo/helper"
)

type ExposerServer struct {
	upgrader         websocket.Upgrader
	globalRouteTable *redis.Client // exposer-route-table:<service-id>:<device-id> => expose server ip:port
	myIP             string
	myPort           int
	mySessionTable   sync.Map // exposer-route-table:<service-id>:<device-id> => *yamux.Session (client)
}

func NewExposerServer(port int) (*ExposerServer, error) {
	myIP, err := helper.GetIP()
	if err != nil {
		return nil, err
	}
	return &ExposerServer{
		upgrader:         websocket.Upgrader{},
		globalRouteTable: redis.NewClient(&redis.Options{Addr: demo.DemoRedisAddr}),
		myIP:             myIP,
		myPort:           port,
		mySessionTable:   sync.Map{},
	}, nil
}

func (s *ExposerServer) Run() {
	go s.keepalive()
	http.HandleFunc("/", s.serve)
	log.Printf("[exposer server] listening on :%d", s.myPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.myPort), nil); err != nil {
		panic(err)
	}
}

func (s *ExposerServer) serve(w http.ResponseWriter, r *http.Request) {
	edgeFlowType := EdgeFlowType(r.Header.Get(EdgeFlowTypeHeaderKey))
	edgeDeviceID := r.Header.Get(EdgeDeviceIDHeaderKey)
	edgeServiceID := r.Header.Get(EdgeServiceIDHeaderKey)
	if edgeFlowType == EdgeFlowTypeExpose {
		s.expose(w, r, edgeDeviceID, edgeServiceID)
	} else if edgeFlowType == EdgeFlowTypeAccess {
		s.access(w, r, edgeDeviceID, edgeServiceID)
	} else {
		helper.RespString(w, 400, "bad request, not support the flow type: "+string(edgeFlowType))
	}
}

func (s *ExposerServer) expose(w http.ResponseWriter, r *http.Request, edgeDeviceID, edgeServiceID string) {
	log.Printf("[exposer server][device %s, service %s] expose request", edgeDeviceID, edgeServiceID)
	// 应该添加对 device id 和 service id 的校验。本 demo 省略该逻辑。
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] expose websocket upgrade error: %s", edgeDeviceID, edgeServiceID, err.Error())
		helper.RespString(w, 500, "upgrade error: "+err.Error())
		return
	}
	log.Printf("[exposer server][device %s, service %s] expose websocket upgrade success", edgeDeviceID, edgeServiceID)
	// 构建一个 session
	session, err := yamux.Client(&helper.WebsocketConnWrapper{WsConn: wsConn}, nil)
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] make yamux client session error: %s", edgeDeviceID, edgeServiceID, err.Error())
		helper.RespString(w, 500, "internal error: "+err.Error())
		return
	}
	log.Printf("[exposer server][device %s, service %s] make yamux client session success", edgeDeviceID, edgeServiceID)
	// 记录到全局路由表（redis）
	routeKey := helper.RouteKey(edgeServiceID, edgeDeviceID)
	err = s.globalRouteTable.Set(routeKey, s.myIPPort(), 60*time.Second).Err()
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] record route table %s error: %s", edgeDeviceID, edgeServiceID, s.myIPPort(), err.Error())
		helper.RespString(w, 500, "internal error: "+err.Error())
		return
	}
	log.Printf("[exposer server][device %s, service %s] record route table %s success", edgeDeviceID, edgeServiceID, s.myIPPort())
	// 将会话保存到会话表中
	s.mySessionTable.Store(routeKey, session)
	// 等待断开连接
	<-session.CloseChan()
	log.Printf("[exposer server][device %s, service %s] yamux client session has closed, will remove route table and session table", edgeDeviceID, edgeServiceID)
	// 断连后清空路由表
	s.mySessionTable.Delete(routeKey)
	s.globalRouteTable.Del(routeKey)
}

func (s *ExposerServer) myIPPort() string {
	return fmt.Sprintf("%s:%d", s.myIP, s.myPort)
}

func (s *ExposerServer) access(w http.ResponseWriter, r *http.Request, edgeDeviceID, edgeServiceID string) {
	log.Printf("[exposer server][device %s, service %s] access request", edgeDeviceID, edgeServiceID)
	sessionI, ok := s.mySessionTable.Load(helper.RouteKey(edgeServiceID, edgeDeviceID))
	if !ok {
		log.Printf("[exposer server][device %s, service %s] session not found", edgeDeviceID, edgeServiceID)
		helper.RespString(w, 400, "bad request: not found the session")
	}
	log.Printf("[exposer server][device %s, service %s] get session success", edgeDeviceID, edgeServiceID)
	session := sessionI.(*yamux.Session)
	if session.IsClosed() {
		log.Printf("[exposer server][device %s, service %s] session closed", edgeDeviceID, edgeServiceID)
		helper.RespString(w, 400, "bad request: session closed")
		return
	}
	nextConn, err := session.Open()
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] open session error: %s", edgeDeviceID, edgeServiceID, err.Error())
		helper.RespString(w, 500, "open session error: "+err.Error())
	}
	log.Printf("[exposer server][device %s, service %s] open session success", edgeDeviceID, edgeServiceID)
	defer nextConn.Close()
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] access websocket upgrade error: %s", edgeDeviceID, edgeServiceID, err.Error())
		helper.RespString(w, 500, "upgrade error: "+err.Error())
		return
	}
	log.Printf("[exposer server][device %s, service %s] access websocket upgrade success", edgeDeviceID, edgeServiceID)
	wsConnWrapper := &helper.WebsocketConnWrapper{WsConn: wsConn}
	defer wsConnWrapper.Close()
	// _, err = helper. (nextConn, wsConnWrapper)
	err = helper.IORelay(nextConn, wsConnWrapper)
	if err != nil {
		log.Printf("[exposer server][device %s, service %s] access IORelay error: %s", edgeDeviceID, edgeServiceID, err.Error())
	}
	log.Printf("[exposer server][device %s, service %s] access finish", edgeDeviceID, edgeServiceID)
}

func (s *ExposerServer) keepalive() {
	for {
		s.mySessionTable.Range(func(key, value interface{}) bool {
			session := value.(*yamux.Session)
			if session.IsClosed() {
				log.Printf("[exposer server][keepalive] session %s closed, will remove route table and session table", key)
				s.mySessionTable.Delete(key)
				s.globalRouteTable.Del(key.(string))
			} else {
				s.globalRouteTable.Set(key.(string), s.myIPPort(), 60*time.Second)
			}
			return true
		})
		time.Sleep(5 * time.Second)
	}
}
