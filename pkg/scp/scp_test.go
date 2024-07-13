package scp

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gliderlabs/ssh"
)

func TestScp(t *testing.T) {
	addr := ":2200"
	user := "donut"
	privateKey := newPrivateKey(t)
	signer, err := newSigner(privateKey)
	require.NoError(t, err)
	done := make(chan struct{})

	wd, err := os.Getwd()
	require.NoError(t, err)
	src, dst := filepath.Join(wd, "scp_test.go"), filepath.Join(wd, "something_else.go")

	srv := &ssh.Server{Addr: addr, HostSigners: []ssh.Signer{signer}, Handler: func(s ssh.Session) {
		cmd := s.RawCommand()
		u := s.User()
		assert.Equal(t, u, user)
		expectedCmd := fmt.Sprintf("scp -qt \"%s\"", dst)
		assert.Equal(t, cmd, expectedCmd)

		// Read SCP protocol initiation
		buf := make([]byte, 1)
		_, err := s.Read(buf)
		require.NoError(t, err)
		require.Equal(t, byte('C'), buf[0])

		// Read file header
		header := make([]byte, 512)
		n, err := s.Read(header)
		require.NoError(t, err)
		header = header[:n]

		// Send OK response
		_, err = s.Write([]byte{0})
		require.NoError(t, err)

		// retrieve file size
		headerParts := strings.Split(string(header), " ")
		require.Len(t, headerParts, 3)
		fileSize, err := strconv.ParseInt(headerParts[1], 10, 64)
		require.NoError(t, err)

		// Read file content
		fileBuf := bytes.NewBuffer(nil)
		_, err = io.CopyN(fileBuf, s, fileSize)
		require.NoError(t, err)
		srcFile, err := os.Open(src)
		defer srcFile.Close()
		require.NoError(t, err)
		srcFileContent, err := io.ReadAll(srcFile)
		require.NoError(t, err)
		assert.Equal(t, srcFileContent, fileBuf.Bytes())

		// Send OK response after file transfer
		_, err = s.Write([]byte{0})
		require.NoError(t, err)

		err = s.CloseWrite()
		require.NoError(t, err)

		close(done)
	}}
	ctx := context.Background()
	defer srv.Shutdown(ctx)

	go func() {
		err := srv.ListenAndServe()
		require.ErrorIs(t, err, ssh.ErrServerClosed)
	}()
	// wait for server to start
	time.Sleep(time.Millisecond * 200)

	scpClient, err := New(addr, user, privateKey)
	require.NoError(t, err)

	err = scpClient.Scp(ctx, src, dst)
	assert.NoError(t, err)

	select {
	case <-done:
		return
	case <-time.After(time.Millisecond * 200):
		t.Error("ssh server hasn't been called")
	}
}

func newPrivateKey(t *testing.T) string {
	size := 4096
	key, err := rsa.GenerateKey(rand.Reader, size)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	buf := bytes.NewBuffer(nil)
	base64.NewEncoder(base64.RawStdEncoding, buf).Write(keyPEM)
	// base64.RawStdEncoding.Encode(buf, keyPEM)
	// return string(buf)
	return buf.String()
}
