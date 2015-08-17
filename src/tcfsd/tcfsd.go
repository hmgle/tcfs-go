package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
)

func handleConn(conn net.Conn) {
	var err error
	for {
		buf := make([]byte, 4)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			break
		}
		msglen := binary.BigEndian.Uint32(buf)
		buf = make([]byte, msglen)
		_, err = io.ReadFull(conn, buf)
		// println(string(buf))
	}
}

func main() {
	l, e := net.Listen("tcp", ":9876")
	if e != nil {
		log.Fatal(e)
		return
	}
	for {
		conn, e := l.Accept()
		if e != nil {
			log.Fatal(e)
			continue
		}
		go handleConn(conn)
	}
}
