package tcfs

import (
	"encoding/binary"
	"log"
	"syscall"
)

func utimeHandle(tConn *TcfsConn, msgbuf []byte) {
	rootdir := tConn.RootDir
	buf := tConn.Buf
	fixpath := rootdir + string(msgbuf[16:])
	// Linux
	// atime := binary.BigEndian.Uint64(msgbuf[0:8])
	// mtime := binary.BigEndian.Uint64(msgbuf[8:16])
	// err := syscall.Utime(fixpath, &syscall.Utimbuf{int64(atime), int64(mtime)})
	// TODO
	err := syscall.Utimes(fixpath, []syscall.Timeval{})
	if err != nil {
		log.Print("Can't create", err)
		binary.BigEndian.PutUint32(buf[0:4], 4)
		ret := -13
		binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
		tConn.Write(buf[:8])
	} else {
		binary.BigEndian.PutUint32(buf[0:4], 4)
		binary.BigEndian.PutUint32(buf[4:8], 0)
		tConn.Write(buf[:8])
	}
}
