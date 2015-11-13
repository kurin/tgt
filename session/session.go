// Package session provides functions to aggregate iSCSI connections into a
// single iSCSI session.
package session

import (
	"errors"
	"fmt"

	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
)

type Pool struct {
	sessions map[uint64]*SCSI
}

type SCSI struct {
	conns []*srv.Conn
}

// Assign takes a new connection and assigns it to a SCSI session in the
// pool.  If no session exists, a new one is created.
func (p Pool) Assign(conn *srv.Conn) error {
	var txt []byte
	for conn.Recv() {
		m := conn.Msg()
		if !m.IsLogon() {
			return errors.New("connection without logon")
		}
		txt = append(txt, m.Data()...)
		if m.IsTerminal() {
			break
		}
	}
	if err := conn.Err(); err != nil {
		return err
	}
	kv, err := packet.ParseKVText(txt)
	if err != nil {
		return err
	}
	for k, v := range kv {
		fmt.Printf("%s = %s\n", k, v)
	}
	return nil
}
