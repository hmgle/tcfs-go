# TCFS-GO

TCFS, a Lightweight Network File System.

Tcfs-go is the tcfs server writen in Golang.

## How to use?

- Server(assume IP address is 192.168.0.100):

```
go get github.com/hmgle/tcfs-go/tcfsd
tcfsd -dir .
# can use absolute path:
# tcfsd -dir "/tmp"
# tcfsd -dir "c:" # Windows
```

- Client:

```
sudo apt-get install libfuse-dev
git clone https://github.com/tcfs/tcfs.git
cd tcfs
mkdir mountpoint
make
# 192.168.0.100 is the server IP address
./tcfs --server 192.168.0.100 mountpoint
ls -shal mountpoint
# access mountpoint
# ...
# unmount tcfs
fusermount -u mountpoint
# or sudo umount mountpoint
```

## Protocol

[TCFS protocol](https://github.com/tcfs/tcfs/blob/master/protocol.adoc)
