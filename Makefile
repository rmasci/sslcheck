compilerFlag=-gcflags=-trimpath=$(shell pwd) -asmflags=-trimpath=$(shell pwd)
goFiles=main.go
progName=sslcheck
ver=$(shell date +"BuildDate:_%Y%m%d-%H:%M")
all: mac linux
$(progName)l: mac linux 

windows:$(goFiles)
	GOOS=windows GOARCH=386  go build $(compilerFlag) -ldflags="-X main.version=\"$(ver)\""  -o binaries/win32/$(progName) $(goFiles)
	GOOS=windows GOARCH=amd64  go build $(compilerFlag) -ldflags="-X main.version=\"$(ver)\""  -o binaries/win64/$(progName) $(goFiles)


linux: $(goFiles)
	GOOS=linux GOARCH=386  go build $(compilerFlag) -ldflags="-X main.version=\"$(ver)\""  -o binaries/linux32/$(progName) $(goFiles)
	GOOS=linux GOARCH=amd64  go build $(compilerFlag) -ldflags="-X main.version=\"$(ver)\""  -o binaries/linux/$(progName) $(goFiles)

mac: $(goFiles)
	GOOS=darwin GOARCH=amd64  go build $(compilerFlag) -ldflags="-X main.version=\"$(ver)\"" -o binaries/mac/$(progName) $(goFiles)

