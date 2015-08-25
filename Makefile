export GOPATH := $(PWD)

all:: bin/tcfsd

bin/tcfsd: src/tcfsd/*.go 
	go install tcfsd

clean::
	rm -rf bin/*
