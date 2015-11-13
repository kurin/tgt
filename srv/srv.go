// Package srv handles individual connections in an iSCSI session.
package srv

import (
	"io"
	"net"

	"github.com/kurin/tgt/packet"
)

type Collector struct {
	Listener net.Listener
}

func (c Collector) Collect() (*Conn, error) {
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{
		c: conn,
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
	if c.err != nil {
		return false
	}
	c.msg, c.err = packet.Next(c.c)
	return c.err == nil
}

func (c *Conn) Err() error {
	return c.err
}

func (c *Conn) Msg() *packet.Message {
	return c.msg
}

func (c *Conn) Send(msg *packet.Message) error {
	b := msg.Bytes()
	if b == nil {
		return io.ErrUnexpectedEOF
	}
	_, err := c.c.Write(b)
	return err
}
