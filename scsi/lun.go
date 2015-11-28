package scsi

import (
	"bytes"
	"errors"

	"github.com/kurin/tgt/packet"
)

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

type InquiryData struct {
	PeripheralQualifier int
	PeripheralType      int
	Removable           bool
	Version             int
	SupportsACA         bool
	Hierarchical        bool
	SupportsSCC         bool
	HasACC              bool
	TargetGroupSupport  int
	ThirdPartyCopy      bool
	Protect             bool
	EnclosureServices   bool
	Multiport           bool
	MediaChanger        bool
	Vendor              [8]byte
	Product             [16]byte
	RevisionLevel       [4]byte
	SerialNumber        uint64
}

func (id *InquiryData) bytes() []byte {
	buf := &bytes.Buffer{}
	var b byte
	b = (uint8(id.PeripheralQualifier) << 5) & 0xe0
	b |= uint8(id.PeripheralType) & 0x1f
	buf.WriteByte(b)
	b = 0
	if id.Removable {
		b = 0x80
	}
	buf.WriteByte(b)
	buf.WriteByte(byte(id.Version))
	b = 0x02
	if id.SupportsACA {
		b |= 0x20
	}
	if id.Hierarchical {
		b |= 0x10
	}
	buf.WriteByte(b)
	buf.WriteByte(0x00)
	// byte 5
	b = 0
	if id.SupportsSCC {
		b |= 0x80
	}
	if id.HasACC {
		b |= 0x40
	}
	b |= byte(id.TargetGroupSupport) << 4 & 0x30
	if id.ThirdPartyCopy {
		b |= 0x08
	}
	if id.Protect {
		b |= 0x01
	}
	buf.WriteByte(b)
	// byte 6
	b = 0
	if id.EnclosureServices {
		b |= 0x40
	}
	if id.Multiport {
		b |= 0x10
	}
	if id.MediaChanger {
		b |= 0x08
	}
	buf.WriteByte(b)
	buf.WriteByte(0x02)
	buf.Write(id.Vendor[:])
	buf.Write(id.Product[:])
	buf.Write(id.RevisionLevel[:])
	buf.Write(packet.MarshalUint64(id.SerialNumber))
	for i := 0; i < 12; i++ {
		buf.WriteByte(0x00)
	}
	data := buf.Bytes()
	data[4] = byte(len(data) - 4)
	return data
}

type Interface interface {
	TestUnitReady() (ready bool, err error)
	ReadCapacity10(pmi bool, lba uint32) (rlba, blocksize uint32, err error)
	ReadCapacity16(pmi bool, lba uint64) (*Capacity, error)
	Inquiry() (*InquiryData, error)
	VitalProductData(code byte) ([]byte, error)
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

func (t *Target) handleSCSICmd(s *Session, m *packet.Message) (err error) {
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
	defer func() {
		if err == nil {
			err = s.Send(resp)
		}
	}()
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
		switch sa {
		case 0x10:
			pmi := m.CDB[14]&0x01 == 0x01
			lba := packet.ParseUint(m.CDB[2:10])
			capacity, err := t.LUNs[lun].ReadCapacity16(pmi, lba)
			if err != nil {
				resp.Status = 0x02
				resp.SCSIResponse = 0x01
				break
			}
			resp.RawData = capacity.bytes()
		}
	case 0x12:
		resp.OpCode = packet.OpSCSIIn
		resp.HasStatus = true
		alloc := int(packet.ParseUint(m.CDB[3:5]))
		evpd := m.CDB[1]&0x01 == 0x01
		if evpd {
			t.handleVPD(s, m, resp)
			return
		}
		inq, err := t.LUNs[lun].Inquiry()
		if err != nil {
			resp.Status = 0x02
			resp.SCSIResponse = 0x01
			break
		}
		resp.RawData = inq.bytes()[:alloc]
	default:
		setError(resp, ErrUnsupportedCommand)
	}
	return
}

func setError(resp *packet.Message, serr SCSIError) {
	buf := &bytes.Buffer{}
	data := serr.Bytes()
	buf.Write(packet.MarshalUint64(uint64(len(data)))[6:])
	buf.Write(data)
	resp.RawData = buf.Bytes()
}

func (t *Target) handleVPD(s *Session, m, resp *packet.Message) {
	lun := int(m.LUN)
	alloc := int(packet.ParseUint(m.CDB[3:5]))
	vpd, err := t.LUNs[lun].VitalProductData(m.CDB[2])
	if err != nil {
		resp.Status = 0x02
		resp.SCSIResponse = 0x01
		return
	}
	resp.RawData = vpd
	if len(vpd) >= alloc {
		resp.RawData = resp.RawData[:alloc]
	}
}
