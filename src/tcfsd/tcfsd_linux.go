package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"syscall"
	"tcfs"
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
	GETATTR  = 0x01
	READLINK = 0x02
	GETDIR   = 0x03
	MKNOD    = 0x04
	MKDIR    = 0x05
	SYMLINK  = 0x06
	UNLINK   = 0x07
	RMDIR    = 0x08
	RENAME   = 0x09
	CHMOD    = 0x0A
	CHOWN    = 0x0B
	TRUNCATE = 0x0C
	UTIME    = 0x0D
	OPEN     = 0x0E
	READ     = 0x0F
	WRITE    = 0x10
	READDIR  = 0x11
	RELEASE  = 0x12
	CREATE   = 0x13
)

func handleConn(tConn *TcfsConn) {
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
			fixpath := rootdir + string(msgbuf)
			var stat syscall.Stat_t
			err = syscall.Lstat(fixpath, &stat)
			if err != nil {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				var ret int32 = -2
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
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
			// fmt.Println(fixpath)
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
			// fmt.Println(string(msgbuf))
			// fmt.Println(fixpath)
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
			atime := binary.BigEndian.Uint64(msgbuf[0:8])
			mtime := binary.BigEndian.Uint64(msgbuf[8:16])
			fixpath := rootdir + string(msgbuf[16:])
			err := syscall.Utime(fixpath, &syscall.Utimbuf{int64(atime), int64(mtime)})
			if err != nil {
				log.Print("Can't create", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				tConn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			tConn.Write(buf[:8])
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
			// fmt.Println(size)
			// fmt.Println(len(msgbuf))
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

var (
	port         = flag.String("port", ":9876", "port to listen to")
	rootpath     = flag.String("dir", "rootdir", "path to share")
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
