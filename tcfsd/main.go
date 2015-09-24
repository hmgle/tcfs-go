package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/hmgle/tcfs-go"
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
		newConn := tcfs.TcfsConn{
			*rootpath,
			conn,
			make([]byte, 4096*1024),
			map[uintptr]*os.File{},
			cipher,
		}
		go tcfs.HandleConn(&newConn)
	}
}
