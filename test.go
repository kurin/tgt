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

func (fs *FakeSCSI) ReadCapacity16(bool, uint64) (*scsi.Capacity, error) {
	return &scsi.Capacity{}, nil
}

func main() {

	t := &scsi.Target{
		// Target-specific options
		Name: "a",
		LUNs: []scsi.Interface{
			&FakeSCSI{}, // lun 0
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
		session, err := pg.Attach(conn)
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			if err := session.Run(); err != nil {
				log.Println(err)
			}
		}()
	}
}
