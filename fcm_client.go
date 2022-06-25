package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"net/http"
	"time"
)

// FCMClient structure
type FCMClient struct {
	SenderId          int64
	HttpClient        http.Client
	AppId             string
	GcmToken          string
	FcmToken          string
	AndroidId         uint64
	SecurityToken     uint64
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	authSecret        []byte
	PersistentIds     []string
	HeartbeatInterval time.Duration
	socket            FCMSocketHandler
	OnDataMessage     func(message []byte)
}

func (f *FCMClient) LoadKeys(privateKeyBase64 string, authSecretBase64 string) error {
	privateKeyString, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to base64 decode the private key: %s", err.Error()))
		return err
	}
	privateKey, err := DecodePrivateKey(privateKeyString)
	if err != nil {
		return err
	}
	authSecretKeyString, err := base64.StdEncoding.DecodeString(authSecretBase64)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to base64 decode the auth secret key: %s", err.Error()))
		return err
	}
	f.privateKey = privateKey
	f.publicKey = &privateKey.PublicKey
	f.authSecret = authSecretKeyString

	return nil
}

func (f *FCMClient) CreateAppId() string {
	f.AppId = fmt.Sprintf(AppIdBase, uuid.New().String())
	return f.AppId
}

func (f *FCMClient) StartListening() error {
	return f.connect()
}

func (f *FCMClient) connect() error {
	tlsConfig := &tls.Config{
		GetConfigForClient: func(c *tls.ClientHelloInfo) (*tls.Config, error) {
			err := c.Conn.(*net.TCPConn).SetKeepAlive(true)
			if err != nil {
				err = errors.New(fmt.Sprintf("failed to enable a keep-alive on this OS: %s", err.Error()))
				return nil, err
			}
			return nil, nil
		},
	}

	socket, err := tls.Dial("tcp", FcmSocketAddress, tlsConfig)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to connect to the FCM server: %s", err.Error()))
		return err
	}

	f.socket.IsAlive = true
	f.socket.Socket = socket
	f.socket.OnMessage = f.onMessage
	f.socket.Init()

	loginRequest, err := CreateLoginRequestRaw(&f.AndroidId, &f.SecurityToken, f.PersistentIds)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to create a login request packet: %s", err.Error()))
		return err
	}

	err = f.startLoginHandshake(loginRequest)
	if err != nil {
		return err
	}

	return f.socket.StartSocketHandler()
}

func (f *FCMClient) startLoginHandshake(loginRequest []byte) error {
	_, err := f.socket.Socket.Write(loginRequest)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to send a handshake to the FCM server: %s", err.Error()))
		f.socket.close(err)
		return err
	}
	return nil
}

func (f *FCMClient) onMessage(messageTag int, messageObject interface{}) error {
	if messageTag == KLoginResponseTag {
		f.PersistentIds = nil
	} else if messageTag == KHeartbeatPingTag {
		err := f.socket.SendHeartbeatPing()
		if err != nil {
			return err
		}
	} else if messageTag == KDataMessageStanzaTag {
		dataMessage, ok := messageObject.(*DataMessageStanza)
		if ok {
			err := f.onDataMessage(dataMessage)
			if err != nil {
				return err
			}
		} else {
			err := errors.New("the received message is corrupted and couldn't be casted as DataMessageStanza")
			return err
		}
	}
	return nil
}

func (f *FCMClient) onDataMessage(message *DataMessageStanza) error {
	if StringsSliceContains(f.PersistentIds, *message.PersistentId) {
		return nil
	}
	var err error
	var cryptoKey []byte
	var encryption []byte
	for _, data := range message.AppData {
		if *data.Key == "crypto-key" {
			cryptoKey, err = base64.URLEncoding.DecodeString((*data.Value)[3:])
			if err != nil {
				err = errors.New(fmt.Sprintf("failed to base64 decode the crypto-key received from the server: %s", err.Error()))
				return err
			}
		}
		if *data.Key == "encryption" {
			encryption, err = base64.URLEncoding.DecodeString((*data.Value)[5:])
			if err != nil {
				err = errors.New(fmt.Sprintf("failed to base64 decode the encryption received from the server: %s", err.Error()))
				return err
			}
		}
	}
	rawData := message.RawData
	decryptedMessage, err := DecryptMessage(cryptoKey, encryption, rawData, f.authSecret, f.privateKey)
	if err != nil {
		return err
	}
	f.PersistentIds = append(f.PersistentIds, *message.PersistentId)
	go f.OnDataMessage(decryptedMessage)
	return nil
}
