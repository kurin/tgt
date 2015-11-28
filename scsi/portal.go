package scsi

import (
	"bytes"
	"errors"
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

func (pg *PortalGroup) Attach(c *Conn) (*Session, error) {
	s, err := NewSession()
	if err != nil {
		return nil, err
	}
	s.conn = c
	if !c.Recv() {
		return nil, fmt.Errorf("no packets on new connection %v", c)
	}
	m := c.Msg()
	kv := packet.ParseKVText(m.RawData)
	tname := kv["TargetName"]
	if tname == "" {
		return nil, errors.New("no target name given")
	}
	var tgt *Target
	for _, t := range pg.Targets {
		if t.Name == tname {
			tgt = t
		}
	}
	if tgt == nil {
		return nil, fmt.Errorf("no target: %v", tname)
	}
	s.target = tgt
	go func() {
		s.messages <- m
		for c.Recv() {
			s.messages <- c.Msg()
		}
	}()
	return s, nil
}

type Conn struct {
	ID uint16

	c    net.Conn
	err  error
	msg  *packet.Message
	auth bool
}

func (c *Conn) Close() error {
	return c.c.Close()
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
	fmt.Printf("=== outgoing msglen: %d\n", len(b))
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
