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
		OpCode: packet.OpSCSIResp,
		StatSN: m.ExpStatSN,
	}
	m.Response(resp)
	return resp, nil
}

func (h *Handler) Close() error { return nil }
