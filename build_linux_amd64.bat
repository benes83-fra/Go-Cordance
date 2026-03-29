echo "building for linux amd64" 

set GOOS=linux
set GOARCH=amd64
set GOROOT=C:\msys64\mingw64\lib\go
set GOPATH=C:\msys64\mingw64\
set CC=C:\msys64\ucrt64\bin\x86_64-w64-mingw32-gcc.exe

go clean modcache
go get
go mod tidy
go mod verify
go build -ldflags="-s -w" -o go-cordance .\cmd\game\main.go
go build -ldflags="-s -w" -o go-cordance-editor .\cmd\editor\main.go 
echo "build done"

