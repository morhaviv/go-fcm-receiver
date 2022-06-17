package go_fcm_receiver

const CheckInUrl = "https://android.clients.google.com/checkin"
const RegisterUrl = "https://android.clients.google.com/c2dm/register3"

const FcmSubscribeUrl = "https://fcm.googleapis.com/fcm/connect/subscribe"
const FcmEndpointUrl = "https://fcm.googleapis.com/fcm/send"

const AppIdBase = "wp:receiver.push.com#$%s"

const FcmSocketAddress = "mtalk.google.com:5228"

var FcmServerKey = []byte{4, 51, 148, 247, 223, 161, 235, 177, 220, 3, 162, 94, 21, 113, 219, 72, 211, 46, 237, 237, 178, 52, 219, 183, 71, 58, 12, 143, 196, 204, 225, 111, 60, 140, 132, 223, 171, 182, 102, 62, 242, 12, 212, 139, 254, 227, 249, 118, 47, 20, 28, 99, 8, 106, 111, 45, 177, 26, 149, 176, 206, 55, 192, 156, 110}

// Processing the version, tag, and size packets (assuming minimum length
// size packet). Only used during the login handshake.
const (
	MCS_VERSION_TAG_AND_SIZE = 0
	MCS_TAG_AND_SIZE         = 1
	// Processing the size packet alone.
	MCS_SIZE = 2
	// Processing the protocol buffer bytes (for those messages with non-zero sizes).
	MCS_PROTO_BYTES = 3

	// # of bytes a MCS version packet consumes.
	kVersionPacketLen = 1
	// # of bytes a tag packet consumes.
	kTagPacketLen = 1
	// Max # of bytes a length packet consumes. A Varint32 can consume up to 5 bytes
	// (the msb in each byte is reserved for denoting whether more bytes follow).
	// Although the protocol only allows for 4KiB payloads currently, and the socket
	// stream buffer is only of size 8KiB, it's possible for certain applications to
	// have larger message sizes. When payload is larger than 4KiB, an temporary
	// in-memory buffer is used instead of the normal in-place socket stream buffer.
	kSizePacketLenMin = 1
	kSizePacketLenMax = 5

	// The current MCS protocol version.
	KMCSVersion = 41

	// MCS Message tags.
	// WARNING= the order of these tags must remain the same, as the tag values
	// must be consistent with those used on the server.
	kHeartbeatPingTag       = 0
	kHeartbeatAckTag        = 1
	KLoginRequestTag        = 2
	kLoginResponseTag       = 3
	kCloseTag               = 4
	kMessageStanzaTag       = 5
	kPresenceStanzaTag      = 6
	kIqStanzaTag            = 7
	kDataMessageStanzaTag   = 8
	kBatchPresenceStanzaTag = 9
	kStreamErrorStanzaTag   = 10
	kHttpRequestTag         = 11
	kHttpResponseTag        = 12
	kBindAccountRequestTag  = 13
	kBindAccountResponseTag = 14
	kTalkMetadataTag        = 15
	kNumProtoTypes          = 16
)
