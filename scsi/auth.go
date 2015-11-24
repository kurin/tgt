package scsi

import (
	"errors"

	"github.com/kurin/tgt/packet"
)

type authHandler struct{}

func (a *authHandler) handle(s *Session, m *packet.Message) error {
	if !m.Cont && m.NSG == packet.FullFeaturePhase {
		resp := &packet.Message{
			OpCode:   packet.OpLoginResp,
			Transit:  true,
			NSG:      packet.FullFeaturePhase,
			StatSN:   m.ExpStatSN,
			TaskTag:  m.TaskTag,
			ExpCmdSN: m.CmdSN,
			MaxCmdSN: m.CmdSN,
			RawData: packet.MarshalKVText(map[string]string{
				"HeaderDigest": "None",
				"DataDigest":   "None",
			}),
		}
		return s.Send(resp)
	}
	return errors.New("can't perform auth")
}
