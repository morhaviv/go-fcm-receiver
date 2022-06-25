# go-fcm-receiver

A library to subscribe to a GCM/FCM (Firebase Cloud Messaging) sender using [protobuf](https://developers.google.com/protocol-buffers).

The library was inspired by [push-receiver](https://www.npmjs.com/package/push-receiver) and by [this blog post](https://medium.com/@MatthieuLemoine/my-journey-to-bring-web-push-support-to-node-and-electron-ce70eea1c0b0).

## The difference between this library and other libraries

- This library receives FCM notifications the same way an Android device does. This library is an FCM client.
- Other libraries (such as go-fcm) sends notifications via fcm, and does not receive notifications. Those libraries are an FCM server-side.  

## Install

`
import "github.com/morhaviv/go-fcm-receiver"
`

## Requirements

- Firebase sender id to receive notification

## Usage

### Creating a new device and starting to listen
```Go
package main

import (
	go_fcm_receiver "github.com/morhaviv/go-fcm-receiver"
)

func main() {
	newDevice := go_fcm_receiver.FCMClient{
		SenderId:          845754665638,
		OnDataMessage: func(message []byte) {
			fmt.Println("Received a message:", string(message))
		},
	}
	privateKey, authSecret, err := newDevice.CreateNewKeys()
	if err != nil {
		panic(err)
	}
	fcmToken, gcmToken, androidId, securityToken, err := newDevice.Register()
	if err != nil {
		panic(err)
	}
	SaveDeviceDetails(fcmToken, gcmToken, androidId, securityToken, privateKey, authSecret)
	err = newDevice.StartListening()
	if err != nil {
		panic(err)
	}

}
```

### Starting to listen an old device
```Go
package main

import (
	go_fcm_receiver "github.com/morhaviv/go-fcm-receiver"
)

func main() {
	oldDevice := go_fcm_receiver.FCMClient{
		SenderId:          845754665638, // Firebase Project ID
		GcmToken:          "<GCM_TOKEN>",
		FcmToken:          "<FCM_TOKEN>",
		AndroidId:         5240887932061714513, // The androidId returned when the device was created
		SecurityToken:     69534515778185919, // The securityToken returned when the device was created
		OnDataMessage: func(message []byte) {
			fmt.Println("Received a message:", string(message))
		},
	}

	err := oldDevice.StartListening()
	if err != nil {
		panic(err)
	}
}
```
