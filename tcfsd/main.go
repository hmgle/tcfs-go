package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/hmgle/tcfs-go"
)

type TcfsConn struct {
	RootDir    string
	Conn       net.Conn
	Buf        []byte
	OpenedFile map[uintptr]*os.File

	cipher *tcfs.Cipher
}

func (c *TcfsConn) Read(b []byte) (int, error) {
	if c.cipher == nil {
		return c.Conn.Read(b)
	}
	cipherData := make([]byte, len(b))
	n, err := c.Conn.Read(cipherData)
	if n > 0 {
		c.cipher.Decrypt(b[0:n], cipherData[0:n])
	}
	return n, err
}

func (c *TcfsConn) Write(b []byte) (int, error) {
	if c.cipher == nil {
		return c.Conn.Write(b)
	}
	cipherData := make([]byte, len(b))
	c.cipher.Encrypt(cipherData, b)
	return c.Conn.Write(cipherData)
}

func (c *TcfsConn) Close() {
	c.Conn.Close()
}

const (
	GETATTR  uint32 = 0x01
	READLINK uint32 = 0x02
	GETDIR   uint32 = 0x03
	MKNOD    uint32 = 0x04
	MKDIR    uint32 = 0x05
	SYMLINK  uint32 = 0x06
	UNLINK   uint32 = 0x07
	RMDIR    uint32 = 0x08
	RENAME   uint32 = 0x09
	CHMOD    uint32 = 0x0A
	CHOWN    uint32 = 0x0B
	TRUNCATE uint32 = 0x0C
	UTIME    uint32 = 0x0D
	OPEN     uint32 = 0x0E
	READ     uint32 = 0x0F
	WRITE    uint32 = 0x10
	READDIR  uint32 = 0x11
	RELEASE  uint32 = 0x12
	CREATE   uint32 = 0x13
)

var (
	port         = flag.String("port", ":9876", "port to listen to")
	rootpath     = flag.String("dir", ".", "path to share")
	cryptoMethod = flag.String("crypto", "rc4", "encryption method")
	key          = flag.String("key", "", "password used to encrypt the data")
)

func main() {
	var cipher *tcfs.Cipher
	flag.Parse()
	if len(*key) > 0 {
		cipher = tcfs.NewCipher(*cryptoMethod, []byte(*key))
	} else {
		cipher = nil
	}
	l, e := net.Listen("tcp", *port)
	if e != nil {
		log.Fatal(e)
	}
	for {
		conn, e := l.Accept()
		if e != nil {
			log.Print(e)
			continue
		}
		newConn := TcfsConn{
			*rootpath,
			conn,
			make([]byte, 4096*1024),
			map[uintptr]*os.File{},
			cipher,
		}
		go handleConn(&newConn)
	}
}
