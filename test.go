package main

import (
	"log"
	"net"

	"github.com/kurin/tgt/scsi"
)

type FakeSCSI struct{}

func (fs *FakeSCSI) TestUnitReady() (bool, error) {
	return true, nil
}

func (fs *FakeSCSI) ReadCapacity10(bool, uint32) (lba, blocksize uint32, err error) {
	return 0xffff, 0x200, nil
}

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

	l, err := net.Listen("tcp", ":3260")
	if err != nil {
		log.Fatal(err)
	}

	pg := &scsi.PortalGroup{
		// Portal-Group specific options, like auth
		Targets: []*scsi.Target{
			t,
		},
		L: l,
	}

	for {
		conn, err := pg.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		session, err := scsi.Attach(conn)
		if err != nil {
			log.Println(err)
			continue
		}
		go session.Run()
	}
}
