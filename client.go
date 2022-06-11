package go_fcm_receiver

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func (f *FCMClient) StartListening() {
	loginRequest := CreateLoginRequestRaw(&f.androidId, &f.securityToken, "", f.PersistentIds)
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
