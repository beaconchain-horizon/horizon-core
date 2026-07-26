git add go.mod go.sum
git commit -m "build: add go-github and oauth2 dependencies for webhook"
go run cmd/api/main.go
