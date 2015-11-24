package main

import (
	"log"
	"net"
)

type FakeSCSI struct{}

func main() {
	lun := &scsi.LUN{
		// LUN specific options
		Device: &FakeSCSI{},
	}

	t := &scsi.Target{
		// Target-specific options
		LUNS: []*scsi.LUN{
			lun, // lun 0
		},
	}

	pg := &scsi.PortalGroup{
		// Portal-Group specific options, like auth
		Targets: []*scsi.Target{
			t,
		},
	}

	l, err := net.Listen("tcp", ":3260")
	if err != nil {
		log.Fatal(err)
	}

	pg.AddListener(l)

	for {
		conn, err := pg.Accept()
		session, err := scsi.Attach(conn)
		go session.Run()
	}
}
