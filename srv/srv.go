// Package srv handles individual connections in an iSCSI session.
package srv

import (
	"bytes"
	"fmt"
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
	ID uint16

	c    net.Conn
	err  error
	msg  *packet.Message
	auth bool
}

func (c *Conn) Close() error {
	return c.Close()
}

func (c *Conn) Recv() bool {
	if c.err != nil {
		return false
	}
	c.msg, c.err = packet.Next(c.c)
	///
	if c.err == nil {
		fmt.Println("-----------------------")
		fmt.Println(c.msg)
	}
	///
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
	///
	m, errr := packet.Next(bytes.NewReader(b))
	if errr != nil {
		return fmt.Errorf("rawr: %v", errr)
	}
	fmt.Println("******************")
	fmt.Println(m)
	///
	if b == nil {
		return io.ErrUnexpectedEOF
	}
	_, err := c.c.Write(b)
	return err
}

// Handler receives iSCSI messages and, optionally, returns an iSCSI response.
type Handler interface {
	// Handle takes a single iSCSI PDU and composes an appropriate response.
	// Further messages with the same Task Tag will be mapped to the same
	// handler.
	Handle(m *packet.Message) (*packet.Message, error)
	Close() error
}
