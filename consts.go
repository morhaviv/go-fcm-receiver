package go_fcm_receiver

const CheckInUrl = "https://android.clients.google.com/checkin"
const RegisterUrl = "https://android.clients.google.com/c2dm/register3"

const FcmSubscribeUrl = "https://fcm.googleapis.com/fcm/connect/subscribe"
const FcmEndpointUrl = "https://fcm.googleapis.com/fcm/send"

const AppIdBase = "wp:receiver.push.com#$%s"

const FcmSocketAddress = "mtalk.google.com:5228"

var FcmServerKey = []byte{4, 51, 148, 247, 223, 161, 235, 177, 220, 3, 162, 94, 21, 113, 219, 72, 211, 46, 237, 237, 178, 52, 219, 183, 71, 58, 12, 143, 196, 204, 225, 111, 60, 140, 132, 223, 171, 182, 102, 62, 242, 12, 212, 139, 254, 227, 249, 118, 47, 20, 28, 99, 8, 106, 111, 45, 177, 26, 149, 176, 206, 55, 192, 156, 110}
