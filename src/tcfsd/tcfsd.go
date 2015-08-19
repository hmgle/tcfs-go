package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

func handleConn(conn net.Conn) {
	var err error
	buf := make([]byte, 4096*1024) // 4MB
	openedFile := map[uintptr]*os.File{}
	reGetattr := regexp.MustCompile("^getattr(.*)")
	reReaddir := regexp.MustCompile("^readdir(.*)")
	reOpen := regexp.MustCompile("^open(.*)")
	reRead := regexp.MustCompile("^read(.*)")
	reWrite := regexp.MustCompile("^write(.*)")
	reTruncate := regexp.MustCompile("^truncate(.*)")
	reRelease := regexp.MustCompile("^release(.*)")
	reMkdir := regexp.MustCompile("^mkdir(.*)")
	reRmdir := regexp.MustCompile("^rmdir(.*)")
	reUnlink := regexp.MustCompile("^unlink(.*)")
	reCreate := regexp.MustCompile("^create(.*)")
	reUtime := regexp.MustCompile("^utime(.*)")
	rootdir := "/home/gle/code_repo/cloud_lib/tcfs-go/rootdir"
	for {
		_, err = io.ReadFull(conn, buf[:4])
		if err != nil {
			break
		}
		msglen := binary.BigEndian.Uint32(buf[:4])
		_, err = io.ReadFull(conn, buf[:msglen])
		// println(string(buf))
		if matched := reGetattr.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
			fixpath := rootdir + matched[1]
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

		} else if matched := reReaddir.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
			fixpath := rootdir + matched[1]
			fileList := []byte{}
			err = filepath.Walk(fixpath, func(path string, f os.FileInfo, err error) error {
				rp, _ := filepath.Rel(rootdir, path)
				fileList = append(fileList, rp...)
				fileList = append(fileList, 0)
				return nil
			})
			fmt.Println(fileList)
			binary.BigEndian.PutUint32(buf[:4], uint32(len(fileList))+4)
			binary.BigEndian.PutUint32(buf[4:8], 0)
			copy(buf[8:], fileList)
			conn.Write(buf[:len(fileList)+8])
		} else if matched := reOpen.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
			flag := binary.BigEndian.Uint64([]byte(matched[1])[0:4])
			fixpath := rootdir + matched[1][4:]
			f, err := os.OpenFile(fixpath, int(flag), 0x666)
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
		} else if matched := reRead.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reWrite.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reTruncate.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reRelease.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reMkdir.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reRmdir.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reUnlink.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reCreate.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else if matched := reUtime.FindStringSubmatch(string(buf[:msglen])); len(matched) > 1 {
		} else {
		}
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
