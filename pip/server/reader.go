package server

import (
	"encoding/binary"
	"errors"
	"io"
)

var ErrMsgOverflow = errors.New("message buffer overflow")

const msgSizeBytes = 4

func read(c connWithErrHandler, bufSize, maxMsgSize int) {
	in := make([]byte, bufSize)
	sizeBuf := make([]byte, 0, msgSizeBytes)

	msgBuf := make([]byte, 0, maxMsgSize)
	var size uint32

	for {
		n, err := c.c.Read(in)
		if n > 0 {
			b := in[:n]
			for len(b) > 0 {
				if size > 0 {
					m := int(size)
					if m > len(b) {
						msgBuf = append(msgBuf, b...)
						size -= uint32(len(b))
						b = b[len(b):]

						continue
					}

					msgBuf = append(msgBuf, b[:m]...)
					b = b[m:]
					size = 0
				} else {
					m := cap(sizeBuf) - len(sizeBuf)
					if m > len(b) {
						sizeBuf = append(sizeBuf, b...)
						b = b[len(b):]

						continue
					}

					sizeBuf = append(sizeBuf, b[:m]...)
					b = b[m:]

					size = binary.LittleEndian.Uint32(sizeBuf)
					msgBuf = msgBuf[:0]
					if size > uint32(cap(msgBuf)) {
						c.handle(ErrMsgOverflow)
						return
					}
				}
			}
		}

		if err != nil {
			if err != io.EOF {
				c.handle(err)
			}

			break
		}
	}
}
