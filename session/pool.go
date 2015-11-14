package session

import (
	"errors"
	"fmt"

	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
	"github.com/kurin/tgt/target"
)

// Pool is a collection of existing iSCSI sessions and registered iSCSI
// targets.
type Pool struct {
	sessions map[uint16]*Session
	targets  map[string]*target.Target
}

func (p *Pool) Session(tsih uint16) (*Session, error) {
	if tsih == 0 {
		s, err := New()
		if err != nil {
			return nil, err
		}
		if p.sessions == nil {
			p.sessions = make(map[uint16]*Session)
		}
		p.sessions[s.TSIH] = s
		return s, nil
	}
	s, ok := p.sessions[tsih]
	if !ok {
		return nil, fmt.Errorf("session: no such tsih: %d", tsih)
	}
	return s, nil
}

// Accept takes a new connection and assigns it to a SCSI session in the pool.
// If no session exists, a new one is created.
func (p *Pool) Accept(conn *srv.Conn) (*Session, error) {
	if !conn.Recv() {
		err := conn.Err()
		if err != nil {
			return nil, err
		}
		return nil, errors.New("session: no packet from connection, but no error either")
	}
	m := conn.Msg()
	s, err := p.Session(m.TSIH)
	if err != nil {
		return nil, err
	}
	if s.target == nil {
		txt := m.RawData[:m.DataLen]
		params := packet.ParseKVText(txt)
		tname, ok := params["TargetName"]
		if !ok {
			return nil, errors.New("session: TargetName not in first login packet")
		}
		tgt, err := p.GetTarget(tname)
		if err != nil {
			return nil, err
		}
		s.target = tgt
	}
	go func() {
		// put the message back on the queue first
		s.messages <- m
		s.AddConn(conn)
	}()
	return s, nil
}

// RegisterTarget registers an iSCSI target with the pool.
func (p *Pool) RegisterTarget(t *target.Target) error {
	if p.targets == nil {
		p.targets = make(map[string]*target.Target)
	}
	if t.Name == "" {
		return errors.New("session: iSCSI targets require a name")
	}
	if _, ok := p.targets[t.Name]; ok {
		return fmt.Errorf("session: target %s already exists", t.Name)
	}
	p.targets[t.Name] = t
	return nil
}

// GetTarget gets the iSCSI target from the pool.
func (p *Pool) GetTarget(name string) (*target.Target, error) {
	tgt, ok := p.targets[name]
	if !ok {
		return nil, fmt.Errorf("session: no such target: %s", name)
	}
	return tgt, nil
}
