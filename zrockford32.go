package zrockford32

import (
	"errors"
	"io"
	"strconv"
)

// Encodings

const encodeStd = "YBNDRFG8EJKMCPQX0T1VW2SZA345H769"
const encodeLwr = "ybndrfg8ejkmcpqx0t1vw2sza345h769"

type Encoding struct {
	encoder   string
	decodeMap [256]byte
}

func NewEncoding(encoder string) *Encoding {
	e := new(Encoding)
	e.encoder = encoder

	for i := 0; i < len(e.decodeMap); i++ {
		e.decodeMap[i] = 0xFF
	}

	for i := 0; i < len(encoder); i++ {
		e.decodeMap[encoder[i]] = byte(i)
	}

	return e
}

var StdEncoding = NewEncoding(encodeStd)
var LwrEncoding = NewEncoding(encodeLwr)

func (e *Encoding) encode(dst, src []byte, bits int) int {
	off := 0

	for i := 0; i < bits || (bits < 0 && len(src) > 0); i += 5 {
		b0 := src[0]
		b1 := byte(0)

		if len(src) > 1 {
			b1 = src[1]
		}

		char := byte(0)
		offset := uint(i % 8)

		if offset < 4 {
			char = b0 & (31 << (3 - offset)) >> (3 - offset)
		} else {
			char = b0 & (31 >> (offset - 3)) << (offset - 3)
			char |= b1 & (255 << (11 - offset)) >> (11 - offset)
		}

		if bits >= 0 && i+5 > bits {
			char &= 255 << uint((i+5)-bits)
		}

		dst[off] = e.encoder[char]
		off++

		if offset > 2 {
			src = src[1:]
		}
	}

	return off
}

func (e *Encoding) EncodeBits(dst, src []byte, bits int) int {
	if bits < 0 {
		return 0
	}

	return e.encode(dst, src, bits)
}

func (e *Encoding) Encode(dst, src []byte) int {
	return e.encode(dst, src, -1)
}

func (e *Encoding) EncodeToString(src []byte) string {
	buffer := make([]byte, e.EncodedLen(len(src)))
	n := e.Encode(buffer, src)

	return string(buffer[:n])
}

func (e *Encoding) EncodeBitsToString(src []byte, bits int) string {
	dst := make([]byte, e.EncodedLen(len((src))))
	n := e.EncodeBits(dst, src, bits)
	return string(dst[:n])
}

func (e *Encoding) EncodedLen(n int) int {
	return (n + 4) / 5 * 8
}

func (e *Encoding) DecodedLen(n int) int {
	return (n + 7) / 8 * 5
}

// Encoder

type encoder struct {
	io.WriteCloser
	encoding *Encoding
	writer   io.Writer
	buffer   [5]byte
	nbuffer  int
	output   [1024]byte
	err      error
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if e.err != nil {
		return 0, e.err
	}

	// Leading fringe.
	if e.nbuffer > 0 {
		var i int
		for i = 0; i < len(p) && e.nbuffer < 5; i++ {
			e.buffer[e.nbuffer] = p[i]
			e.nbuffer++
		}
		n += i
		p = p[i:]
		if e.nbuffer < 5 {
			return
		}
		m := e.encoding.Encode(e.output[0:], e.buffer[0:])
		if _, e.err = e.writer.Write(e.output[0:m]); e.err != nil {
			return n, e.err
		}
		e.nbuffer = 0
	}

	// Large interior chunks.
	for len(p) >= 5 {
		nn := len(e.output) / 8 * 5
		if nn > len(p) {
			nn = len(p)
			nn -= nn % 5
		}
		m := e.encoding.Encode(e.output[0:], p[0:nn])
		if _, e.err = e.writer.Write(e.output[0:m]); e.err != nil {
			return n, e.err
		}
		n += nn
		p = p[nn:]
	}

	// Trailing fringe.
	for i := 0; i < len(p); i++ {
		e.buffer[i] = p[i]
	}
	e.nbuffer = len(p)
	n += len(p)
	return
}

