package scsi

import (
	"bytes"
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

type Capacity struct {
	LBA                  uint64
	Blocksize            uint32
	ProtectionType       uint8
	PIExponent           uint8
	LogicalExponent      uint8
	ThinProvisioned      bool
	ThinProvReturnsZeros bool
	LowestLBA            uint16
}

func (c *Capacity) bytes() []byte {
	// table 111
	// http://www.seagate.com/staticfiles/support/disc/manuals/Interface%20manuals/100293068c.pdf
	buf := &bytes.Buffer{}
	buf.Write(packet.MarshalUint64(c.LBA))
	buf.Write(packet.MarshalUint64(uint64(c.Blocksize))[4:])
	var b byte
	if c.ProtectionType > 0 {
		b |= 0x01
		b |= c.ProtectionType << 1
		b &= 0x0f
	}
	buf.WriteByte(b)
	b = c.PIExponent << 4
	b |= c.LogicalExponent
	buf.WriteByte(b)
	lowLBA := packet.MarshalUint64(uint64(c.LowestLBA))[6:]
	lowLBA[0] &= 0x3f
	if c.ThinProvisioned {
		lowLBA[0] &= 0x80
	}
	if c.ThinProvReturnsZeros {
		lowLBA[0] &= 0x40
	}
	return buf.Bytes()
}

type Interface interface {
	TestUnitReady() (ready bool, err error)
	ReadCapacity10(pmi bool, lba uint32) (rlba, blocksize uint32, err error)
	ReadCapacity16(pmi bool, lba uint64) (*Capacity, error)
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
		resp.OpCode = packet.OpSCSIIn
		resp.HasStatus = true
		var data []byte
		pmi := m.CDB[8]&0x01 == 0x01
		lba := uint32(packet.ParseUint(m.CDB[2:6]))
		rlba, blocksize, err := t.LUNs[lun].ReadCapacity10(pmi, lba)
		if err != nil {
			resp.Status = 0x02
			resp.SCSIResponse = 0x01
			break
		}
		data = append(data, packet.MarshalUint64(uint64(rlba))[4:]...)
		data = append(data, packet.MarshalUint64(uint64(blocksize))[4:]...)
		resp.RawData = data
	case 0x9e:
		resp.OpCode = packet.OpSCSIIn
		resp.HasStatus = true
		sa := m.CDB[1] & 0x1f
		if sa == 0x10 {
			pmi := m.CDB[14]&0x01 == 0x01
			lba := packet.ParseUint(m.CDB[2:10])
			capacity, err := t.LUNs[lun].ReadCapacity16(pmi, lba)
			if err != nil {
				resp.Status = 0x02
				resp.SCSIResponse = 0x01
				break
			}
			resp.RawData = capacity.bytes()
			break
		}
	default:
		return fmt.Errorf("no handler for CDB command %x", opCode)
	}
	return s.Send(resp)
}
