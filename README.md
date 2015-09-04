# TCFS-GO

TCFS, a Lightweight Network File System.

Tcfs-go is the tcfs server writen in Golang.

## How to use?

- Server:

```
git clone https://github.com/hmgle/tcfs-go.git
cd tcfs-go
make
bin/tcfsd -dir .
```

- Client:

```
sudo apt-get install libfuse-dev
git clone https://github.com/hmgle/tcfs.git
cd tcfs
mkdir mountpoint
make
./tcfs --server 127.0.0.1 mountpoint
ls -shal mountpoint
```
