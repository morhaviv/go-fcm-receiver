package go_fcm_receiver

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// FCMClient structure
type FCMClient struct {
	VapidKey              string
	ApiKey                string
	AppId                 string
	ProjectID             string
	HttpClient            http.Client
	GcmToken              string
	FcmToken              string
	AndroidId             uint64
	SecurityToken         uint64
	privateKey            *ecdsa.PrivateKey
	publicKey             *ecdsa.PublicKey
	authSecret            []byte
	PersistentIds         []string
	persistentMutex       sync.Mutex
	HeartbeatInterval     time.Duration
	socket                FCMSocketHandler
	OnDataMessage         func(message []byte)
	OnRawMessage          func(message []*AppData)
	AndroidApp            *AndroidFCM
	InstallationAuthToken *string
}

// AndroidFCM App
type AndroidFCM struct {
	GcmSenderId        string
	GmsAppId           string
	AndroidPackage     string
	AndroidPackageCert string
}

func (f *FCMClient) RemovePersistentId(id string) {
	f.persistentMutex.Lock()
	defer f.persistentMutex.Unlock()

	for i, persistentId := range f.PersistentIds {
		if persistentId == id {
			f.PersistentIds = append(f.PersistentIds[:i], f.PersistentIds[i+1:]...)
			return
		}
	}
}

func GenerateFirebaseFID() (string, error) {
	// A valid FID has exactly 22 base64 characters, which is 132 bits, or 16.5
	// bytes. Our implementation generates a 17 byte array instead.
	fid := make([]byte, 17)
	_, err := rand.Read(fid)
	if err != nil {
		return "", err
	}

	// Replace the first 4 random bits with the constant FID header of 0b0111.
	fid[0] = 0b01110000 + (fid[0] % 0b00010000)

	return base64.StdEncoding.EncodeToString(fid), nil
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

func (f *FCMClient) StartListening() error {
	if f.AndroidId == 0 || f.SecurityToken == 0 {
		err := errors.New("client's AndroidId and SecurityToken hasn't been set. use FCMClient.Register() to generate a new AndroidId and SecurityToken")
		return err
	}
	if f.privateKey == nil || f.authSecret == nil {
		err := errors.New("client's private key hasn't been set. use FCMClient.LoadKeys() or FCMClient.CreateNewKeys()")
		return err
	}
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
	f.socket.HeartbeatInterval = f.HeartbeatInterval
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
	if messageTag == KHeartbeatPingTag {
		err := f.socket.SendHeartbeatPing()
		if err != nil {
			return err
		}
	} else if messageTag == KCloseTag {
		err := errors.New("server returned close tag")
		f.socket.close(err)
		return err
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
	if StringsSliceContains(f.PersistentIds, message.GetPersistentId()) {
		return nil
	}
	var err error
	var cryptoKey []byte
	var encryption []byte
	isRaw := true

	for _, data := range message.AppData {
		if *data.Key == "crypto-key" {
			isRaw = false
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

	f.PersistentIds = append(f.PersistentIds, message.GetPersistentId())
	go func(persistentId string) {
		ttl := DefaultFcmMessageTtl
		if message.Ttl != nil {
			ttl = time.Duration(*message.Ttl) * time.Second
		}
		<-time.After(ttl)
		f.RemovePersistentId(persistentId)
	}(message.GetPersistentId())

	if !isRaw {
		rawData := message.RawData
		decryptedMessage, err := DecryptMessage(cryptoKey, encryption, rawData, f.authSecret, f.privateKey)
		if err != nil {
			return err
		}
		go f.OnDataMessage(decryptedMessage)
	} else {
		go f.OnRawMessage(message.AppData)
	}

	return nil
}

func (f *FCMClient) Close() {
	f.socket.close(errors.New("close was manually called"))
}
