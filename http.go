package go_fcm_receiver

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	pb "go-fcm-receiver/proto"
	"io"
	"log"
	"net/http"
)

func (f *FCMClient) SendCheckInRequest(requestBody *pb.AndroidCheckinRequest) (*pb.AndroidCheckinResponse, error) {
	data, err := proto.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
		return nil, err

	}

	buff := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", CheckInUrl, buff)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-protobuf")
	req.Header.Add("User-Agent", "")

	resp, err := f.HttpClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var responsePb pb.AndroidCheckinResponse
	err = proto.Unmarshal(result, &responsePb)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &responsePb, nil
}
