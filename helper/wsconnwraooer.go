package helper

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// https://github.com/gorilla/websocket/issues/282

type WebsocketConnWrapper struct {
	reader io.Reader
	WsConn *websocket.Conn
}

var (
	_ io.ReadWriteCloser = &WebsocketConnWrapper{}
	_ net.Conn           = &WebsocketConnWrapper{}
)

func (w *WebsocketConnWrapper) Write(p []byte) (int, error) {
	// log.Println("[websocket] Write", string(p))
	err := w.WsConn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *WebsocketConnWrapper) Read(p []byte) (int, error) {
	// log.Println("[websocket] Read")
	for {
		if w.reader == nil {
			// Advance to next message.
			var err error
			_, w.reader, err = w.WsConn.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := w.reader.Read(p)
		if err == io.EOF {
			// At end of message.
			w.reader = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}

func (w *WebsocketConnWrapper) Close() error {
	// log.Println("[websocket] Close")
	return w.WsConn.Close()
}

// LocalAddr implements net.Conn
func (w *WebsocketConnWrapper) LocalAddr() net.Addr {
	// log.Println("[websocket] LocalAddr")
	return w.WsConn.UnderlyingConn().LocalAddr()
}

// RemoteAddr implements net.Conn
func (w *WebsocketConnWrapper) RemoteAddr() net.Addr {
	// log.Println("[websocket] RemoteAddr")
	return w.WsConn.UnderlyingConn().RemoteAddr()
}

// SetDeadline implements net.Conn
func (w *WebsocketConnWrapper) SetDeadline(t time.Time) error {
	// log.Println("[websocket] SetDeadline")
	return w.WsConn.UnderlyingConn().SetDeadline(t)
}

// SetReadDeadline implements net.Conn
func (w *WebsocketConnWrapper) SetReadDeadline(t time.Time) error {
	// log.Println("[websocket] SetReadDeadline")
	return w.WsConn.UnderlyingConn().SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn
func (w *WebsocketConnWrapper) SetWriteDeadline(t time.Time) error {
	// log.Println("[websocket] SetWriteDeadline")
	return w.WsConn.UnderlyingConn().SetWriteDeadline(t)
}
