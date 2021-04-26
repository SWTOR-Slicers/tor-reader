set GOOS=darwin
set GOARCH=amd64

go build -compiler gc -o tor-reader-mac

set GOOS=linux
go build -compiler gc -o tor-reader-linux

set GOOS=windows
go build -compiler gc -o tor-reader-windows.exe

