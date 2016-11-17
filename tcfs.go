package tcfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type TcfsConn struct {
	RootDir    string
	Conn       net.Conn
	Buf        []byte
	OpenedFile map[uintptr]*os.File

	Cipher *Cipher
}

func (c *TcfsConn) Read(b []byte) (int, error) {
	if c.Cipher == nil {
		return c.Conn.Read(b)
	}
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.Cipher.Decrypt(b[0:n], b[0:n])
	}
	return n, err
}

func (c *TcfsConn) Write(b []byte) (int, error) {
	if c.Cipher == nil {
		return c.Conn.Write(b)
	}
	c.Cipher.Encrypt(b, b)
	return c.Conn.Write(b)
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

func HandleConn(tConn *TcfsConn) {
	defer tConn.Conn.Close()
	var err error
	buf := tConn.Buf
	openedFile := tConn.OpenedFile
	rootdir := tConn.RootDir
	for {
		_, err = io.ReadFull(tConn, buf[:4])
		if err != nil {
			break
		}
		msglen := binary.BigEndian.Uint32(buf[:4])
		if msglen < 4 || msglen > (4096*1024) {
			log.Fatal("msglen = ", msglen)
		}
		_, err = io.ReadFull(tConn, buf[:msglen])

		tcfsOp := binary.BigEndian.Uint32(buf[0:4])
		msgbuf := buf[4:msglen]
		switch tcfsOp {
		case GETATTR:
			getattrHandle(tConn, msgbuf)
		case READLINK:
		case GETDIR:
		case MKNOD:
		case MKDIR:
			mode := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			if err := os.MkdirAll(fixpath, os.FileMode(mode)); err != nil {
				log.Print("Can't create dir", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case SYMLINK:
		case UNLINK:
			// FIXME
			fixpath := rootdir + string(msgbuf)
			if err := os.Remove(fixpath); err != nil {
				log.Print("Can't rmdir", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case RMDIR:
			fixpath := rootdir + string(msgbuf)
			if err := os.Remove(fixpath); err != nil {
				log.Print("Can't rmdir", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case RENAME:
		case CHMOD:
			mode := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			err := os.Chmod(fixpath, os.FileMode(mode))
			if err != nil {
				log.Print("Can't chmod", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case CHOWN:
		case TRUNCATE:
			newSize := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			err := os.Truncate(fixpath, int64(newSize))
			if err != nil {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13 // EACCES
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case UTIME:
			utimeHandle(tConn, msgbuf)
		case OPEN:
			flag := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			f, err := os.OpenFile(fixpath, int(flag), os.ModePerm)
			if err != nil {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				var ret int32 = -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			fd := f.Fd()
			openedFile[fd] = f
			binary.BigEndian.PutUint32(buf[0:4], 8)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			binary.BigEndian.PutUint32(buf[8:12], uint32(fd))
			tConn.Write(buf[:12])
		case READ:
			findex := binary.BigEndian.Uint32(msgbuf[:4])
			offset := binary.BigEndian.Uint32(msgbuf[4:8])
			size := binary.BigEndian.Uint32(msgbuf[8:12])
			f := openedFile[uintptr(findex)]
			readbuf := make([]byte, size)
			readed, err := f.ReadAt(readbuf, int64(offset))
			if err != nil && err != io.EOF {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				var ret int32 = -9
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			if readed == 0 {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				binary.BigEndian.PutUint32(buf[4:8], 0)
				tConn.Write(buf[:8])
			} else if readed > 0 {
				binary.BigEndian.PutUint32(buf[0:4], 4+uint32(readed))
				binary.BigEndian.PutUint32(buf[4:8], uint32(readed))
				copy(buf[8:], readbuf)
				tConn.Write(buf[:8+readed])
			}
		case WRITE:
			findex := binary.BigEndian.Uint32(msgbuf[:4])
			offset := binary.BigEndian.Uint32(msgbuf[4:8])
			size := binary.BigEndian.Uint32(msgbuf[8:12])
			wbuf := msgbuf[12 : 12+size]
			f := openedFile[uintptr(findex)]
			writed, _ := f.WriteAt(wbuf, int64(offset))
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], uint32(writed))
			tConn.Write(buf[:8])
		case READDIR:

			fixpath := rootdir + string(msgbuf)
			fileList := []byte{}
			dirInfo, err := ioutil.ReadDir(fixpath)
			if err != nil {
				log.Print("Can't ReadDir", err)
				continue
			}
			for _, f := range dirInfo {
				fileList = append(fileList, []byte(f.Name())...)
				fileList = append(fileList, 0)
			}

			binary.BigEndian.PutUint32(buf[:4], uint32(len(fileList))+4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			copy(buf[8:], fileList)
			tConn.Write(buf[:len(fileList)+8])
		case RELEASE:
			findex := binary.BigEndian.Uint32(msgbuf[:4])
			f := openedFile[uintptr(findex)]
			err := f.Close()
			if err != nil {
				fmt.Println(err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -9
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
		case CREATE:
			// mode := binary.BigEndian.Uint32([]byte(matched[1])[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			f, err := os.Create(fixpath)
			if err != nil {
				log.Print("Can't create", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			fd := f.Fd()
			openedFile[fd] = f
			binary.BigEndian.PutUint32(buf[0:4], 8)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			binary.BigEndian.PutUint32(buf[8:12], uint32(fd))
			tConn.Write(buf[:12])
		default:
			log.Print("bad tcfsOp: ", tcfsOp)
		}
	}
}
