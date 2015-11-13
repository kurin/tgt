// Package auth exports authentication methods for iSCSI.
package auth

import (
	"errors"

	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
)

type AllowAll struct {
	tsih   uint16
	connID uint16
}

func (a *AllowAll) Own(conn *srv.Conn) error {
	for conn.Recv() {
		m := conn.Msg()
		if !(m.OpCode == packet.OpLoginReq) {
			return errors.New("auth: not a login packet")
		}
		if a.tsih == 0 {
			a.tsih = m.TSIH
			a.connID = m.ConnID
		}
		if m.Transit && m.NSG == packet.FullFeaturePhase {
			resp := &packet.Message{
				OpCode: packet.OpLoginResp,
				TSIH:   42,
			}
			if err := conn.Send(resp); err != nil {
				return err
			}
		}
	}
	return conn.Err()
}

func (a *AllowAll) TSIH() uint16 {
	return a.tsih
}

func (a *AllowAll) ConnID() uint16 {
	return a.connID
}
