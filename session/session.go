// Package session provides functions to aggregate iSCSI connections into a
// single iSCSI session.
package session

import (
	"crypto/rand"
	"fmt"

	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
	"github.com/kurin/tgt/target"
)

// Session is an iSCSI session.
type Session struct {
	ISID uint64
	TSIH uint16

	target   *target.Target
	conns    map[uint16]*srv.Conn
	tasks    map[uint32]srv.Handler
	messages chan *packet.Message
}

// New creates a new session.
func New() (*Session, error) {
	var tsih uint16
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	tsih += uint16(b[0]) << 8
	tsih += uint16(b[1])

	return &Session{
		TSIH:     tsih,
		tasks:    make(map[uint32]srv.Handler),
		conns:    make(map[uint16]*srv.Conn),
		messages: make(chan *packet.Message),
	}, nil
}

// AddConn adds a connection to the session.
func (s *Session) AddConn(conn *srv.Conn) {
	s.conns[conn.ID] = conn
	go func() {
		for conn.Recv() {
			s.messages <- conn.Msg()
		}
	}()
}

func (s *Session) Recv() *packet.Message {
	return <-s.messages
}

// Send sends a message to the initiator on the appropriate connection.
func (s *Session) Send(m *packet.Message) error {
	m.ISID = s.ISID
	m.TSIH = s.TSIH
	conn, ok := s.conns[m.ConnID]
	if !ok {
		return fmt.Errorf("session: cannot send message: no such connection %x", m.ConnID)
	}
	return conn.Send(m)
}

// Dispatch receives iSCSI PDUs and passes them to the appropriate task
// handlers.
func (s *Session) Dispatch(m *packet.Message) (*packet.Message, error) {
	h, ok := s.tasks[m.TaskTag]
	if ok {
		return h.Handle(m)
	}

	hf, ok := s.target.TaskMap[m.OpCode]
	if !ok {
		return nil, fmt.Errorf("session: iSCSI option %q unsupported", m.OpCode)
	}

	h = hf()
	s.tasks[m.TaskTag] = h
	return h.Handle(m)
}
