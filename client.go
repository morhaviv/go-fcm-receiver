package go_fcm_receiver

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"

	pb "go-fcm-receiver/proto"
	"log"
	"net"
)

func (f *FCMClient) StartListening() {

	loginRequest := CreateLoginRequest(&f.androidId, &f.securityToken, "", f.PersistentIds)
	f.connect(loginRequest)
}

func (f *FCMClient) connect(loginRequest *pb.LoginRequest) {
	test1, err := json.Marshal(loginRequest)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(string(test1))

	loginRequestData, err := proto.Marshal(loginRequest)
	if err != nil {
		log.Print(err)
		return
	}

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
	fmt.Println(string(loginRequestData))

	n, err := conn.Write(loginRequestData)
	if err != nil {
		log.Println(n, err)
		return
	}

	fmt.Println("Getting response")
	i := 0
	var buf []byte
	for {
		i++
		buf = make([]byte, 50)
		n, err = conn.Read(buf)
		if err != nil {
			log.Println(n, err)
			return
		}
		fmt.Println(buf)
		fmt.Println(string(buf[:n]))
		fmt.Println(i)
	}

}
