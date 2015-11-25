package scsi

import (
	"errors"
	"fmt"

	"github.com/kurin/tgt/packet"
)

type scsiError struct {
	status   byte
	response byte
	sense    byte
	code     string
}

type Interface interface {
	TestUnitReady() (ready bool, err error)
	ReadCapacity10(pmi bool, lba uint32) (rlba, blocksize uint32, err error)
}

type Target struct {
	Name string
	LUNs []Interface
}

func (t *Target) handleAuth(s *Session, m *packet.Message) error {
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

func (t *Target) handleSCSICmd(s *Session, m *packet.Message) error {
	opCode := m.CDB[0]
	lun := int(m.LUN)
	resp := &packet.Message{
		OpCode:   packet.OpSCSIResp,
		Final:    true,
		StatSN:   m.ExpStatSN,
		TaskTag:  m.TaskTag,
		ExpCmdSN: m.CmdSN + 1,
		MaxCmdSN: m.CmdSN + 10,
	}
	switch opCode {
	case 0x00:
		ready, err := t.LUNs[lun].TestUnitReady()
		if err != nil {
			resp.Status = 0x02
			resp.SCSIResponse = 0x01
		}
		if !ready {
			resp.Status = 0x08
			resp.SCSIResponse = 0x01
		}
		return s.Send(resp)
	case 0x25:
		var data []byte
		pmi := m.CDB[8]&0x01 == 0x01
		lba := uint32(packet.ParseUint(m.CDB[2:6]))
		rlba, blocksize, err := t.LUNs[lun].ReadCapacity10(pmi, lba)
		if err != nil {
			resp.Status = 0x02
			resp.SCSIResponse = 0x01
		}
		data = append(data, packet.MarshalUint64(uint64(rlba))[4:]...)
		data = append(data, packet.MarshalUint64(uint64(blocksize))[4:]...)
		resp.OpCode = packet.OpSCSIIn
		resp.RawData = data
		resp.HasStatus = true
	default:
		return fmt.Errorf("no handler for CDB command %x", opCode)
	}
	return s.Send(resp)
}
