package zrockford32_test

import (
	"encoding/base32"
	"encoding/hex"
	"testing"

	"github.com/checksum0/go-zrockford32"
)

func BenchmarkEncodeBytes(b *testing.B) {
	decoded := []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}
	dst := make([]byte, zrockford32.StdEncoding.EncodedLen(len(decoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := zrockford32.StdEncoding.Encode(dst, decoded)
		_ = dst[:n]
	}
}

func BenchmarkEncodeBase32(b *testing.B) {
	decoded := []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}
	dst := make([]byte, base32.StdEncoding.EncodedLen(len(decoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		base32.StdEncoding.Encode(dst, decoded)
		_ = dst
	}
}

func BenchmarkEncodeHex(b *testing.B) {
	decoded := []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}
	dst := make([]byte, hex.EncodedLen(len(decoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := hex.Encode(dst, decoded)
		_ = dst[:n]
	}
}

func BenchmarkDecodeBytes(b *testing.B) {
	encoded := []byte("AB3SR12X8FHFNVZAE075FKN3A7XH8VDK6JS22K0")
	dst := make([]byte, zrockford32.StdEncoding.DecodedLen(len(encoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n, err := zrockford32.StdEncoding.Decode(dst, encoded)
		if err != nil {
			b.Fatalf("decode error: %v", err)
		}
		_ = dst[:n]
	}
}

func BenchmarkDecodeBase32(b *testing.B) {
	encoded := []byte("YBZWESVPHF4FCTXYIQ53FKCZY5P4HTDK6JWVVKQ=")
	dst := make([]byte, base32.StdEncoding.DecodedLen(len(encoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n, err := base32.StdEncoding.Decode(dst, encoded)
		if err != nil {
			b.Fatalf("decode error: %v", err)
		}
		_ = dst[:n]
	}
}

func BenchmarkDecodeHex(b *testing.B) {
	encoded := []byte("c073624aaf3978514ef8443bb2a859c75fc3cc6af26d5aaa")
	dst := make([]byte, hex.DecodedLen(len(encoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n, err := hex.Decode(dst, encoded)
		if err != nil {
			b.Fatalf("decode error: %v", err)
		}
		_ = dst[:n]
	}
}
