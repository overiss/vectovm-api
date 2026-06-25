package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const dekSize = 32

var ErrInvalidKeySize = errors.New("invalid key size")

type SSMCredentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// Envelope encrypts data with per-user DEKs wrapped by a Vault master key (AWS KMS-style).
type Envelope struct {
	masterKey []byte
}

func NewEnvelope(masterKeyB64 string) (*Envelope, error) {
	key, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode master key: %w", err)
	}
	if len(key) != dekSize {
		return nil, ErrInvalidKeySize
	}
	return &Envelope{masterKey: key}, nil
}

func (e *Envelope) GenerateUserDEK() (dek, encryptedDEK []byte, err error) {
	dek = make([]byte, dekSize)
	if _, err = io.ReadFull(rand.Reader, dek); err != nil {
		return nil, nil, fmt.Errorf("generate dek: %w", err)
	}

	encryptedDEK, err = encryptAESGCM(e.masterKey, dek)
	if err != nil {
		return nil, nil, fmt.Errorf("wrap dek: %w", err)
	}
	return dek, encryptedDEK, nil
}

func (e *Envelope) UnwrapDEK(encryptedDEK []byte) ([]byte, error) {
	dek, err := decryptAESGCM(e.masterKey, encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("unwrap dek: %w", err)
	}
	if len(dek) != dekSize {
		return nil, ErrInvalidKeySize
	}
	return dek, nil
}

func (e *Envelope) EncryptCredentials(dek []byte, creds SSMCredentials) ([]byte, error) {
	payload, err := json.Marshal(creds)
	if err != nil {
		return nil, fmt.Errorf("marshal credentials: %w", err)
	}
	encrypted, err := encryptAESGCM(dek, payload)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}
	return encrypted, nil
}

func (e *Envelope) DecryptCredentials(dek []byte, encrypted []byte) (SSMCredentials, error) {
	payload, err := decryptAESGCM(dek, encrypted)
	if err != nil {
		return SSMCredentials{}, fmt.Errorf("decrypt credentials: %w", err)
	}

	var creds SSMCredentials
	if err := json.Unmarshal(payload, &creds); err != nil {
		return SSMCredentials{}, fmt.Errorf("unmarshal credentials: %w", err)
	}
	return creds, nil
}

func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, encrypted, nil)
}
