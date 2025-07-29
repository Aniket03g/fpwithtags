package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v56/github"
	"github.com/yourusername/github-app-boilerplate/config"
)

type Client struct {
	appClient     *github.Client
	config       *config.Config
	privateKey   *rsa.PrivateKey
}

func NewClient(cfg *config.Config) (*Client, error) {
	// Parse the private key
	block, _ := pem.Decode([]byte(cfg.GitHubPrivateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse private key: invalid PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	transport := ghinstallation.NewAppsTransport(
		ghinstallation.WithPrivateKey(privateKey),
		cfg.GitHubAppID,
	)

	return &Client{
		appClient:     github.NewClient(transport.Client()),
		config:       cfg,
		privateKey:   privateKey,
	}, nil
}

// GetInstallationToken returns an installation token for the given installation ID
func (c *Client) GetInstallationToken(installationID int64) (string, error) {
	tr := ghinstallation.NewFromAppsTransport(
		ghinstallation.NewAppsTransport(
			ghinstallation.WithPrivateKey(c.privateKey),
			c.config.GitHubAppID,
		),
		installationID,
	)

	token, err := tr.Token(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get installation token: %v", err)
	}

	log.Printf("Successfully generated installation token for installation ID %d", installationID)
	return token, nil
}

// NewInstallationClient creates a new GitHub client authenticated as an installation
func (c *Client) NewInstallationClient(installationID int64) (*github.Client, error) {
	tr := ghinstallation.NewFromAppsTransport(
		ghinstallation.NewAppsTransport(
			ghconfiguration.WithPrivateKey(c.privateKey),
			c.config.GitHubAppID,
		),
		installationID,
	)

	return github.NewClient(tr.Client()), nil
}

// VerifyWebhookSignature verifies the signature of a webhook payload
func (c *Client) VerifyWebhookSignature(payload []byte, signature string) error {
	return ghinstallation.VerifyPayload(payload, signature, []byte(c.config.GitHubWebhookSecret))
}
