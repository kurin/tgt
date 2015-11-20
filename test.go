package main

import (
	"log"
	"net"

	"github.com/kurin/tgt/auth"
	"github.com/kurin/tgt/packet"
	"github.com/kurin/tgt/scsi"
	"github.com/kurin/tgt/session"
	"github.com/kurin/tgt/srv"
	"github.com/kurin/tgt/target"
)

func main() {
	l, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal(err)
	}
	c := &srv.Collector{
		Listener: l,
	}

	pool := &session.Pool{}
	pool.RegisterTarget(&target.Target{
		Name: "a",
		TaskMap: map[packet.OpCode]func() srv.Handler{
			packet.OpLoginReq: auth.NewAllowAll,
			packet.OpSCSICmd:  scsi.New,
		},
	})

	conn, err := c.Collect()
	if err != nil {
		log.Fatal(err)
	}
	s, err := pool.Accept(conn)
	if err != nil {
		log.Fatal(err)
	}
	for {
		m := s.Recv()
		resp, err := s.Dispatch(m)
		if err != nil {
			log.Fatal(err)
		}
		if err := s.Send(resp); err != nil {
			log.Fatal(err)
		}
	}
}
