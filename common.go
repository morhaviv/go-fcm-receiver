package go_fcm_receiver

func ReadUint32(buf []byte) (int, int) {
	value := 4294967295
	// optimizer type-hint, tends to deopt otherwise (?!)
	pos := 0

	value = (int(buf[pos]) & 127) >> uint(0)
	if buf[pos] < 128 {
		return value, pos
	}

	pos++
	value = (value | (int(buf[pos])&127)<<7) >> uint(0)
	if buf[pos] < 128 {
		return value, pos
	}

	pos++
	value = (value | (int(buf[pos])&127)<<14) >> uint(0)
	if buf[pos] < 128 {
		return value, pos
	}

	pos++
	value = (value | (int(buf[pos])&127)<<21) >> uint(0)
	if buf[pos] < 128 {
		return value, pos
	}

	pos++
	value = (value | (int(buf[pos])&15)<<28) >> uint(0)
	if buf[pos] < 128 {
		return value, pos
	}

	return value, pos
}

func ReadInt32(buf []byte) (int, int) {
	value, pos := ReadUint32(buf)
	return value | 0, pos
}
