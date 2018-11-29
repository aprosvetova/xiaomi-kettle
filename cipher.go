package kettle

const KeySize = 256

func CipherInit(key []byte) []byte {
	var perm = make([]byte, KeySize)
	for i := range perm {
		perm[i] = byte(i)
	}
	var j uint8
	keyLen := uint8(len(key))
	for ia := range perm {
		j += perm[ia] + key[uint8(ia)%keyLen]
		perm[ia], perm[j] = perm[j], perm[ia]
	}
	return perm
}

func CipherCrypt(input, perm []byte) []byte {
	var index1, index2 uint8
	output := make([]byte, len(input))
	for i := 0; i < len(input); i++ {
		index1++
		index2 += perm[index1]
		perm[index1], perm[index2] = perm[index2], perm[index1]
		idx := perm[index1] + perm[index2]
		output[i] = input[i] ^ perm[idx]
	}
	return output
}

func Cipher(key, input []byte) []byte {
	perm := CipherInit(key)
	return CipherCrypt(input, perm)
}

func MixA(mac []byte, productID int) []byte {
	return []byte{mac[0], mac[2], mac[5], uint8(productID & 0xff), uint8(productID & 0xff), mac[4], mac[5], mac[1]}
}

func MixB(mac []byte, productID int) []byte {
	return []byte{mac[0], mac[2], mac[5], uint8((productID >> 8) & 0xff), mac[4], mac[0], mac[5], uint8(productID & 0xff)}
}
