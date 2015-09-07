package tcfs

import (
	"crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
	"log"
)

type Cipher struct {
	enc cipher.Stream
	dec cipher.Stream
}

type chiperCreator func(key []byte) (*Cipher, error)

var cipherMap = map[string]chiperCreator{
	"rc4": newRC4Cipher,
}

func secretToKey(secret []byte, size int) []byte {
	// size mod 16 must be 0
	h := md5.New()
	buf := make([]byte, size)
	count := size / md5.Size
	// repeatly fill the key with the secret
	for i := 0; i < count; i++ {
		h.Write(secret)
		copy(buf[md5.Size*i:md5.Size*(i+1)], h.Sum(nil))
	}
	return buf
}

func newRC4Cipher(secret []byte) (*Cipher, error) {
	c, err := rc4.NewCipher(secretToKey(secret, 16))
	if err != nil {
		return nil, err
	}
	c2 := *c

	return &Cipher{c, &c2}, nil
}

func NewCipher(cryptoMethod string, secret []byte) *Cipher {
	cc := cipherMap[cryptoMethod]
	if cc == nil {
		log.Fatalf("unsupported crypto method %s", cryptoMethod)
	}
	c, err := cc(secret)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func (c *Cipher) Encrypt(dst, src []byte) {
	c.enc.XORKeyStream(dst, src)
}

func (c *Cipher) Decrypt(dst, src []byte) {
	c.dec.XORKeyStream(dst, src)
}
