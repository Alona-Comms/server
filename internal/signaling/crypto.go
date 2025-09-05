package signaling

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateEd25519KeyPair() (publicKey string, privateKey string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	publicKey = base64.StdEncoding.EncodeToString(pub)
	privateKey = base64.StdEncoding.EncodeToString(priv)

	return publicKey, privateKey, nil
}

func ValidatePublicKey(publicKeyB64 string) error {
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: got %d, expected %d",
			len(keyBytes), ed25519.PublicKeySize)
	}

	return nil
}

func ParseEd25519PublicKey(publicKeyB64 string) (ed25519.PublicKey, error) {
	if err := ValidatePublicKey(publicKeyB64); err != nil {
		return nil, err
	}

	keyBytes, _ := base64.StdEncoding.DecodeString(publicKeyB64)
	return ed25519.PublicKey(keyBytes), nil
}

func ParseEd25519PrivateKey(privateKeyB64 string) (ed25519.PrivateKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: got %d, expected %d",
			len(keyBytes), ed25519.PrivateKeySize)
	}

	return ed25519.PrivateKey(keyBytes), nil
}
