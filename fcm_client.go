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
	SenderId      int64
	HttpClient    *http.Client
	AppId         string
	GcmToken      string
	FcmToken      string
	Socket        *tls.Conn
	androidId     uint64
	securityToken uint64
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	authSecret    []byte
	PersistentIds []string
	socket        *FCMSocketHandler
}

func (f *FCMClient) CreateAppId() string {
	f.AppId = fmt.Sprintf(generic.AppIdBase, uuid.New().String())
	return f.AppId
}

func (f *FCMClient) StartListening() {
	loginRequest := fcm_protos.CreateLoginRequestRaw(&f.androidId, &f.securityToken, "", f.PersistentIds)
	f.connect(loginRequest)
}

func (f *FCMClient) connect(loginRequest []byte) {
	tlsConfig := &tls.Config{
		GetConfigForClient: func(c *tls.ClientHelloInfo) (*tls.Config, error) {
			c.Conn.(*net.TCPConn).SetKeepAlive(true) // Todo: Check if successful
			//c.Conn.(*net.TCPConn).
			return nil, nil
		},
	}

	socket, err := tls.Dial("tcp", generic.FcmSocketAddress, tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}

	fcmSocket := FCMSocketHandler{
		Socket:    socket,
		OnMessage: f.onMessage,
	}
	f.socket = &fcmSocket
	fcmSocket.Init()

	fmt.Println("Token ", f.FcmToken)

	f.startLoginHandshake(loginRequest)
	fcmSocket.StartSocketHandler()
}

func (f *FCMClient) startLoginHandshake(loginRequest []byte) {
	n, err := f.socket.Socket.Write(loginRequest)
	if err != nil {
		log.Println(n, err)
		return
	}
}

func (f *FCMClient) onMessage(messageTag int, messageObject interface{}) {
	if messageTag == generic.KDataMessageStanzaTag {
		dataMessage, ok := messageObject.(*fcm_protos.DataMessageStanza)
		if ok {
			f.onDataMessage(dataMessage)
		} else {
			err := errors.New("error casting message to DataMessageStanza")
			log.Println(err)
		}
	}
}

func (f *FCMClient) onDataMessage(message *fcm_protos.DataMessageStanza) error {
	var err error
	var cryptoKey []byte
	var encryption []byte
	for _, data := range message.AppData {
		if *data.Key == "crypto-key" {
			fmt.Println("cryptoKey", *data.Value, len(*data.Value), []byte(*data.Value))
			cryptoKey, err = base64.URLEncoding.DecodeString((*data.Value)[3:])
			if err != nil {
				log.Println(err)
				return err
			}
		}
		if *data.Key == "encryption" {
			fmt.Println("encryption", *data.Value, len(*data.Value), []byte(*data.Value))
			encryption, err = base64.URLEncoding.DecodeString((*data.Value)[5:])
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	rawData := message.RawData

	//privateKey := ecdsa.PrivateKey{}
	//privateKey.D = new(big.Int).SetBytes([]byte(f.privateKey))
	//privateKey.PublicKey.Curve = elliptic.P256()
	//privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.PublicKey.Curve.ScalarBaseMult(privateKey.D.Bytes())
	err = DecryptMessage(cryptoKey, encryption, rawData, f.authSecret, f.privateKey)
	if err != nil {
		return err
	}
	return nil
}
