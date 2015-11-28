package scsi

type SCSIError interface {
	error
	Bytes() []byte
}

type scsiError struct {
	reason    string
	status    byte
	response  byte
	sense     byte
	code      byte
	qualifier byte
}

var (
	ErrUnsupportedCommand = &scsiError{
		reason:   "illegal request: inavlid/unsupported command",
		status:   0x00,
		response: 0x02,
		sense:    0x05,
		code:     0x20,
	}
)

func (se *scsiError) Error() string {
	return se.reason
}

func (se *scsiError) Bytes() []byte {
	data := make([]byte, 20)
	data[0] = 0x70
	data[1] = se.sense & 0x0f
	data[7] = 20 - 7
	data[12] = se.code
	data[13] = se.qualifier
	return data
}
