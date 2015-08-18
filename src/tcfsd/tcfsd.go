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
		buf := make([]byte, 4)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			break
		}
		msglen := binary.BigEndian.Uint32(buf)
		buf = make([]byte, msglen)
		_, err = io.ReadFull(conn, buf)
		// println(string(buf))
		if matched := reGetattr.FindStringSubmatch(string(buf)); len(matched) > 1 {
			fixpath := rootdir + matched[1]
			var stat syscall.Stat_t
			err = syscall.Lstat(fixpath, &stat)
			if err != nil {
				buf = make([]byte, 4+4)
				binary.BigEndian.PutUint32(buf[0:4], 4)
				var ret int32 = -2
				binary.BigEndian.PutUint32(buf[4:8], uint32(ret))
				conn.Write(buf)
				continue
			}
			buf = make([]byte, 4+11*4)
			binary.BigEndian.PutUint32(buf[0:4], 11*4)
			binary.BigEndian.PutUint32(buf[4:8], uint32(stat.Dev))
			binary.BigEndian.PutUint32(buf[8:12], uint32(stat.Ino))
			binary.BigEndian.PutUint32(buf[12:16], stat.Mode)
			binary.BigEndian.PutUint32(buf[16:20], uint32(stat.Nlink))
			binary.BigEndian.PutUint32(buf[20:24], stat.Uid)
			binary.BigEndian.PutUint32(buf[24:28], stat.Gid)
			binary.BigEndian.PutUint32(buf[28:32], uint32(stat.Size))
			binary.BigEndian.PutUint32(buf[32:36], uint32(stat.Atim.Sec))
			binary.BigEndian.PutUint32(buf[36:40], uint32(stat.Mtim.Sec))
			binary.BigEndian.PutUint32(buf[40:44], uint32(stat.Ctim.Sec))
			conn.Write(buf)

		} else if matched := reReaddir.FindStringSubmatch(string(buf)); len(matched) > 1 {
			fixpath := rootdir + matched[1]
			fileList := []string{}
			err = filepath.Walk(fixpath, func(path string, f os.FileInfo, err error) error {
				fileList = append(fileList, path)
				return nil
			})
			for _, file := range fileList {
				fmt.Println(file)
			}
		} else if matched := reOpen.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reRead.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reWrite.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reTruncate.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reRelease.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reMkdir.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reRmdir.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reUnlink.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reCreate.FindStringSubmatch(string(buf)); len(matched) > 1 {
		} else if matched := reUtime.FindStringSubmatch(string(buf)); len(matched) > 1 {
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
