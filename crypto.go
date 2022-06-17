package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/xakep666/ecego"
	"log"
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
	//publicKey := PubBytes(&privateKey.PublicKey)
	authSecret, err := CreateAuthSecret()
	if err != nil {
		return nil, nil, nil, nil
	}

	//privateKeyEncoded := base64.RawURLEncoding.EncodeToString(privateKey.D.Bytes())
	//publicKeyEncoded := base64.RawURLEncoding.EncodeToString(publicKey)
	//authSecretEncoded := base64.RawURLEncoding.EncodeToString(authSecret)

	return privateKey, &privateKey.PublicKey, authSecret, nil
}

func CreateAuthSecret() ([]byte, error) {
	authSecret := make([]byte, 16)
	_, err := rand.Read(authSecret)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return authSecret, nil
}

func CreatePrivateKeyP256() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Println(err)
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
		return nil, err
	}
	return message, nil
}
