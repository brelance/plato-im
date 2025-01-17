package tcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func ReadData(conn *net.TCPConn) ([]byte, error) {
	var dataLen uint32
	dataLenBuf := make([]byte, 4)

	readFixedData(conn, dataLenBuf)
	_, err := binary.Decode(dataLenBuf, binary.BigEndian, &dataLen)
	if err != nil {
		return nil, fmt.Errorf("Read headlen error:%s", err.Error())
	}

	if dataLen <= 0 {
		return nil, fmt.Errorf("wrong headlen:%d", dataLen)
	}

	dataBuf := make([]byte, dataLen)
	err = readFixedData(conn, dataBuf)
	if err != nil {
		return nil, fmt.Errorf("Read data error:%s", err.Error())
	}
	return dataBuf, err
}

func readFixedData(conn *net.TCPConn, buf []byte) error {
	err := (*conn).SetReadDeadline(time.Now().Add(time.Duration(120) * time.Second))
	if err != nil {
		return err
	}

	totalReadLen := len(buf)
	pos := 0
	for {
		len, err := conn.Read(buf[pos:])
		if err != nil {
			return err
		}
		pos += len
		if pos >= totalReadLen {
			break
		}
	}
	return nil
}
