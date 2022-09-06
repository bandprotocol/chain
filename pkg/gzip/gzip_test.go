package gzip_test

import (
	"bytes"
	gz "compress/gzip"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/gzip"
)

func TestUncompress(t *testing.T) {
	file1 := []byte("file")
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(file1)
	zw.Close()
	gzipFile := buf.Bytes()

	accFile, err := gzip.Uncompress(gzipFile, 10)
	require.NoError(t, err)
	require.Equal(t, file1, accFile)

	accFile, err = gzip.Uncompress(gzipFile, 2)
	require.Error(t, err)

	_, err = gzip.Uncompress(file1, 999)
	require.Error(t, err)
}

func TestCompress(t *testing.T) {
	bytes := []byte("file")
	compressedBytes, err := gzip.Compress(bytes)
	require.Equal(
		t,
		[]byte{
			0x1f,
			0x8b,
			0x8,
			0x0,
			0x0,
			0x0,
			0x0,
			0x0,
			0x0,
			0xff,
			0x4a,
			0xcb,
			0xcc,
			0x49,
			0x5,
			0x4,
			0x0,
			0x0,
			0xff,
			0xff,
			0x10,
			0x36,
			0x9f,
			0x8c,
			0x4,
			0x0,
			0x0,
			0x0,
		},
		compressedBytes,
	)
	require.NoError(t, err)
}

func TestIsGzip(t *testing.T) {
	file1 := []byte("file")
	var buf bytes.Buffer
	zw := gz.NewWriter(&buf)
	zw.Write(file1)
	zw.Close()
	gzipFile := buf.Bytes()
	require.True(t, gzip.IsGzipped(gzipFile))
	require.False(t, gzip.IsGzipped(file1))
}
