// Package scsi provides a handler that implements SCSI commands.
package scsi

import (
	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
)

func New() srv.Handler { return &Handler{} }

type Handler struct{}

func (h *Handler) Handle(m *packet.Message) (*packet.Message, error) {
	resp := &packet.Message{
		OpCode:   packet.OpSCSIResp,
		StatSN:   m.ExpStatSN,
		ExpCmdSN: m.CmdSN + 1,
		MaxCmdSN: m.CmdSN + 11,
	}
	m.Response(resp)
	return resp, nil
}

func (h *Handler) Close() error { return nil }

type cdb struct {
	op opCode
}

type opCode byte

const (
	opTestUnitReady  opCode = 0x00
	opReadCapacity10        = 0x25
)

func parseCDB(data []byte) (*cdb, error) {
	return &cdb{
		op: opCode(data[0]),
	}, nil
}
