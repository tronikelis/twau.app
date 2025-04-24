package ws

import (
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

type writeCloserSafe struct {
	inner io.WriteCloser
	mu    *sync.Mutex
}

func newWriteCloserSafe(inner io.WriteCloser, mu *sync.Mutex) writeCloserSafe {
	mu.Lock()
	return writeCloserSafe{
		inner: inner,
		mu:    mu,
	}
}

func (self writeCloserSafe) Write(bytes []byte) (int, error) {
	return self.inner.Write(bytes)
}

func (self writeCloserSafe) Close() error {
	defer self.mu.Unlock()
	return self.inner.Close()
}

type ConnSafe struct {
	// Connections support one concurrent reader and one concurrent writer.
	// Applications are responsible for ensuring that no more than one goroutine calls the write methods
	// (NextWriter, SetWriteDeadline, WriteMessage, WriteJSON, EnableWriteCompression, SetCompressionLevel) concurrently and that no more than one goroutine calls the read methods
	// (NextReader, SetReadDeadline, ReadMessage, ReadJSON, SetPongHandler, SetPingHandler) concurrently.
	// The Close and WriteControl methods can be called concurrently with all other methods.
	conn    *websocket.Conn
	readMu  *sync.Mutex
	writeMu *sync.Mutex
}

func NewConnSafe(conn *websocket.Conn) *ConnSafe {
	return &ConnSafe{
		conn:    conn,
		readMu:  &sync.Mutex{},
		writeMu: &sync.Mutex{},
	}
}

func (self *ConnSafe) Close() error {
	return self.conn.Close()
}

func (self *ConnSafe) ReadMessage() (int, []byte, error) {
	self.readMu.Lock()
	defer self.readMu.Unlock()
	return self.conn.ReadMessage()
}

// users MUST call .Close() when they are done, this unlocks the mutex for other goroutines
func (self *ConnSafe) NextWriter(messageType int) (io.WriteCloser, error) {
	inner, err := self.conn.NextWriter(messageType)
	if err != nil {
		return nil, err
	}

	return newWriteCloserSafe(inner, self.writeMu), nil
}
