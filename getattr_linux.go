package tcfs

import (
	"encoding/binary"
	"syscall"
)

func getattrHandle(tConn *TcfsConn, msgbuf []byte) {
	rootdir := tConn.RootDir
	buf := tConn.Buf
	fixpath := rootdir + string(msgbuf)

	var stat syscall.Stat_t
	err := syscall.Lstat(fixpath, &stat)
	if err != nil {
		binary.BigEndian.PutUint32(buf[0:4], 4)
		var ret int32 = -2
		binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
		tConn.Write(buf[:8])
		return
	}
	binary.BigEndian.PutUint32(buf[0:4], 11*4)
	binary.BigEndian.PutUint32(buf[4:8], 0)
	binary.BigEndian.PutUint32(buf[8:12], uint32(stat.Dev))
	binary.BigEndian.PutUint32(buf[12:16], uint32(stat.Ino))
	binary.BigEndian.PutUint32(buf[16:20], stat.Mode)
	binary.BigEndian.PutUint32(buf[20:24], uint32(stat.Nlink))
	binary.BigEndian.PutUint32(buf[24:28], stat.Uid)
	binary.BigEndian.PutUint32(buf[28:32], stat.Gid)
	binary.BigEndian.PutUint32(buf[32:36], uint32(stat.Size))
	binary.BigEndian.PutUint32(buf[36:40], uint32(stat.Atim.Sec))
	binary.BigEndian.PutUint32(buf[40:44], uint32(stat.Mtim.Sec))
	binary.BigEndian.PutUint32(buf[44:48], uint32(stat.Ctim.Sec))
	tConn.Write(buf[:48])
}
