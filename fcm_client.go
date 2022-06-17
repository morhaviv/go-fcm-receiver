package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-fcm-receiver/fcm_protos"
	"go-fcm-receiver/generic"
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
	androidId     uint64
	securityToken uint64
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	authSecret    []byte
	PersistentIds []string
	Socket        *FCMSocketHandler
	OnDataMessage func(message []byte)
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

	fcmSocket := FCMSocketHandler{
		Socket:    socket,
		OnMessage: f.onMessage,
		OnClose: func() {
			f.IsAlive = false
		},
	}

	f.Socket = &fcmSocket
	fcmSocket.Init()

	fmt.Println("FCM Token: ", f.FcmToken)
	loginRequest := fcm_protos.CreateLoginRequestRaw(&f.androidId, &f.securityToken, "", f.PersistentIds)
	err = f.startLoginHandshake(loginRequest)
	if err != nil {
		return
	}
	fcmSocket.StartSocketHandler()
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
	if messageTag == generic.KDataMessageStanzaTag {
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
	decryptedMessage, err := DecryptMessage(cryptoKey, encryption, rawData, f.authSecret, f.privateKey)
	if err != nil {
		return err
	}
	f.PersistentIds = append(f.PersistentIds, *message.PersistentId)
	go f.OnDataMessage(decryptedMessage)
	return nil
}
