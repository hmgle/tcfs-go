package tcfs

import (
	"encoding/binary"
)

func utimeHandle(tConn *TcfsConn, msgbuf []byte) {
	// TODO
	buf := tConn.Buf
	binary.BigEndian.PutUint32(buf[0:4], 4)
	binary.BigEndian.PutUint32(buf[4:8], 0)
	tConn.Write(buf[:8])
}
