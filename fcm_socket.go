package go_fcm_receiver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"go-fcm-receiver/fcm_protos"
	"go-fcm-receiver/generic"
	"log"
	"strconv"
	"sync"
	"time"
)

type FCMSocketHandler struct {
	Socket            *tls.Conn
	state             int
	data              []byte
	sizePacketSoFar   int
	messageTag        int
	messageSize       int
	handshakeComplete bool
	isWaitingForData  bool
	onDataMutex       sync.Mutex
	OnMessage         func(messageTag int, messageObject interface{})
}

func (f *FCMSocketHandler) StartSocketHandler() {
	go f.readData()
	time.Sleep(time.Minute * 5)
}

func (f *FCMSocketHandler) readData() {
	for {
		var buffer []byte
		buffer = make([]byte, 1)
		_, err := f.Socket.Read(buffer)
		if err != nil {
			f.close()
			log.Println(err)
			return
		}
		f.data = append(f.data, buffer...)
		go f.onData()
	}
}

func (f *FCMSocketHandler) onData() error {
	f.onDataMutex.Lock()
	defer f.onDataMutex.Unlock()

	if f.isWaitingForData {
		f.isWaitingForData = false
		err := f.waitForData()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FCMSocketHandler) waitForData() error {
	minBytesNeeded := 0

	switch f.state {
	case generic.MCS_VERSION_TAG_AND_SIZE:
		minBytesNeeded = generic.KVersionPacketLen + generic.KTagPacketLen + generic.KSizePacketLenMin
		break
	case generic.MCS_TAG_AND_SIZE:
		minBytesNeeded = generic.KTagPacketLen + generic.KSizePacketLenMin
		break
	case generic.MCS_SIZE:
		minBytesNeeded = f.sizePacketSoFar + 1
		break
	case generic.MCS_PROTO_BYTES:
		minBytesNeeded = f.messageSize
		break
	default:
		err := errors.New(`Unexpected state: ` + strconv.Itoa(f.state))
		return err
	}

	if len(f.data) < minBytesNeeded {
		f.isWaitingForData = true
		return nil
	}

	switch f.state {
	case generic.MCS_VERSION_TAG_AND_SIZE:
		f.onGotVersion()
		break
	case generic.MCS_TAG_AND_SIZE:
		f.onGotMessageTag()
		break
	case generic.MCS_SIZE:
		f.onGotMessageSize()
		break
	case generic.MCS_PROTO_BYTES:
		f.onGotMessageBytes()
		break
	default:
		err := errors.New(`Unexpected state: ` + strconv.Itoa(f.state))
		return err
	}

	return nil
}

func (f *FCMSocketHandler) onGotVersion() error {
	version := int(f.data[0])
	f.data = f.data[1:]

	if version < generic.KMCSVersion && version != 38 {
		err := errors.New("Got wrong version: " + strconv.Itoa(version))
		return err
	}

	// Process the LoginResponse message tag.
	err := f.onGotMessageTag()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) onGotMessageTag() error {
	fmt.Println(f.data)
	f.messageTag = int(f.data[0])
	f.data = f.data[1:]

	err := f.onGotMessageSize()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) onGotMessageSize() error {
	var pos int
	f.messageSize, pos = ReadInt32(f.data)
	pos += 1

	f.data = f.data[pos:]

	f.sizePacketSoFar = 0

	if f.messageSize > 0 {
		f.state = generic.MCS_PROTO_BYTES
		err := f.waitForData()
		if err != nil {
			return err
		}
	} else {
		err := f.onGotMessageBytes()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FCMSocketHandler) onGotMessageBytes() error {
	protobuf, err := f.buildProtobufFromTag(f.data[:f.messageSize])
	if err != nil {
		return err
	}
	if protobuf == nil {
		f.data = f.data[f.messageSize:]
		err = errors.New("Unknown message tag " + strconv.Itoa(f.messageTag))
		log.Println(err)
		return err
	}

	if f.messageSize == 0 {
		// Todo: DO
		f.OnMessage(f.messageTag, nil)
		err = f.getNextMessage()
		if err != nil {
			return err
		}
		return nil
	}
	//
	if len(f.data) < f.messageSize {
		f.state = generic.MCS_PROTO_BYTES
		err = f.waitForData()
		if err != nil {
			return err
		}
		return nil
	}
	//
	fmt.Println("The message is: ")
	fmt.Println(f.data[:f.messageSize])
	f.data = f.data[f.messageSize:]

	f.OnMessage(f.messageTag, protobuf)

	if f.messageTag == generic.KLoginResponseTag {
		if !f.handshakeComplete {
			log.Println("Handshake complete")
			f.handshakeComplete = true
		}
	}

	err = f.getNextMessage()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) getNextMessage() error {
	f.messageTag = 0
	f.messageSize = 0
	f.state = generic.MCS_TAG_AND_SIZE
	err := f.waitForData()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) buildProtobufFromTag(buffer []byte) (interface{}, error) {
	switch f.messageTag {
	case generic.KHeartbeatPingTag:
		heartbeatPing, err := fcm_protos.DecodeHeartbeatPing(buffer)
		if err != nil {
			return nil, err
		}
		return heartbeatPing, nil
	case generic.KHeartbeatAckTag:
		heartbeatAck, err := fcm_protos.DecodeHeartbeatAck(buffer)
		if err != nil {
			return nil, err
		}
		return heartbeatAck, nil
	case generic.KLoginRequestTag:
		loginRequest, err := fcm_protos.DecodeLoginRequest(buffer)
		if err != nil {
			return nil, err
		}
		return loginRequest, nil
	case generic.KLoginResponseTag:
		loginResponse, err := fcm_protos.DecodeLoginResponse(buffer)
		if err != nil {
			return nil, err
		}
		return loginResponse, nil
	case generic.KCloseTag:
		closeObject, err := fcm_protos.DecodeClose(buffer)
		if err != nil {
			return nil, err
		}
		return closeObject, nil
	case generic.KIqStanzaTag:
		iqStanza, err := fcm_protos.DecodeIqStanza(buffer)
		if err != nil {
			return nil, err
		}
		return iqStanza, nil
	case generic.KDataMessageStanzaTag:
		dataMessageStanza, err := fcm_protos.DecodeDataMessageStanza(buffer)
		if err != nil {
			return nil, err
		}
		return dataMessageStanza, nil
	case generic.KStreamErrorStanzaTag:
		streamErrorStanza, err := fcm_protos.DecodeStreamErrorStanza(buffer)
		if err != nil {
			return nil, err
		}
		return streamErrorStanza, nil
	default:
		return nil, nil
	}
}

func (f *FCMSocketHandler) Init() {
	f.state = generic.MCS_VERSION_TAG_AND_SIZE
	f.data = nil
	f.sizePacketSoFar = 0
	f.messageTag = 0
	f.messageSize = 0
	f.handshakeComplete = false
	f.isWaitingForData = true
}

func (f *FCMSocketHandler) close() {
	if f.Socket != nil {
		f.Socket.Close()
	}
	f.Init()
}
