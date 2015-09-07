export GOPATH := $(PWD)

all:: bin/tcfsd

bin/tcfsd: src/tcfsd/*.go src/tcfs/*.go
	go install tcfsd

clean::
	rm -rf bin/*
