package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/morhaviv/go-fcm-receiver/fcm_protos"
	"github.com/morhaviv/go-fcm-receiver/generic"
	"log"
	"net"
	"net/http"
)

// FCMClient structure
type FCMClient struct {
	IsAlive       bool
	SenderId      int64
	HttpClient    *http.Client
	AppId         string
	GcmToken      string
	FcmToken      string
	AndroidId     uint64
	SecurityToken uint64
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	authSecret    []byte
	PersistentIds []string
	Socket        FCMSocketHandler
	OnDataMessage func(message []byte)
}

func (f *FCMClient) LoadKeys(privateKeyBase64 string, authSecretBase64 string) error {
	// Todo: change variable names to comments...
	privateKeyString, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		log.Println(err)
		return err
	}
	privateKey, err := generic.DecodePrivateKey(privateKeyString)
	if err != nil {
		return err
	}
	authSecretKeyString, err := base64.StdEncoding.DecodeString(authSecretBase64)
	if err != nil {
		log.Println(err)
		return err
	}
	f.privateKey = privateKey
	f.publicKey = &privateKey.PublicKey
	f.authSecret = authSecretKeyString

	return nil
}

func (f *FCMClient) CreateAppId() string {
	f.AppId = fmt.Sprintf(generic.AppIdBase, uuid.New().String())
	return f.AppId
}

func (f *FCMClient) StartListening() {
	f.connect()
}

func (f *FCMClient) connect() {
	tlsConfig := &tls.Config{
		GetConfigForClient: func(c *tls.ClientHelloInfo) (*tls.Config, error) {
			err := c.Conn.(*net.TCPConn).SetKeepAlive(true)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	}

	socket, err := tls.Dial("tcp", generic.FcmSocketAddress, tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}
	f.IsAlive = true

	f.Socket.Socket = socket
	f.Socket.OnMessage = f.onMessage
	f.Socket.OnClose = func() {
		f.IsAlive = false
	}
	f.Socket.Init()

	fmt.Println("FCM Token:", f.FcmToken)
	loginRequest := fcm_protos.CreateLoginRequestRaw(&f.AndroidId, &f.SecurityToken, "", f.PersistentIds)
	err = f.startLoginHandshake(loginRequest)
	if err != nil {
		return
	}
	f.Socket.StartSocketHandler()
}

func (f *FCMClient) startLoginHandshake(loginRequest []byte) error {
	n, err := f.Socket.Socket.Write(loginRequest)
	if err != nil {
		log.Println(n, err)
		f.Socket.close()
		f.IsAlive = false
		return err
	}
	return nil
}

func (f *FCMClient) onMessage(messageTag int, messageObject interface{}) error {
	if messageTag == generic.KLoginResponseTag {
		f.PersistentIds = nil
	} else if messageTag == generic.KHeartbeatPingTag {
		err := f.Socket.SendHeartbeatPing()
		if err != nil {
			return err
		}
	} else if messageTag == generic.KDataMessageStanzaTag {
		dataMessage, ok := messageObject.(*fcm_protos.DataMessageStanza)
		if ok {
			err := f.onDataMessage(dataMessage)
			if err != nil {
				return err
			}
		} else {
			err := errors.New("error casting message to DataMessageStanza")
			log.Println(err)
			return err
		}
	}
	return nil
}

func (f *FCMClient) onDataMessage(message *fcm_protos.DataMessageStanza) error {
	if generic.StringsSliceContains(f.PersistentIds, *message.PersistentId) {
		return nil
	}
	var err error
	var cryptoKey []byte
	var encryption []byte
	for _, data := range message.AppData {
		if *data.Key == "crypto-key" {
			cryptoKey, err = base64.URLEncoding.DecodeString((*data.Value)[3:])
			if err != nil {
				log.Println(err)
				return err
			}
		}
		if *data.Key == "encryption" {
			encryption, err = base64.URLEncoding.DecodeString((*data.Value)[5:])
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	rawData := message.RawData
	decryptedMessage, err := generic.DecryptMessage(cryptoKey, encryption, rawData, f.authSecret, f.privateKey)
	if err != nil {
		return err
	}
	f.PersistentIds = append(f.PersistentIds, *message.PersistentId)
	go f.OnDataMessage(decryptedMessage)
	return nil
}
