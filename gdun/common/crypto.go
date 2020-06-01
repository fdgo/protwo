package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

//----------------------------------------------------------------------------

func Encrypt(src []byte, key []byte, iv []byte) ([]byte, error) {
	if src == nil || len(src) == 0 {
		return nil, ERR_INVALID_SOURCE
	}
	if key == nil || len(key) < 16 {
		return nil, ERR_INVALID_KEY
	}

	b, err := aes.NewCipher(key[:16])
	if err != nil {
		return nil, err
	}

	paddedSrc := src
	if len(paddedSrc)%b.BlockSize() != 0 {
		nPadding := b.BlockSize() - len(paddedSrc)%b.BlockSize()
		padding := bytes.Repeat([]byte{byte(nPadding)}, nPadding)
		paddedSrc = append(src, padding...)
	}

	dst := make([]byte, len(paddedSrc))
	bm := cipher.NewCBCEncrypter(b, iv[:aes.BlockSize])
	bm.CryptBlocks(dst, paddedSrc)

	return dst, nil
}

//----------------------------------------------------------------------------

func Decrypt(src []byte, key []byte, iv []byte) ([]byte, error) {
	if src == nil || len(src) == 0 || len(src)%16 != 0 {
		return nil, ERR_INVALID_SOURCE
	}
	if key == nil || len(key) < 16 {
		return nil, ERR_INVALID_KEY
	}

	b, err := aes.NewCipher(key[:16])
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(src))
	bm := cipher.NewCBCDecrypter(b, iv[:aes.BlockSize])
	bm.CryptBlocks(dst, src)

	n := len(dst)
	end := n - int(dst[n-1])

	// The following scenarios should never occur
	//   if the encrypted text is correct.
	if end < 0 {
		end = 0
	}
	if end > n-1 {
		end = n - 1
	}

	return dst[:end], nil
}

//----------------------------------------------------------------------------
