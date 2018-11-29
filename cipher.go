package kettle

func cipherInit(key []byte) []byte {
	var perm = make([]uint8, 0, 256)
	perm = append(perm, []uint8(key)...)
	var ia int
	var j uint8
	keyLen := uint8(len(key))
	for ia <= 255 {
		j += uint8(perm[ia] + key[uint8(ia)%keyLen])
		perm[ia], perm[j] = perm[j], perm[ia]
		ia++
	}
	return perm
}

func cipherCrypt(input, perm []byte) []byte {
	var index1, index2 uint8
	output := make([]byte, len(input))
	for i := 0; i < len(input); i++ {
		index1++
		index2 += uint8(perm[index1])
		perm[index1], perm[index2] = perm[index2], perm[index1]
		idx := perm[index1] + perm[index2]
		output[i] = input[i] ^ perm[idx]
	}
	return output
}

func cipher(key, input []byte) []byte {
	perm := cipherInit(key)
	return cipherCrypt(input, perm)
}

func mixA(mac []byte, productID int) []byte {
	return []byte{mac[0], mac[2], mac[5], uint8(productID & 0xff), uint8(productID & 0xff), mac[4], mac[5], mac[1]}
}

func mixB(mac []byte, productID int) []byte {
	return []byte{mac[0], mac[2], mac[5], uint8((productID >> 8) & 0xff), mac[4], mac[0], mac[5], uint8(productID & 0xff)}
}
