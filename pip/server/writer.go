package server

import (
	"encoding/binary"
	"time"
)

func write(c connWithErrHandler, in chan []byte, msgs pool, bufSize int, d time.Duration) {
	t := time.NewTicker(d)
	defer t.Stop()

	out := make([]byte, 0, bufSize)
	sizeBuf := make([]byte, msgSizeBytes)

	for {
		select {
		case msg, ok := <-in:
			if !ok {
				flush(c, out)
				return
			}

			b := msg
			s := sizeBuf
			binary.LittleEndian.PutUint32(s, uint32(len(b)))

			for len(s) > 0 {
				m := cap(out) - len(out)
				if m > len(s) {
					out = append(out, s...)
					break
				}

				out = append(out, s[:m]...)
				s = s[m:]

				if !flush(c, out) {
					msgs.put(msg)
					go ignore(in, msgs)
					return
				}

				out = out[:0]
			}

			for len(b) > 0 {
				m := cap(out) - len(out)
				if m > len(b) {
					out = append(out, b...)
					break
				}

				out = append(out, b[:m]...)
				b = b[m:]

				if !flush(c, out) {
					msgs.put(msg)
					go ignore(in, msgs)
					return
				}

				out = out[:0]
			}
			msgs.put(msg)

		case <-t.C:
			if !flush(c, out) {
				go ignore(in, msgs)
				return
			}

			out = out[:0]
		}
	}
}

func flush(c connWithErrHandler, b []byte) bool {
	if len(b) > 0 {
		_, err := c.c.Write(b)
		if err != nil {
			c.handle(err)
			return false
		}
	}

	return true
}

func ignore(in chan []byte, msgs pool) {
	for msg := range in {
		msgs.put(msg)
	}
}
