package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/xakep666/ecego"
)

func PubBytes(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

func CreateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, []byte, error) {
	privateKey, err := CreatePrivateKeyP256()
	if err != nil {
		return nil, nil, nil, err
	}
	authSecret, err := CreateAuthSecret()
	if err != nil {
		return nil, nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, authSecret, nil
}

func CreateAuthSecret() ([]byte, error) {
	authSecret := make([]byte, 16)
	_, err := rand.Read(authSecret)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to create a random auth secret: %s", err.Error()))
		return nil, err
	}
	return authSecret, nil
}

func CreatePrivateKeyP256() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to create a private key: %s", err.Error()))
		return nil, err
	}
	return privateKey, nil
}

func EncodePrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	derKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to encode the private key: %s", err.Error()))
		return nil, err
	}

	return derKey, nil
}

func DecodePrivateKey(key []byte) (*ecdsa.PrivateKey, error) {
	privateKey, err := x509.ParseECPrivateKey(key)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to decode the private key: %s", err.Error()))
		return nil, err
	}
	return privateKey, nil
}

func DecryptMessage(cryptoKey []byte, encryption []byte, rawData []byte, authSecret []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	engineOption := ecego.WithAuthSecret(authSecret)
	engine := ecego.NewEngine(ecego.SingleKey(privateKey), engineOption)
	params := ecego.OperationalParams{
		Version: ecego.AESGCM,
		Salt:    encryption,
		DH:      cryptoKey,
	}
	message, err := engine.Decrypt(rawData, []byte{}, params)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to decrypt the message from the server: %s", err.Error()))
		return nil, err
	}
	return message, nil
}
