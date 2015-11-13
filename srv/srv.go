// Package srv handles individual connections in an iSCSI session.
package srv

import (
	"errors"
	"net"

	"github.com/kurin/tgt/packet"
)

type Collector struct {
	Listener net.Listener
}

func (c Collector) Collect() (*Conn, error) {
	c, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{
		c: c,
	}, nil
}

func (c Collector) Close() error {
	return c.Listener.Close()
}

type Conn struct {
	c   net.Conn
	err error
	msg *packet.Message
}

func (c *Conn) Close() error {
	return c.Close()
}

func (c *Conn) Recv() bool {
	c.msg, c.err = packet.Next(c.c)
	return c.err == nil
}

func (c *Conn) Send(msg *packet.Message) error {
	return errors.New("not implemented")
}
