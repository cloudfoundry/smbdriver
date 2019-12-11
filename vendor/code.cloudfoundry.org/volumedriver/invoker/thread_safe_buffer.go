package invoker

import (
	"bytes"
	"sync"
)

type Buffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func (s *Buffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.Write(p)
}

func (s *Buffer) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.String()
}
