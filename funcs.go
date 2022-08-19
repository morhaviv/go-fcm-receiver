package go_fcm_receiver

import (
	"errors"
)

func StringsSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ReadUint32(buf []byte) (int, int, error) {
	value := 4294967295
	// optimizer type-hint, tends to deopt otherwise (?!)
	pos := 0

	value = (int(buf[pos]) & 127) >> uint(0)
	if buf[pos] < 128 {
		return value, pos, nil
	}
	if len(buf) < 2 {
		return pos, value, errors.New("not enough bytes for ReadUint32")
	}

	pos++
	value = (value | (int(buf[pos])&127)<<7) >> uint(0)
	if buf[pos] < 128 {
		return value, pos, nil
	}
	if len(buf) < 3 {
		return pos, value, errors.New("not enough bytes for ReadUint32")
	}

	pos++
	value = (value | (int(buf[pos])&127)<<14) >> uint(0)
	if buf[pos] < 128 {
		return value, pos, nil
	}
	if len(buf) < 3 {
		return pos, value, errors.New("not enough bytes for ReadUint32")
	}

	pos++
	value = (value | (int(buf[pos])&127)<<21) >> uint(0)
	if buf[pos] < 128 {
		return value, pos, nil
	}
	if len(buf) < 4 {
		return pos, value, errors.New("not enough bytes for ReadUint32")
	}

	pos++
	value = (value | (int(buf[pos])&15)<<28) >> uint(0)
	if buf[pos] < 128 {
		return value, pos, nil
	}

	return value, pos, nil
}

func ReadInt32(buf []byte) (int, int, error) {
	value, pos, err := ReadUint32(buf)
	return value | 0, pos, err
}
