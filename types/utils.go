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
