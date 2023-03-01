package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// This provides a default client configuration, but it's recommended
// this is replaced by the user with application specific settings using
// the WithClient function at the time a GraphQL is constructed.
var defaultClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// Config represents the mandatory settings needed to work with Vault.
type Config struct {
	Address   string
	MountPath string
	Token     string
	Client    *http.Client
}

// Vault provides support to access Hashicorp's Vault product for keys.
type Vault struct {
	address   string
	token     string
	mountPath string
	client    *http.Client
	mu        sync.RWMutex
	store     map[string]string
}

// New constructs a vault for use.
func New(cfg Config) (*Vault, error) {
	if cfg.Client == nil {
		cfg.Client = &defaultClient
	}

	return &Vault{
		address:   cfg.Address,
		token:     cfg.Token,
		mountPath: cfg.MountPath,
		client:    cfg.Client,
		store:     make(map[string]string),
	}, nil
}

// SetToken allows the user to change out the token to use on calls.
func (v *Vault) SetToken(token string) {
	v.token = token
}

// =============================================================================

// Error variables for this set of API calls.
var (
	ErrAlreadyInitialized = errors.New("already initalized")
	ErrBadRequest         = errors.New("bad request")
	ErrPathInUse          = errors.New("path in use")
)

// SystemInitResponse represents the response from a system init call.
type SystemInitResponse struct {
	KeysB64   []string `json:"keys_base64"`
	RootToken string   `json:"root_token"`
}

// SystemInit provides support to initialize a vault system for use.
func (v *Vault) SystemInit(ctx context.Context, shares int, threshold int) (SystemInitResponse, error) {
	url := fmt.Sprintf("%s/v1/sys/init", v.address)

	cfg := struct {
		Shares    int `json:"secret_shares"`
		Threshold int `json:"secret_threshold"`
	}{
		Shares:    shares,
		Threshold: threshold,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return SystemInitResponse{}, fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &b)
	if err != nil {
		return SystemInitResponse{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "Vault is already initialized") {
			return SystemInitResponse{}, ErrAlreadyInitialized
		}
		return SystemInitResponse{}, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SystemInitResponse{}, fmt.Errorf("status code: %s", resp.Status)
	}

	var response SystemInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return SystemInitResponse{}, fmt.Errorf("json decode: %w", err)
	}

	return response, nil
}

// Unseal does what the unseal command does.
func (v *Vault) Unseal(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/v1/sys/unseal", v.address)

	cfg := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &b)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return ErrBadRequest
		default:
			return fmt.Errorf("status code: %s", resp.Status)
		}
	}

	return nil
}

// Mount accepts a mount point and mounts vault to that point.
func (v *Vault) Mount(ctx context.Context) error {
	mounts, err := v.listMounts(ctx)
	if err != nil {
		return fmt.Errorf("error getting mount list: %w", err)
	}

	// Mount already exists so we'll do nothing.
	if _, ok := mounts[v.mountPath]; ok {
		return nil
	}

	url := fmt.Sprintf("%s/v1/sys/mounts/%s", v.address, v.mountPath)

	cfg := struct {
		Type    string            `json:"type"`
		Options map[string]string `json:"options"`
	}{
		Type:    "kv-v2",
		Options: map[string]string{"version": "2"},
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cfg); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		if !strings.Contains(err.Error(), "path is already in use at") {
			return ErrPathInUse
		}
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("status code: %s", resp.Status)
	}

	return nil
}

// CheckToken validates the specified token exists.
func (v *Vault) CheckToken(ctx context.Context, token string) error {
	url := fmt.Sprintf("%s/v1/auth/token/lookup", v.address)

	t := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(t); err != nil {
		return fmt.Errorf("encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return fmt.Errorf("lookup request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token doesn't exist: %s", token)
	}

	return nil
}

// =============================================================================

// listMounts returns the set of mount points that exist.
func (v *Vault) listMounts(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/sys/mounts", v.address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	return response, nil
}
