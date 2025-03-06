<p align=center>
    <img src="logo.jpg" />
</p>

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
- install docker or colima in order to run docker-compose dev environment
- run the dev environment using `docker compose up` or `docker-compose up`
- run `go run cmd/server/main.go`
- start using api: e.g. create your repository connection `curl -X POST -v http://localhost:8000/connect -d '{"url": "http://github.com/whatever/youwant"}'`

## Internal tech

### CDK

Treenq CDK is responsible for defining infrastracture setup given from the user's `space` state.
It uses only postgres as a dependency. However, any other persistent storage can be implemented as a infra state store.

When the desired infrastracture change is defined the following behaviour is expected.

1. It creates a lock in order to allow infra update to only single process
2. It opens an "open" record to prepare a resource creation
3. It creates the infra resource
   1. On fail it updates the record to status "reverted" and returns the defined error
4. It updates the record to status "done"
5. It unlocks the lock
