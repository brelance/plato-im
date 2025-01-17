package tcp

import (
	"encoding/binary"
)

type DataPkg struct {
	Len  uint32
	Data []byte
}

func (dp *DataPkg) Marshal() []byte {
	buf := make([]byte, 4+len(dp.Data))
	binary.Encode(buf, binary.BigEndian, dp.Len)
	copy(buf[4:], dp.Data)
	return buf
}
