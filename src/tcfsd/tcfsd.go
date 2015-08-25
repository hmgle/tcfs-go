package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

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

func handleConn(conn net.Conn) {
	defer conn.Close()
	var err error
	buf := make([]byte, 4096*1024) // 4MB
	openedFile := map[uintptr]*os.File{}
	rootdir := "/home/gle/code_repo/cloud_lib/tcfs-go/rootdir"
	for {
		_, err = io.ReadFull(conn, buf[:4])
		if err != nil {
			break
		}
		msglen := binary.BigEndian.Uint32(buf[:4])
		if msglen < 4 || msglen > (4096*1024) {
			log.Fatal("msglen = ", msglen)
		}
		_, err = io.ReadFull(conn, buf[:msglen])

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
				conn.Write(buf[:8])
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
			conn.Write(buf[:48])
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
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
		case SYMLINK:
		case UNLINK:
			// FIXME
			fixpath := rootdir + string(msgbuf)
			fmt.Println(fixpath)
			if err := os.Remove(fixpath); err != nil {
				log.Print("Can't rmdir", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
		case RMDIR:
			fixpath := rootdir + string(msgbuf)
			fmt.Println(string(msgbuf))
			fmt.Println(fixpath)
			if err := os.Remove(fixpath); err != nil {
				log.Print("Can't rmdir", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
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
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
		case CHOWN:
		case TRUNCATE:
			newSize := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			err := os.Truncate(fixpath, int64(newSize))
			if err != nil {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13 // EACCES
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
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
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
		case OPEN:
			flag := binary.BigEndian.Uint32(msgbuf[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			f, err := os.OpenFile(fixpath, int(flag), os.ModePerm)
			if err != nil {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				var ret int32 = -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			fd := f.Fd()
			openedFile[fd] = f
			binary.BigEndian.PutUint32(buf[0:4], 8)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			binary.BigEndian.PutUint32(buf[8:12], uint32(fd))
			conn.Write(buf[:12])
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
				conn.Write(buf[:8])
				continue
			}
			if readed == 0 {
				binary.BigEndian.PutUint32(buf[0:4], 4)
				binary.BigEndian.PutUint32(buf[4:8], 0)
				conn.Write(buf[:8])
			} else if readed > 0 {
				binary.BigEndian.PutUint32(buf[0:4], 4+uint32(readed))
				binary.BigEndian.PutUint32(buf[4:8], uint32(readed))
				copy(buf[8:], readbuf)
				conn.Write(buf[:8+readed])
			}
		case WRITE:
			findex := binary.BigEndian.Uint32(msgbuf[:4])
			offset := binary.BigEndian.Uint32(msgbuf[4:8])
			size := binary.BigEndian.Uint32(msgbuf[8:12])
			fmt.Println(size)
			fmt.Println(len(msgbuf))
			wbuf := msgbuf[12 : 12+size]
			f := openedFile[uintptr(findex)]
			writed, _ := f.WriteAt(wbuf, int64(offset))
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], uint32(writed))
			conn.Write(buf[:8])
		case READDIR:
			fixpath := rootdir + string(msgbuf)
			fileList := []byte{}
			err = filepath.Walk(fixpath, func(path string, f os.FileInfo, err error) error {
				// rp, _ := filepath.Rel(rootdir, path)
				rp, _ := filepath.Rel(fixpath, path)
				fileList = append(fileList, rp...)
				fileList = append(fileList, 0)
				return nil
			})
			binary.BigEndian.PutUint32(buf[:4], uint32(len(fileList))+4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			copy(buf[8:], fileList)
			conn.Write(buf[:len(fileList)+8])
		case RELEASE:
			findex := binary.BigEndian.Uint32(msgbuf[:4])
			f := openedFile[uintptr(findex)]
			err := f.Close()
			if err != nil {
				fmt.Println(err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -9
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			binary.BigEndian.PutUint32(buf[0:4], 4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			conn.Write(buf[:8])
		case CREATE:
			// mode := binary.BigEndian.Uint32([]byte(matched[1])[0:4])
			fixpath := rootdir + string(msgbuf[4:])
			f, err := os.Create(fixpath)
			if err != nil {
				log.Print("Can't create", err)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				ret := -13
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf[:8])
				continue
			}
			fd := f.Fd()
			openedFile[fd] = f
			binary.BigEndian.PutUint32(buf[0:4], 8)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			binary.BigEndian.PutUint32(buf[8:12], uint32(fd))
			conn.Write(buf[:12])
		default:
			log.Print("bad tcfsOp: ", tcfsOp)
		}
	}
	fmt.Println("xxxxxxxxxxx, close")
}

var (
	port = flag.String("port", ":9876", "port to listen to")
)

func main() {
	flag.Parse()
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
		go handleConn(conn)
	}
}
