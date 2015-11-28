// Package scsi does scsi stuff
package scsi

import (
	"crypto/rand"
	"fmt"

	"github.com/kurin/tgt/packet"
)

// Session is an iSCSI session.
type Session struct {
	isid uint64
	tsih uint16

	target   *Target
	conn     *Conn
	messages chan *packet.Message
}

func (s *Session) Close() error {
	close(s.messages)
	return s.conn.Close()
}

// New creates a new session.
func NewSession() (*Session, error) {
	var tsih uint16
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	tsih += uint16(b[0]) << 8
	tsih += uint16(b[1])

	return &Session{
		tsih:     tsih,
		messages: make(chan *packet.Message),
	}, nil
}

func (s *Session) Run() error {
	for {
		m, ok := <-s.messages
		if !ok {
			return nil
		}
		if err := s.dispatch(m); err != nil {
			return err
		}
	}
}

func (s *Session) dispatch(m *packet.Message) error {
	switch m.OpCode {
	case packet.OpLoginReq:
		return s.target.handleAuth(s, m)
	case packet.OpSCSICmd:
		return s.target.handleSCSICmd(s, m)
	case packet.OpLogoutReq:
		return s.target.handleLogout(s, m)
	}
	return fmt.Errorf("no handler for op %v", m.OpCode)
}

// Send sends a message to the initiator on the appropriate connection.
func (s *Session) Send(m *packet.Message) error {
	m.ISID = s.isid
	m.TSIH = s.tsih
	return s.conn.Send(m)
}
