package scsi

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/kurin/tgt/packet"
)

type PortalGroup struct {
	L       net.Listener
	Targets []*Target
}

func (pg *PortalGroup) Accept() (*Conn, error) {
	conn, err := pg.L.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{
		c: conn,
	}, nil
}

func (pg *PortalGroup) Close() error {
	return pg.L.Close()
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
	fmt.Printf("=== outgoing msglen: %d", len(b))
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
