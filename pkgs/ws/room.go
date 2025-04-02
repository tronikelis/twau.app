package ws

import (
	"container/list"
	"io"
	"sync"

	"github.com/gorilla/websocket"
)

// concurrency safe
type Room struct {
	conns *list.List
	mu    *sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		mu:    &sync.Mutex{},
		conns: list.New(),
	}
}

// writes to all conns
func (self *Room) WriteAll(write func(writer io.Writer) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	reader, writer := io.Pipe()

	go func() {
		if err := write(writer); err != nil {
			writer.CloseWithError(err)
		} else {
			writer.Close()
		}
	}()

	conns := self.connsUnsafe()

	writers := make([]io.WriteCloser, len(conns))

	for i, v := range conns {
		writer, err := v.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer writer.Close()
		writers[i] = writer
	}

	buf := make([]byte, 0, 512)
	end := false
	for !end {
		n, err := reader.Read(buf[:cap(buf):cap(buf)])
		buf = buf[:n]

		if err != nil {
			end = true
			if err != io.EOF {
				return err
			}
		}

		errChan := make(chan error)
		for _, v := range writers {
			go func() {
				_, err := v.Write(buf)
				errChan <- err
			}()
		}

		for range writers {
			// todo: how to handle this error?
			<-errChan
		}
	}

	return nil
}

func (self *Room) connsUnsafe() []*websocket.Conn {
	var conns []*websocket.Conn

	for v := self.conns.Front(); v != nil; v = v.Next() {
		conns = append(conns, v.Value.(*websocket.Conn))
	}

	return conns
}

func (self *Room) Add(conn *websocket.Conn) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.conns.PushBack(conn)
}

func (self *Room) Delete(conn *websocket.Conn) {
	self.mu.Lock()
	defer self.mu.Unlock()

	for v := self.conns.Front(); v != nil; v = v.Next() {
		if v.Value.(*websocket.Conn) == conn {
			self.conns.Remove(v)
			break
		}
	}
}
