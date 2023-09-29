package runtimetypes

func int32ToBytes(value int32) []byte {
	var data [4]byte
	data[0] = uint8(value)
	data[1] = uint8(value >> 8)
	data[2] = uint8(value >> 16)
	data[3] = uint8(value >> 24)
	return data[:]
}

func bytesToInt32(data []byte) int32 {
	return int32(data[0]) + int32(data[1])<<8 + int32(data[2])<<16 + int32(data[3])<<24
}

func int64ToBytes(value int64) []byte {
	var data [8]byte
	data[0] = uint8(value)
	data[1] = uint8(value >> 8)
	data[2] = uint8(value >> 16)
	data[3] = uint8(value >> 24)
	data[4] = uint8(value >> 32)
	data[5] = uint8(value >> 40)
	data[6] = uint8(value >> 48)
	data[7] = uint8(value >> 56)

	return data[:]
}

func bytesToInt64(data []byte) int64 {
	return int64(data[0]) + int64(data[1])<<8 + int64(data[2])<<16 + int64(data[3])<<24 + int64(data[4])<<32 + int64(data[5])<<40 + int64(data[6])<<48 + int64(data[7])<<56
}
