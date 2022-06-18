# ecego
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/xakep666/ecego)
[![codecov](https://codecov.io/gh/xakep666/ecego/branch/master/graph/badge.svg)](https://codecov.io/gh/xakep666/ecego)
[![Go Report Card](https://goreportcard.com/badge/github.com/xakep666/ecego)](https://goreportcard.com/report/github.com/xakep666/ecego)
![](https://github.com/xakep666/ecego/workflows/Main/badge.svg)

Encrypted Content Encoding implementation in Go to 
operate with encrypted webpush payloads.

Implemented modes:
* aes128gcm (RFC 8188)
* aesgcm (draft-ietf-webpush-encryption-04)
* aesgcm128 (draft-ietf-webpush-encryption-03)
