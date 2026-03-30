echo "building for linux arm64" 

go clean modcache
go get
go mod tidy
go mod verify
go build -ldflags="-s -w" -o go-cordance ./cmd/game/main.go
go build -ldflags="-s -w" -o go-cordance-editor ./cmd/editor/main.go 
echo "build done"

