package zrockford32_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/checksum0/go-zrockford32"
)

type bitTestCase struct {
	bits    int
	decoded []byte
	encoded string
}

var bitTestsLwr = []bitTestCase{
	// Test cases from the spec
	{0, []byte{}, ""},
	{1, []byte{0}, "y"},
	{1, []byte{128}, "0"},
	{2, []byte{64}, "e"},
	{2, []byte{192}, "a"},
	{10, []byte{0, 0}, "yy"},
	{10, []byte{128, 128}, "0n"},
	{20, []byte{139, 136, 128}, "tqre"},
	{24, []byte{240, 191, 199}, "6n9hq"},
	{24, []byte{212, 122, 4}, "4t7ye"},
	// Note: this test varies from what's in the spec by one character!
	{30, []byte{245, 87, 189, 12}, "62m54d"},

	// Edge cases we stumbled on, that are not covered above.
	{8, []byte{0xff}, "9h"},
	{11, []byte{0xff, 0xE0}, "990"},
	{40, []byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{48, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999h"},
	{192, []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "ab3sr12x8fhfnvzae075fkn3a7xh8vdk6js22k0"},

	// Used in the docs.
	{20, []byte{0x10, 0x11, 0x10}, "nyet"},
	{24, []byte{0x10, 0x11, 0x10}, "nyety"},
}

var bitTestsStd = []bitTestCase{
	// Test cases from the spec
	{0, []byte{}, ""},
	{1, []byte{0}, "Y"},
	{1, []byte{128}, "0"},
	{2, []byte{64}, "E"},
	{2, []byte{192}, "A"},
	{10, []byte{0, 0}, "YY"},
	{10, []byte{128, 128}, "0N"},
	{20, []byte{139, 136, 128}, "TQRE"},
	{24, []byte{240, 191, 199}, "6N9HQ"},
	{24, []byte{212, 122, 4}, "4T7YE"},
	// Note: this test varies from what's in the spec by one character!
	{30, []byte{245, 87, 189, 12}, "62M54D"},

	// Edge cases we stumbled on, that are not covered above.
	{8, []byte{0xff}, "9H"},
	{11, []byte{0xff, 0xE0}, "990"},
	{40, []byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{48, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999H"},
	{192, []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "AB3SR12X8FHFNVZAE075FKN3A7XH8VDK6JS22K0"},

	// Used in the docs.
	{20, []byte{0x10, 0x11, 0x10}, "NYET"},
	{24, []byte{0x10, 0x11, 0x10}, "NYETY"},
}

type byteTestCase struct {
	decoded []byte
	encoded string
}

var byteTestsLwr = []byteTestCase{
	// Byte-aligned test cases from the spec
	{[]byte{240, 191, 199}, "6n9hq"},
	{[]byte{212, 122, 4}, "4t7ye"},

	// Edge cases we stumbled on, that are not covered above.
	{[]byte{0xff}, "9h"},
	{[]byte{0xb5}, "sw"},
	{[]byte{0x34, 0x5a}, "gtpy"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999h"},
	{[]byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "ab3sr12x8fhfnvzae075fkn3a7xh8vdk6js22k0"},
}

var byteTestsStd = []byteTestCase{
	// Byte-aligned test cases from the spec
	{[]byte{240, 191, 199}, "6N9HQ"},
	{[]byte{212, 122, 4}, "4T7YE"},

	// Edge cases we stumbled on, that are not covered above.
	{[]byte{0xff}, "9H"},
	{[]byte{0xb5}, "SW"},
	{[]byte{0x34, 0x5a}, "GTPY"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999H"},
	{[]byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "AB3SR12X8FHFNVZAE075FKN3A7XH8VDK6JS22K0"},
}

func TestEncodeBitsLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range bitTestsLwr {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range bitTestsStd {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsStringLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range bitTestsLwr {
		s := enc.EncodeBitsToString(tc.decoded, tc.bits)
		if g, e := s, tc.encoded; g != e {
			t.Errorf("EncodeBitsToString %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsStringStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range bitTestsStd {
		s := enc.EncodeBitsToString(tc.decoded, tc.bits)
		if g, e := s, tc.encoded; g != e {
			t.Errorf("EncodeBitsToString %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBytesLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range byteTestsLwr {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.Encode(dst, tc.decoded)
		dst = dst[:n]

		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("Encode %x wrong result: %q != %q", tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBytesStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range byteTestsStd {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.Encode(dst, tc.decoded)
		dst = dst[:n]

		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("Encode %x wrong result: %q != %q", tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsMasksExcessLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range []bitTestCase{
		{0, []byte{255, 255}, ""},
		{1, []byte{255, 255}, "0"},
		{2, []byte{255, 255}, "a"},
		{3, []byte{255, 255}, "h"},
		{4, []byte{255, 255}, "6"},
		{5, []byte{255, 255}, "9"},
		{6, []byte{255, 255}, "90"},
		{7, []byte{255, 255}, "9a"},
		{8, []byte{255, 255}, "9h"},
		{9, []byte{255, 255}, "96"},
		{10, []byte{255, 255}, "99"},
		{11, []byte{255, 255}, "990"},
		{12, []byte{255, 255}, "99a"},
		{13, []byte{255, 255}, "99h"},
		{14, []byte{255, 255}, "996"},
		{15, []byte{255, 255}, "999"},
		{16, []byte{255, 255}, "9990"},
	} {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
		}
	}
}

func TestEncodeBitsMasksExcessStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range []bitTestCase{
		{0, []byte{255, 255}, ""},
		{1, []byte{255, 255}, "0"},
		{2, []byte{255, 255}, "A"},
		{3, []byte{255, 255}, "H"},
		{4, []byte{255, 255}, "6"},
		{5, []byte{255, 255}, "9"},
		{6, []byte{255, 255}, "90"},
		{7, []byte{255, 255}, "9A"},
		{8, []byte{255, 255}, "9H"},
		{9, []byte{255, 255}, "96"},
		{10, []byte{255, 255}, "99"},
		{11, []byte{255, 255}, "990"},
		{12, []byte{255, 255}, "99A"},
		{13, []byte{255, 255}, "99H"},
		{14, []byte{255, 255}, "996"},
		{15, []byte{255, 255}, "999"},
		{16, []byte{255, 255}, "9990"},
	} {
		dst := make([]byte, enc.EncodedLen(len(tc.decoded)))
		n := enc.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
		}
	}
}

func TestEncoderLwr(t *testing.T) {
	for _, tc := range byteTestsLwr {
		for bs := int64(1); bs < 128; bs += 4 {
			in := bytes.NewReader(tc.decoded)
			buf := new(bytes.Buffer)
			enc := zrockford32.NewEncoder(zrockford32.LwrEncoding, buf)
			for {
				if _, err := io.CopyN(enc, in, bs); io.EOF == err {
					break
				} else if nil != err {
					t.Errorf("Failed to encode: %v", err)
				}
			}
			if err := enc.Close(); nil != err {
				t.Errorf("Failed to close encoder: %v", err)
			}

			if g, e := buf.String(), tc.encoded; g != e {
				t.Errorf("Encode %x wrong result: %q != %q", tc.decoded, g, e)
				continue
			}
		}
	}
}

func TestEncoderStd(t *testing.T) {
	for _, tc := range byteTestsStd {
		for bs := int64(1); bs < 128; bs += 4 {
			in := bytes.NewReader(tc.decoded)
			buf := new(bytes.Buffer)
			enc := zrockford32.NewEncoder(zrockford32.StdEncoding, buf)
			for {
				if _, err := io.CopyN(enc, in, bs); io.EOF == err {
					break
				} else if nil != err {
					t.Errorf("Failed to encode: %v", err)
				}
			}
			if err := enc.Close(); nil != err {
				t.Errorf("Failed to close encoder: %v", err)
			}

			if g, e := buf.String(), tc.encoded; g != e {
				t.Errorf("Encode %x wrong result: %q != %q", tc.decoded, g, e)
				continue
			}
		}
	}
}

func TestDecodeBitsLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range bitTestsLwr {
		dst := make([]byte, enc.DecodedLen(len(tc.encoded)))
		n, err := enc.DecodeBits(dst, []byte(tc.encoded), tc.bits)
		dst = dst[:n]
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBitsStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range bitTestsStd {
		dst := make([]byte, enc.DecodedLen(len(tc.encoded)))
		n, err := enc.DecodeBits(dst, []byte(tc.encoded), tc.bits)
		dst = dst[:n]
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBitsStringLwr(t *testing.T) {
	for _, tc := range bitTestsLwr {
		dec, err := zrockford32.LwrEncoding.DecodeBitsString(tc.encoded, tc.bits)
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dec, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBitsStringStd(t *testing.T) {
	for _, tc := range bitTestsStd {
		dec, err := zrockford32.StdEncoding.DecodeBitsString(tc.encoded, tc.bits)
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dec, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBytesLwr(t *testing.T) {
	enc := zrockford32.LwrEncoding
	for _, tc := range byteTestsLwr {
		dst := make([]byte, enc.DecodedLen(len(tc.encoded)))
		n, err := enc.Decode(dst, []byte(tc.encoded))
		dst = dst[:n]
		if err != nil {
			t.Errorf("Decode %q: error: %v", tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("Decode %q, %x != %x", tc.encoded, g, e)
		}
	}
}

func TestDecodeBytesStd(t *testing.T) {
	enc := zrockford32.StdEncoding
	for _, tc := range byteTestsStd {
		dst := make([]byte, enc.DecodedLen(len(tc.encoded)))
		n, err := enc.Decode(dst, []byte(tc.encoded))
		dst = dst[:n]
		if err != nil {
			t.Errorf("Decode %q: error: %v", tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("Decode %q, %x != %x", tc.encoded, g, e)
		}
	}
}

func TestDecodeBadLwr(t *testing.T) {
	input := `f00!bar`
	_, err := zrockford32.LwrEncoding.DecodeString(input)
	switch err := err.(type) {
	case nil:
		t.Fatalf("expected error from bad decode")
	case zrockford32.CorruptInputError:
		if g, e := err.Error(), `illegal zrockford32 data at input byte 3`; g != e {
			t.Fatalf("wrong error: %q != %q", g, e)
		}
	default:
		t.Fatalf("wrong error from bad decode: %T: %v", err, err)
	}
}

func TestDecodeBadStd(t *testing.T) {
	input := `F00!BAR`
	_, err := zrockford32.StdEncoding.DecodeString(input)
	switch err := err.(type) {
	case nil:
		t.Fatalf("expected error from bad decode")
	case zrockford32.CorruptInputError:
		if g, e := err.Error(), `illegal zrockford32 data at input byte 3`; g != e {
			t.Fatalf("wrong error: %q != %q", g, e)
		}
	default:
		t.Fatalf("wrong error from bad decode: %T: %v", err, err)
	}
}

func TestDecoderLwr(t *testing.T) {
	for _, tc := range byteTestsLwr {
		for bs := int64(1); bs < 128; bs += 4 {
			var buf bytes.Buffer
			dec := zrockford32.NewDecoder(zrockford32.LwrEncoding, bytes.NewReader([]byte(tc.encoded)))
			for {
				if _, err := io.CopyN(&buf, dec, bs); io.EOF == err {
					break
				} else if nil != err {
					t.Errorf("Failed to decode: %v", err)
				}
			}
			if g, e := buf.String(), string(tc.decoded); g != e {
				t.Errorf("Decode %x wrong result: %q != %q", tc.decoded, g, e)
			}
		}
	}
}

func TestDecoderStd(t *testing.T) {
	for _, tc := range byteTestsStd {
		for bs := int64(1); bs < 128; bs += 4 {
			var buf bytes.Buffer
			dec := zrockford32.NewDecoder(zrockford32.StdEncoding, bytes.NewReader([]byte(tc.encoded)))
			for {
				if _, err := io.CopyN(&buf, dec, bs); io.EOF == err {
					break
				} else if nil != err {
					t.Errorf("Failed to decode: %v", err)
				}
			}
			if g, e := buf.String(), string(tc.decoded); g != e {
				t.Errorf("Decode %x wrong result: %q != %q", tc.decoded, g, e)
			}
		}
	}
}
