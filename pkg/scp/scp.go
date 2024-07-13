package scp

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	client scp.Client

	addr string
	mx   sync.Mutex
}

func New(addr, user, privateKey string) (*Client, error) {
	signer, err := newSigner(privateKey)
	if err != nil {
		return nil, err
	}
	conf := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(signer.PublicKey()),
	}

	return &Client{
		client: scp.NewClient(addr, &conf),
		addr:   addr,
	}, nil
}

func newSigner(key string) (ssh.Signer, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to deocde key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	return signer, nil
}

func (c *Client) Scp(ctx context.Context, src, dst string) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to ssh server, addr=%s: %w", c.addr, err)
	}
	defer c.client.Close()

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file, src=%s: %w", src, err)
	}

	if err := c.client.CopyFile(ctx, file, dst, "0644"); err != nil {
		return fmt.Errorf("failed to copy file, src=filepath, dst=path: %w", err)
	}

	return nil
}
