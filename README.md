### treenq
It's a Platform as a Code to let you deliver, build, manage cloud resource and dependencies right from the code in order to:
- get rid of .env files
- give 0 buttons clicking on a platform
- provide easy debugging across the system
- testable infra
- pluggable system

## API Design
TBD

## How to run 
- install go https://go.dev/doc/install
- install sqlite (temporary solution to keep the persistent state) https://sqlite.org/download.html
- run `go run cmd/server/main.go`
- start using api: e.g. create your repository connection `curl -X POST -v http://localhost:8000/connect -d '{"url": "http://github.com/whatever/youwant"}'` 

test webhook
one more
check payload and save to fixtures
one more
didn't come
