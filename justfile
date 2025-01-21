build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/all-in-one-model-context-protocol ./main.go

docs:
  go run scripts/docs/update-doc.go

scan:
  trufflehog git file://. --only-verified

install:
  go install ./...
