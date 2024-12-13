build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/all-in-one-model-context-protocol ./main.go

extract-env:
  go run scripts/extract-env.go