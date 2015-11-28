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

func (fs *FakeSCSI) Inquiry() (*scsi.InquiryData, error) {
	return &scsi.InquiryData{
		Vendor:        [8]byte{'1', '1', 'c', 'a', 'n', 's'},
		Product:       [16]byte{'c', 'o', 'f', 'f', 'e', 'e'},
		RevisionLevel: [4]byte{'1', '.', '0'},
		SerialNumber:  52,
	}, nil
}

func (fs *FakeSCSI) VitalProductData(code byte) ([]byte, error) {
	switch code {
	case 0xb0, 0xb1:
		data := make([]byte, 64)
		data[1] = code
		data[3] = 0x3c
		return data, nil
	}
	return nil, scsi.ErrUnsupportedCommand
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
