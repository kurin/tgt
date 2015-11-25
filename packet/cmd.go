package packet

import "bytes"

func (m *Message) scsiCmdRespBytes() []byte {
	// rfc7143 11.4
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(OpSCSIResp))
	buf.WriteByte(0x80) // 11.4.1 = wtf
	buf.WriteByte(byte(m.SCSIResponse))
	buf.WriteByte(byte(m.Status))

	// Skip through to byte 16
	for i := 0; i < 3*4; i++ {
		buf.WriteByte(0x00)
	}
	buf.Write(MarshalUint64(uint64(m.TaskTag))[4:])
	for i := 0; i < 4; i++ {
		buf.WriteByte(0x00)
	}
	buf.Write(MarshalUint64(uint64(m.StatSN))[4:])
	buf.Write(MarshalUint64(uint64(m.ExpCmdSN))[4:])
	buf.Write(MarshalUint64(uint64(m.MaxCmdSN))[4:])
	for i := 0; i < 3*4; i++ {
		buf.WriteByte(0x00)
	}

	return buf.Bytes()
}

type Status byte

const (
	StatusGood        Status = 0x00
	StatusCheckCond          = 0x02
	StatusBusy               = 0x08
	StatusReservConfl        = 0x18
	StatusTaskSetFull        = 0x28
	StatusActiveACA          = 0x30
	StatusTaskAborted        = 0x40
)

type Response byte

const (
	Complete Response = 0x00
	Failure           = 0x01
)
