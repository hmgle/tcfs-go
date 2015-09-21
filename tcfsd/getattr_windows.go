package main

import (
	"encoding/binary"
	"os"
)

func getattr_handle(tConn *TcfsConn, msgbuf []byte) {
	// FIXME
	rootdir := tConn.RootDir
	buf := tConn.Buf
	fixpath := rootdir + string(msgbuf)
	fiInfo, err2 := os.Lstat(fixpath)
	if err2 != nil {
		binary.BigEndian.PutUint32(buf[0:4], 4)
		var ret int32 = -2
		binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
		tConn.Write(buf[:8])
		return
	}
	binary.BigEndian.PutUint32(buf[0:4], 11*4)
	binary.BigEndian.PutUint32(buf[4:8], 0)
	binary.BigEndian.PutUint32(buf[8:12], 2049)     // dev
	binary.BigEndian.PutUint32(buf[12:16], 3672782) // ino
	if fiInfo.IsDir() {
		binary.BigEndian.PutUint32(buf[16:20], 16893) // mode
	} else {
		binary.BigEndian.PutUint32(buf[16:20], 33261) // mode
	}
	if fiInfo.IsDir() {
		binary.BigEndian.PutUint32(buf[20:24], 2) // nlink
	} else {
		binary.BigEndian.PutUint32(buf[20:24], 1) // nlink
	}
	binary.BigEndian.PutUint32(buf[24:28], 1000)                            // uid
	binary.BigEndian.PutUint32(buf[28:32], 1000)                            // gid
	binary.BigEndian.PutUint32(buf[32:36], uint32(fiInfo.Size()))           // size
	binary.BigEndian.PutUint32(buf[36:40], uint32(fiInfo.ModTime().Unix())) // Atime
	binary.BigEndian.PutUint32(buf[40:44], uint32(fiInfo.ModTime().Unix())) // Mtime
	binary.BigEndian.PutUint32(buf[44:48], uint32(fiInfo.ModTime().Unix())) // Ctime
	tConn.Write(buf[:48])
}
