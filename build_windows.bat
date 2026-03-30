echo "building for Windows amd64" 

go clean modcache
go mod tidy
go mod verify
go build -ldflags="-s -w" -o Go-Cordance.exe .\cmd\game\main.go
go build -ldflags="-s -w" -o Go-Cordance-Editor.exe .\cmd\editor\main.go 
echo "build done"

