package ws

import (
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type connWithData struct {
	conn *ConnSafe
	data any
}

// concurrency safe
type Room struct {
	conns []connWithData
	mu    *sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		mu: &sync.Mutex{},
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

	writers := make([]io.WriteCloser, len(self.conns))

	for i, v := range self.conns {
		writer, err := v.conn.NextWriter(websocket.TextMessage)
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
			errChan <- write(v, self.conns[i].data)
		}()
	}

	var errors ErrorSlice
	for range writers {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	// don't remove this
	if errors == nil {
		return nil
	}

	return errors
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

	writers := make([]io.WriteCloser, len(self.conns))

	for i, v := range self.conns {
		writer, err := v.conn.NextWriter(websocket.TextMessage)
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

	// don't remove this
	if errors == nil {
		return nil
	}

	return errors
}

// data is optional
func (self *Room) Add(conn *ConnSafe, data any) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.conns = append(self.conns, connWithData{
		data: data,
		conn: conn,
	})
}

func (self *Room) Delete(conn *ConnSafe) {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.conns = slices.DeleteFunc(self.conns, func(v connWithData) bool {
		return v.conn == conn
	})
}
