<p align=center>
    <img src="logo.jpg" />
</p>

# treenq

An open-source Platform to simplify app delivery.

## Demo

TBD

> ## 🚧 **Pre-Alpha Notice**: _The project is in pre-alpha stage.

Early-stage teams need:

- **Fast & affordable build** – Build and deploy apps with no docker registry/build step.
- **Secrets & configurations** – Manage credentials and settings effortlessly.

Treenq solves this problem.

## Documentation

TBD, expect it [there](https://treenq.com/started/get-started)

## Motivation

Most PaaS solutions are closed-source, costly, and lock you into their ecosystem.
Treenq offers a fully open-source alternative, giving you full control over your app management and infrastructure.

Treenq is designed to prioritize Application Management over infrastructure concerns.

## Contributor guide

⭐ **Enjoying Treenq?**
Support the project by giving it a star on GitHub! Your support helps us grow!

Or follow updates in our [discord](https://discord.gg/4HB3vEnYW9)

### How to contribute

📢 **Want to contribute?**
We welcome all contributions! Before submitting code, please open an issue to discuss your changes.

#### Run e2e tests

Run `make run-e2e-tests` or if you already run local setup in docker-compose then `go test -v -count=1 -race ./e2e/...`.

Alternatively you can run your dev setup for e2e tests manually.

#### Docker setup explained

## Roadmap

Below I shortly present the goals of the project in a its stages of development

* Pre Alpha version (we are here)
  - [ ] CLI as a UI
  - [ ] Single VM app deployment only
  - [ ] deploy right from the app source folder
  - [ ] docker as app runtime
  - [ ] build logs available
  - [ ] deploy history
  - [ ] rollback to any point in history is available
  - [ ] no docker registry required to deploy a container
  - [ ] domain and TLS via zero configuration
  - [ ] infra state is saved locally, no additional backend installation on the VM to consume CPU/RAM
* Alpha, complete stack to deploy home labs
  - [ ] Multi nodes deployment, a service per node
  - [ ] Secrets API
  - [ ] Remote infra state: s3
* Beta, stack becomes observable
  - [ ] Logging collection
  - [ ] Metrics collection
  - [ ] Computation resource monitoring
  - [ ] Server terminal
* Release candidate 1.0, stack becomes reliable
  - [ ] Disk backup to a configured driver: s3
  - [ ] Health check / failover
  - [ ] Replicas scaling
  - [ ] Zero downtime / rolling update
  - [ ] Resource management quotas
  - [ ] Same nodes replication
  - [ ] Cross nodes replication
  - [ ] Nodes selectors
  - [ ] Inner service networking
* Release 1.0, stack becomes team suitable
  - [ ] Central control plane is available, public API
  - [ ] Team, orgs, etc.
  - [ ] Github orgs blueprint
  - [ ] Deploy on Github events
  - [ ] Deploy from your CI
  - [ ] Deploy from remote docker registry
  - [ ] Buildpacks / railpacks
  - [ ] Environments
* Post 1.0, dev friendly, cloud available
  - [ ] Install apps from templates
  - [ ] Mirror service traffic to another target
  - [ ] Branch preview
  - [ ] Deploy from local changes instantly
  - [ ] Infra boot: prepare VMs for apps
  - [ ] Infra provision, VPC configuration, inner networking API
  - [ ] Multi tenancy
  - [ ] Track cost
