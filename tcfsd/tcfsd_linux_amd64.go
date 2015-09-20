package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"syscall"
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
