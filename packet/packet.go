// Package packet implements the iSCSI PDU packet format as specified in
// rfc7143 section 11.
package packet

import (
	"fmt"
	"io"
	"strings"
)

type opCode int

const (
	// Defined on the initiator.
	opNoopOut     opCode = 0x00
	opSCSICmd            = 0x01
	opSCSITaskReq        = 0x02
	opLoginReq           = 0x03
	opTextReq            = 0x04
	opSCSIOut            = 0x05
	opLogoutReq          = 0x06
	opSNACKReq           = 0x10
	// Defined on the target.
	opNoopIn       opCode = 0x20
	opSCSIResp            = 0x21
	opSCSITaskResp        = 0x22
	opLoginResp           = 0x23
	opTextResp            = 0x24
	opSCSIIn              = 0x25
	opLogoutResp          = 0x26
	opReady               = 0x31
	opAsync               = 0x32
	opReject              = 0x3f
)

var opCodeMap = map[opCode]string{
	opNoopOut:      "NOP-Out",
	opSCSICmd:      "SCSI Command",
	opSCSITaskReq:  "SCSI Task Management FunctionRequest",
	opLoginReq:     "Login Request",
	opTextReq:      "Text Request",
	opSCSIOut:      "SCSI Data-Out (write)",
	opLogoutReq:    "Logout Request",
	opSNACKReq:     "SNACK Request",
	opNoopIn:       "NOP-In",
	opSCSIResp:     "SCSI Response",
	opSCSITaskResp: "SCSI Task Management Function Response",
	opLoginResp:    "Login Response",
	opTextResp:     "Text Response",
	opSCSIIn:       "SCSI Data-In (read)",
	opLogoutResp:   "Logout Response",
	opReady:        "Ready To Transfer (R2T)",
	opAsync:        "Asynchronous Message",
	opReject:       "Reject",
}

func (c opCode) String() string {
	s := opCodeMap[c]
	if s == "" {
		s = fmt.Sprintf("Unknown Code: %x", int(c))
	}
	return s
}

type Message struct {
	opCode    opCode
	header    []byte
	final     bool
	immediate bool
	dataLen   int
	data      []byte
	taskTag   uint32
	ahsLen    int

	connID    uint16
	cmdSN     uint32
	expStatSN uint32

	// valid when opCode == opSCSICmd
	read, write bool
	lun         int

	// valid for opLoginReq
	transit  bool
	cont     bool
	csg, nsg int
	isid     uint64
	tsih     uint16
}

func (m *Message) Data() []byte {
	return m.data[:m.dataLen]
}

func (m *Message) String() string {
	var s []string
	s = append(s, fmt.Sprintf("Op: %v", m.opCode))
	s = append(s, fmt.Sprintf("Final = %v", m.final))
	s = append(s, fmt.Sprintf("Immediate = %v", m.immediate))
	s = append(s, fmt.Sprintf("Total AHS Length = %d", m.ahsLen))
	s = append(s, fmt.Sprintf("Data Segment Length = %d", m.dataLen))
	s = append(s, fmt.Sprintf("Task Tag = %x", m.taskTag))
	switch m.opCode {
	case opSCSICmd:
		s = append(s, fmt.Sprintf("LUN = %d", m.lun))
	case opLoginReq:
		s = append(s, fmt.Sprintf("Transit = %v", m.transit))
		s = append(s, fmt.Sprintf("Continue = %v", m.cont))
		s = append(s, fmt.Sprintf("Connection ID = %x", m.connID))
	}
	return strings.Join(s, "\n")
}

func Next(r io.Reader) (*Message, error) {
	buf := make([]byte, 48) // sync.Pool
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	m, err := parseHeader(buf)
	if err != nil {
		return nil, err
	}
	m.header = buf
	if m.dataLen > 0 {
		dl := m.dataLen
		for dl%4 > 0 {
			dl++
		}
		data := make([]byte, dl)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		m.data = data
	}
	return m, nil
}

// parseUint parses the given slice as a network-byte-ordered integer.  If
// there are more than 8 bytes in data, it overflows.
func parseUint(data []byte) uint64 {
	var out uint64
	for i := 0; i < len(data); i++ {
		out += uint64(data[len(data)-i-1]) << uint(8*i)
	}
	return out
}

func parseHeader(data []byte) (*Message, error) {
	if len(data) < 48 {
		return nil, fmt.Errorf("garbled header")
	}
	// TODO: sync.Pool
	m := &Message{}
	m.immediate = 0x40&data[0] == 0x40
	m.opCode = opCode(data[0] & 0x3f)
	m.final = 0x80&data[1] == 0x80
	m.ahsLen = int(data[4]) * 4
	m.dataLen = int(parseUint(data[5:8]))
	m.taskTag = uint32(parseUint(data[16:20]))
	switch m.opCode {
	case opSCSICmd:
		m.lun = int(parseUint(data[8:16]))
	case opLoginReq:
		m.transit = m.final
		m.cont = data[1]&0x40 == 0x40
		if m.cont && m.transit {
			// rfc7143 11.12.2
			return nil, fmt.Errorf("transit and continue bits set in same login request")
		}
		m.csg = int(data[1]&0xc) >> 2
		m.nsg = int(data[1] & 0x3)
		m.connID = uint16(parseUint(data[20:22]))
		m.cmdSN = uint32(parseUint(data[24:28]))
		m.expStatSN = uint32(parseUint(data[28:32]))
	}
	return m, nil
}

func (m *Message) IsLogon() bool {
	return m.opCode == opLoginReq
}

func (m *Message) IsTerminal() bool {
	return m.final
}
