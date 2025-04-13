package ws

import (
	"container/list"
	"io"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// concurrency safe
type Room struct {
	conns *list.List
	// extra info for websocket conn
	data map[*websocket.Conn]any // idk maybe extract into generic for type safety, but more complicated
	mu   *sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		mu:    &sync.Mutex{},
		conns: list.New(),
		data:  map[*websocket.Conn]any{},
	}
}

type ErrorSlice []error

func (self ErrorSlice) Error() string {
	msgs := make([]string, len(self))
	for i, v := range self {
		msgs[i] = v.Error()
	}
	return strings.Join(msgs, ", ")
}

// returns ErrorSlice on error
// calls `write` concurrently for each conn with its data,
// errors here should probably be just logged and ignored
func (self *Room) WriteEach(write func(writer io.Writer, data any) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	conns := self.unsafeConns()
	writers := make([]io.WriteCloser, len(conns))

	for i, v := range conns {
		writer, err := v.NextWriter(websocket.TextMessage)
		if err != nil {
			return ErrorSlice{err}
		}
		defer writer.Close()
		writers[i] = writer
	}

	errChan := make(chan error)
	for i, v := range writers {
		go func() {
			// todo: this can take a long time, use a timeout probs
			errChan <- write(v, self.data[conns[i]])
		}()
	}

	var errorSlice ErrorSlice
	for range writers {
		if err := <-errChan; err != nil {
			errorSlice = append(errorSlice, err)
		}
	}

	return errorSlice
}

// returns ErrorSlice on error
// writes to all conns,
// errors here should probably be just logged and ignored
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

	conns := self.unsafeConns()
	writers := make([]io.WriteCloser, len(conns))

	for i, v := range conns {
		writer, err := v.NextWriter(websocket.TextMessage)
		if err != nil {
			return ErrorSlice{err}
		}
		defer writer.Close()
		writers[i] = writer
	}

	var errors ErrorSlice

	buf := make([]byte, 0, 512)
	end := false
	for !end {
		n, err := reader.Read(buf[:cap(buf):cap(buf)])
		buf = buf[:n]

		if err != nil {
			end = true
			if err != io.EOF {
				return ErrorSlice{err}
			}
		}

		errChan := make(chan error)
		for _, v := range writers {
			go func() {
				// todo: this can take a long time, use a timeout probs
				_, err := v.Write(buf)
				errChan <- err
			}()
		}

		for range writers {
			if err := <-errChan; err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// WARNING: does not have mutex, used to call inside a mutex
func (self *Room) unsafeConns() []*websocket.Conn {
	conns := make([]*websocket.Conn, self.conns.Len())

	i := 0
	for v := self.conns.Front(); v != nil; v = v.Next() {
		conns[i] = v.Value.(*websocket.Conn)
		i++
	}

	return conns
}

// data is optional
func (self *Room) Add(conn *websocket.Conn, data any) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.conns.PushBack(conn)
	self.data[conn] = data
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

	delete(self.data, conn)
}
