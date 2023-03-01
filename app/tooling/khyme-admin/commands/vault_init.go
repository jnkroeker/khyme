package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jnkroeker/khyme/foundation/vault"
)

const credentialsFileName = "/vault/credentials.json"

func VaultInit(cfg vault.Config) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	vaultSrv, err := vault.New(vault.Config{
		Address:   cfg.Address,
		MountPath: cfg.MountPath,
	})
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	// =========================================================================

	log.Println("Check system is already initialized")

	// NOTE: This assumes the Vault POD is never restarted.

	initResponse, err := checkIfCredFileExists()
	if err == nil {
		log.Printf("rootToken: %s", initResponse.RootToken)
		vaultSrv.SetToken(initResponse.RootToken)

		if err = vaultSrv.CheckToken(ctx, cfg.Token); err == nil {
			log.Printf("application token %q exists", cfg.Token)
			return nil
		}
	}

	// =========================================================================

	log.Println("Initializing vault")

	initResponse, err = vaultSrv.SystemInit(ctx, 1, 1)
	if err != nil {
		if errors.Is(err, vault.ErrAlreadyInitialized) {
			log.Println("vault initialized: %w", err)
			return fmt.Errorf("vault is already initialized but we don't have the credentials file")
		}
		log.Println("unable to initialize: %w", err)
		return fmt.Errorf("unable to initialize Vault instance: %w", err)
	}

	b, err := json.Marshal(initResponse)
	if err != nil {
		log.Println("marshaling error: %w", err)
		return errors.New("unable to marshal")
	}

	if err := os.WriteFile(credentialsFileName, b, 0644); err != nil {
		log.Println("unable to write credentials to file: %w", err)
		return fmt.Errorf("unable to write %s file: %w", credentialsFileName, err)
	}

	log.Printf("rootToken: %s", initResponse.RootToken)
	vaultSrv.SetToken(initResponse.RootToken)

	log.Println("checking token exists")
	err = vaultSrv.CheckToken(ctx, cfg.Token)
	if err == nil {
		log.Printf("token already exists: %s", cfg.Token)
		return nil
	}

	// =========================================================================

	log.Println("Unsealing vault")

	err = vaultSrv.Unseal(ctx, initResponse.KeysB64[0])
	if err != nil {
		if errors.Is(err, vault.ErrBadRequest) {
			return fmt.Errorf("vault is not initialized. Check for old credentials file: %s", credentialsFileName)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	// =========================================================================

	log.Println("Mounting path in vault")

	vaultSrv.SetToken(initResponse.RootToken)
	if err := vaultSrv.Mount(ctx); err != nil {
		if errors.Is(err, vault.ErrPathInUse) {
			return fmt.Errorf("unable to mount path: %w", err)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	return nil
}

func checkIfCredFileExists() (vault.SystemInitResponse, error) {
	if _, err := os.Stat(credentialsFileName); err != nil {
		return vault.SystemInitResponse{}, err
	}

	data, err := os.ReadFile(credentialsFileName)
	if err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("reading %s file: %s", credentialsFileName, err)
	}

	var response vault.SystemInitResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("unmarshalling json: %s", err)
	}

	return response, nil
}
