package server

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrMsgOverflow indicates that message is too big for message buffer.
var ErrMsgOverflow = errors.New("message buffer overflow")

const msgSizeBytes = 4

func startReader(c connWithErrHandler, msgs pool, bufSize int) chan []byte {
	out := make(chan []byte, 1)

	go func() {
		in := make([]byte, bufSize)
		sizeBuf := make([]byte, 0, msgSizeBytes)

		var (
			msgBuf []byte
			size   uint32
		)
		defer func() {
			if msgBuf != nil {
				msgs.put(msgBuf)
			}

			close(out)
		}()

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

						out <- msgBuf
						msgBuf = nil
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
						sizeBuf = sizeBuf[:0]

						msgBuf = msgs.get()
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
	}()

	return out
}