func (e *encoder) Close() error {
	if e.err == nil && e.nbuffer > 0 {
		m := e.encoding.Encode(e.output[0:], e.buffer[0:e.nbuffer])
		_, e.err = e.writer.Write(e.output[0:m])
		e.nbuffer = 0
	}
	return e.err
}

func NewEncoder(encoding *Encoding, writer io.Writer) io.WriteCloser {
	return &encoder{encoding: encoding, writer: writer}
}

// Decoder

type CorruptInputError int64

func (e CorruptInputError) Error() string {
	return "illegal zrockford32 data at input byte " + strconv.FormatInt(int64(e), 10)
}

func (e *Encoding) decode(dst, src []byte, bits int) (int, error) {
	offlen := len(src)
	off := 0

	for len(src) > 0 {
		var dbuffer [8]byte

		j := 0
		for ; j < 8; j++ {
			if len(src) == 0 {
				break
			}

			in := src[0]
			src = src[1:]
			dbuffer[j] = e.decodeMap[in]
			if dbuffer[j] == 0xFF {
				return off, CorruptInputError(offlen - len(src) - 1)
			}
		}

		dst[off+0] = dbuffer[0]<<3 | dbuffer[1]>>2
		dst[off+1] = dbuffer[1]<<6 | dbuffer[2]<<1 | dbuffer[3]>>4
		dst[off+2] = dbuffer[3]<<4 | dbuffer[4]>>1
		dst[off+3] = dbuffer[4]<<7 | dbuffer[5]<<2 | dbuffer[6]>>3
		dst[off+4] = dbuffer[6]<<5 | dbuffer[7]

		if bits < 0 {
			var lookup = []int{0, 1, 1, 2, 2, 3, 4, 4, 5}
			off += lookup[j]
			continue
		}

		bitsInBlock := bits
		if bitsInBlock > 40 {
			bitsInBlock = 40
		}

		off += (bitsInBlock + 7) / 8
		bits -= 40
	}

	return off, nil
}

func (e *Encoding) DecodeBits(dst, src []byte, bits int) (int, error) {
	if bits < 0 {
		return 0, errors.New("cannot decode a negative bit count")
	}

	return e.decode(dst, src, bits)
}

func (e *Encoding) Decode(dst, src []byte) (int, error) {
	return e.decode(dst, src, -1)
}

func (e *Encoding) decodeString(s string, bits int) ([]byte, error) {
	dst := make([]byte, e.DecodedLen(len(s)))
	n, err := e.decode(dst, []byte(s), bits)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}

func (e *Encoding) DecodeBitsString(s string, bits int) ([]byte, error) {
	if bits < 0 {
		return nil, errors.New("cannot decode a negative bit count")
	}

	return e.decodeString(s, bits)
}

func (e *Encoding) DecodeString(s string) ([]byte, error) {
	return e.decodeString(s, -1)
}

type decoder struct {
	io.ReadCloser
	encoding *Encoding
	reader   io.Reader
	buffer   [1024]byte
	nbuffer  int
	eof      bool
	err      error
}

func (d *decoder) Read(p []byte) (int, error) {
	var n int

	if d.nbuffer < 1 && !d.eof {
		buffer := make([]byte, 640)
		l, err := d.reader.Read(buffer)
		if io.EOF == err {
			d.eof = true
		} else if err != nil {
			return n, err
		}

		d.nbuffer, err = d.encoding.Decode(d.buffer[0:], buffer[:l])
		if err != nil {
			return n, err
		}
	}

	for n < len(p) && d.nbuffer > 0 {
		m := copy(p[n:], d.buffer[:(min(d.nbuffer, len(p)))])
		d.nbuffer -= m
		for i := 0; i < d.nbuffer; i++ {
			d.buffer[i] = d.buffer[m+i]
		}
		n += m
	}

	if d.eof == true && d.nbuffer == 0 {
		return n, io.EOF
	}

	return n, nil
}

func NewDecoder(encoding *Encoding, reader io.Reader) io.Reader {
	return &decoder{encoding: encoding, reader: reader}
}

// Misc. functions

func min(a, b int) int {
	if a <= b {
		return a
	}

	return b
}
