# go-fcm-receiver

A library to subscribe to a GCM/FCM (Firebase Cloud Messaging) sender using [protobuf](https://developers.google.com/protocol-buffers).

The library was inspired by [push-receiver](https://www.npmjs.com/package/push-receiver) and by [this blog post](https://medium.com/@MatthieuLemoine/my-journey-to-bring-web-push-support-to-node-and-electron-ce70eea1c0b0).

## The difference between this library and others

- This library receives FCM notifications the same way an Android device does. This library is an FCM client.
- Other libraries (such as go-fcm) sends notifications via fcm, and does not receive notifications. Those libraries are an FCM server-side.  

## Install

`
import "github.com/morhaviv/go-fcm-receiver"
`

## Requirements

- Firebase sender id to receive notification

## Usage

```Go
package main

// Soon
```
