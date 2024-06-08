build:
	go run -ldflags="-X 'treenq/src/handlers.Version=unstable'" ./cmd/server/.