package go_fcm_receiver

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

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

	fmt.Println(string(f.data))
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
	case MCS_VERSION_TAG_AND_SIZE:
		minBytesNeeded = kVersionPacketLen + kTagPacketLen + kSizePacketLenMin
		break
	case MCS_TAG_AND_SIZE:
		minBytesNeeded = kTagPacketLen + kSizePacketLenMin
		break
	case MCS_SIZE:
		minBytesNeeded = f.sizePacketSoFar + 1
		break
	case MCS_PROTO_BYTES:
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
	case MCS_VERSION_TAG_AND_SIZE:
		f.onGotVersion()
		break
	case MCS_TAG_AND_SIZE:
		f.onGotMessageTag()
		break
	case MCS_SIZE:
		f.onGotMessageSize()
		break
	case MCS_PROTO_BYTES:
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

	if version < kMCSVersion && version != 38 {
		err := errors.New("Got wrong version: " + strconv.Itoa(version))
		return err
	}

	fmt.Println("version ", version, strconv.Itoa(version))

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

	fmt.Println("MESSAGE TAG", f.messageTag, strconv.Itoa(f.messageTag))

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
	fmt.Println("POSSSS", pos)

	fmt.Println("MESSAGE SIZE", f.messageSize, strconv.Itoa(f.messageSize))

	f.data = f.data[pos:]

	f.sizePacketSoFar = 0

	if f.messageSize > 0 {
		f.state = MCS_PROTO_BYTES
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
	protobuf := f.buildProtobufFromTag(f.data[:f.messageSize])
	if protobuf == nil {
		f.data = f.data[f.messageSize:]
		err := errors.New("Unknown message tag " + strconv.Itoa(f.messageTag))
		log.Println(err)
		return err
	}

	//// Messages with no content are valid; just use the default protobuf for
	//// that tag.
	if f.messageSize == 0 {
		// Todo: DO
		//this.emit('message', {tag: this._messageTag, object: {}});
		err := f.getNextMessage()
		if err != nil {
			return err
		}
		return nil
	}
	//
	if len(f.data) < f.messageSize {
		f.state = MCS_PROTO_BYTES
		err := f.waitForData()
		if err != nil {
			return err
		}
		return nil
	}
	//
	fmt.Println("The message is: ")
	fmt.Println(f.data[:f.messageSize])
	f.data = f.data[f.messageSize:]
	//const object = protobuf.toObject(message, {
	//longs : String,
	//	enums : String,
	//		bytes : Buffer,
	//});
	//
	//this.emit('message', {tag: this._messageTag, object: object});
	//
	if f.messageTag == kLoginResponseTag {
		if !f.handshakeComplete {
			fmt.Println("Handshake complete")
			f.handshakeComplete = true
		}
	}

	err := f.getNextMessage()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) getNextMessage() error {
	f.messageTag = 0
	f.messageSize = 0
	f.state = MCS_TAG_AND_SIZE
	err := f.waitForData()
	if err != nil {
		return err
	}
	return nil
}

func (f *FCMSocketHandler) buildProtobufFromTag(buffer []byte) interface{} {
	return f.messageTag
	return nil
	//switch f.messageTag {
	//case kHeartbeatPingTag:
	//	return CreateHeartBeatPing(buffer)
	//	//return proto.lookupType('mcs_proto.HeartbeatPing')
	//case kHeartbeatAckTag:
	//	return CreateHeartBeatAck(buffer)
	//	//return proto.lookupType('mcs_proto.HeartbeatAck')
	//case kLoginRequestTag:
	//	return CreateLoginRequest(buffer)
	//	//return proto.lookupType('mcs_proto.LoginRequest')
	//case kLoginResponseTag:
	//	return CreateLoginResponse(buffer)
	//	//return proto.lookupType('mcs_proto.LoginResponse')
	//case kCloseTag:
	//	return CreateClose(buffer)
	//	//return proto.lookupType('mcs_proto.Close')
	//case kIqStanzaTag:
	//	return CreateIqStanza(buffer)
	//	//return proto.lookupType('mcs_proto.IqStanza')
	//case kDataMessageStanzaTag:
	//	return CreateDataMessageStanza(buffer)
	//	//return proto.lookupType('mcs_proto.DataMessageStanza')
	//case kStreamErrorStanzaTag:
	//	return CreateStreamErrorStanza(buffer)
	//	//return proto.lookupType('mcs_proto.StreamErrorStanza')
	//default:
	//	return nil
	//}
}

func (f *FCMSocketHandler) Init() {
	f.state = MCS_VERSION_TAG_AND_SIZE
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
