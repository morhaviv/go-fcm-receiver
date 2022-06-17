package go_fcm_receiver

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"go-fcm-receiver/proto"
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
	privateKey    string
	publicKey     string
	authSecret    string
	PersistentIds []string
	socket        *FCMSocketHandler
}

func (f *FCMClient) CreateAppId() string {
	f.AppId = fmt.Sprintf(AppIdBase, uuid.New().String())
	return f.AppId
}

func (f *FCMClient) StartListening() {
	loginRequest := proto.CreateLoginRequestRaw(&f.androidId, &f.securityToken, "", f.PersistentIds)
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

	socket, err := tls.Dial("tcp", FcmSocketAddress, tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}

	fcmSocket := FCMSocketHandler{Socket: socket}
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
