// Package target provides functionality for iSCSI targets.
package target

import (
	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/srv"
)

// Target is an iSCSI target.
type Target struct {
	Name    string
	Alias   string
	TaskMap map[packet.OpCode]func() srv.Handler
}
