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

	conn, err := tls.Dial("tcp", FcmSocketAddress, tlsConfig)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	fmt.Println("Writing login request")
	fmt.Println(loginRequest)

	fmt.Println(string(loginRequest))

	n, err := conn.Write(loginRequest)
	if err != nil {
		log.Println(n, err)
		return
	}

	fmt.Println("Getting response")

	buf := make([]byte, 1)
	n, err = conn.Read(buf)
	if err != nil {
		log.Println(n, err)
		return
	}
	fmt.Println("got response")
	fmt.Println(buf)

	buf2 := make([]byte, 64)
	n, err = conn.Read(buf2)
	if err != nil {
		log.Println(n, err)
		return
	}
	fmt.Println("got response")
	fmt.Println(buf2)
	fmt.Println(string(buf[:n]))
}
