// Copyright 2024 Brisk, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This shows an example of how to generate a SSH RSA Private/Public key pair and save it locally

package shared

import (
	. "brisk-supervisor/shared/logger"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"os"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func CreateKey(ctx context.Context, name string) (string, string, error) {

	//need to make this unique per project - at the moment we'll just overwrite
	//	savePrivateFileTo := "/tmp/.ssh/id_rsa.priv"
	//savePublicFileTo := "/home/brisk/.ssh/authorized_keys"
	savePublicFileTo := "/keys/authorized_keys"
	authorizedKeyPath := "/var/lib/bastion/authorized_keys"
	bitSize := 4096

	privateKey, err := generatePrivateKey(ctx, bitSize)
	if err != nil {
		Logger(ctx).Error(err.Error())
		return "", "", err
	}

	publicKeyBytes, err := generatePublicKey(ctx, &privateKey.PublicKey)
	if err != nil {
		Logger(ctx).Error(err.Error())
		return "", "", err
	}

	privateKeyBytes := encodePrivateKeyToPEM(ctx, privateKey)

	// this isn't required but lets do it for now for debug
	// err = writeKeyToFile(privateKeyBytes, savePrivateFileTo)
	// if err != nil {
	// 	Logger(ctx).Fatal(err)
	// }
	rrsyncCommand := os.Getenv("RRSYNC_COMMAND")

	rrsyncAndKey := rrsyncCommand + " " + string(publicKeyBytes)

	//writing this to our own ssh
	err = WriteKeyToFile(ctx, []byte(rrsyncAndKey), savePublicFileTo)
	if err != nil {
		//this is so we can log into the super
		Logger(ctx).Errorf("Error writing key to file: %s", err.Error())
		Logger(ctx).Error("What the fuck is happening here")
		return "", "", err
	}
	Logger(ctx).Debug("About to add to bastion file")
	bastionRrsyncAndKey := os.Getenv("BASTION_RRSYNC_COMMAND") + " " + string(publicKeyBytes)
	err = appendKeyToFile(ctx, []byte(bastionRrsyncAndKey), authorizedKeyPath)
	Logger(ctx).Debug("Appended to bastion key file")
	// this is so we can get past the bastion
	if err != nil {
		Logger(ctx).Errorf("Error appending key to file: %s", err.Error())
		if viper.GetBool("USE_DOCKER_COMPOSE") {
			err = nil //clear this error
			Logger(ctx).Warn("We are using docker compose so we don't error out when we can't write to the bastion file - this is not an issue unless you are sshing through a bastion host")
		} else {
			return "", "", err
		}
	}
	return string(privateKeyBytes), string(publicKeyBytes), nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(ctx context.Context, bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	Logger(ctx).Debug("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(ctx context.Context, privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(ctx context.Context, privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	Logger(ctx).Debug("Public key generated")
	return pubKeyBytes, nil
}
func ReadKeyFromFile(ctx context.Context, keyFile string) ([]byte, error) {
	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func WriteKeyToFile(ctx context.Context, keyBytes []byte, saveFileTo string) error {

	f, err := os.OpenFile(saveFileTo, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(string(keyBytes)); err != nil {
		return err
	}

	Logger(ctx).Debugf("Key saved to: %s", saveFileTo)
	return nil
}

// appendKeyToFile appends keys to a file - used for authorized hosts on bastion
func appendKeyToFile(ctx context.Context, keyBytes []byte, saveFileTo string) error {
	// we are appending to a NFS version 4 so we should have locking by default

	f, err := os.OpenFile(saveFileTo,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Logger(ctx).Debug(err)
	}
	defer f.Close()
	if _, err := f.WriteString(string(keyBytes)); err != nil {
		Logger(ctx).Error(err)
		return err
	}
	Logger(ctx).Debugf("Saved key to %s", saveFileTo)
	return nil
}
