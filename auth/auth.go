// Package auth exports authentication methods for iSCSI.
package auth

import (
	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
)

type AllowAll struct{}

func NewAllowAll() srv.Handler { return &AllowAll{} }

func (a *AllowAll) Handle(m *packet.Message) (*packet.Message, error) {
	if !m.Cont && m.NSG == packet.FullFeaturePhase {
		resp := &packet.Message{
			OpCode:  packet.OpLoginResp,
			Transit: true,
			NSG:     packet.FullFeaturePhase,
			StatSN:  1,
		}
		m.Response(resp)
		return resp, nil
	}
	resp := &packet.Message{
		OpCode: packet.OpLoginResp,
		CSG:    m.CSG,
		NSG:    packet.LoginOperationalNegotiation,
	}
	m.Response(resp)
	return resp, nil
}

func (a *AllowAll) Close() error { return nil }
