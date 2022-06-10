package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"log"
	"math/big"
)

func PubBytes(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

func CreateKeys() (string, string, string, error) {
	privateKey, _, _ := CreatePrivateKeyP256()
	publicKey := CreatePublicKey(privateKey)
	authSecret, err := CreateAuthSecret()
	if err != nil {
		return "", "", "", nil
	}

	privateKeyEncoded := base64.RawURLEncoding.EncodeToString(privateKey)
	publicKeyEncoded := base64.RawURLEncoding.EncodeToString(publicKey)
	authSecretEncoded := base64.RawURLEncoding.EncodeToString(authSecret)

	return privateKeyEncoded, publicKeyEncoded, authSecretEncoded, nil
}

func CreateAuthSecret() ([]byte, error) {
	authSecret := make([]byte, 16)
	_, err := rand.Read(authSecret)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return authSecret, nil
}

func CreatePrivateKeyP256() ([]byte, *big.Int, *big.Int) {
	privateKey, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil
	}
	return privateKey, x, y
}

func CreatePublicKey(privateKey []byte) []byte {
	var pri ecdsa.PrivateKey
	pri.D = new(big.Int).SetBytes(privateKey)
	pri.PublicKey.Curve = elliptic.P256()
	pri.PublicKey.X, pri.PublicKey.Y = pri.PublicKey.Curve.ScalarBaseMult(pri.D.Bytes())
	return PubBytes(&pri.PublicKey)
}
