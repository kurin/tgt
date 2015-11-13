// Package session provides functions to aggregate iSCSI connections into a
// single iSCSI session.
package session

import (
	"sync"

	"github.com/kurin/tgt/srv"
)

type Pool struct {
	mux      sync.Mutex
	sessions map[uint16]*SCSI
}

func NewPool() *Pool {
	return &Pool{
		sessions: make(map[uint16]*SCSI),
	}
}

type SCSI struct {
	conns []*srv.Conn
}

// Owner takes posession of the iSCSI connection.
type Owner interface {
	// Own passes the iSCSI connection to the owner.  The function blocks while
	// the owner has it, and no other callers should use the connection until Own
	// returns.  If Own returns with no error, then the caller is responsible for
	// the iSCSI connection again.
	Own(*srv.Conn) error
}

// Authenticator is an Owner that retains and exposes certain connection
// variables after authentication is complete.
type Authenticator interface {
	Owner
	TSIH() uint16
	ConnID() uint16
}

// Assign takes a new connection and assigns it to a SCSI session in the
// pool.  If no session exists, a new one is created.
func (p Pool) Assign(conn *srv.Conn, auth Authenticator) error {
	if err := auth.Own(conn); err != nil {
		return err
	}
	// TODO: measure performance and try a readlock here
	p.mux.Lock()
	defer p.mux.Unlock()
	scsi, ok := p.sessions[auth.TSIH()]
	if ok {
		scsi.conns = append(scsi.conns, conn)
		return nil
	}
	p.sessions[auth.TSIH()] = &SCSI{}
	return nil
}
