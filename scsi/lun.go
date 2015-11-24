package scsi

type Target struct {
	LUNS []*LUN
}

type LUN struct {
	Device Interface
}

type Interface interface {
	TestUnitReady() (bool, error)
	ReadCapacity10(bool, uint32) (uint32, uint32, error)
}
