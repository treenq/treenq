<p align=center>
    <img src="logo.jpg" />
</p>

# treenq

An open-source Platform as Code for Kubernetes to simplify app delivery, cloud resource provisioning, and infrastructure management.

## Demo

TBD

# Treenq

> ### 🚧 **Pre-Alpha Notice**: _The project is in pre-alpha stage, the alpha test planned in_ **2025 Q2**

Treenq is a Platform as a Service (as Code) solves the infrastructure and App Delivery complexity in Kubernetes.

Early-stage teams need:

- 🚀 **Fast & affordable CI/CD** – Build and deploy apps without complex pipelines.
- 🛠 **Database & cloud resources** – Provision services seamlessly.
- 🔐 **Secure secrets & configurations** – Manage credentials and settings effortlessly.
- ⚡ **Instant previews** – Get early feedback with temporary environments.

Treenq solves this problem.

And finally you can install an open source platform and use it for free in order to get:

- 🚀 **App Delivery** – Deploy from Dockerfile/Containerfile or build & run commands.
- 🌍 **Custom Domains** – Instantly get a third-level domain for your app.
- 🗄️ **Database Provisioning** – Securely provision and inject database credentials.
- ⚖️ **Optimized Resource Allocation** – Use minimal cloud resources efficiently.

And many more planned Features:

- 🏗 **Environment Management** – Easily configure dev, staging, and production.
- 📊 **API-Level Metrics & Alerts** – Track and optimize performance.
- 🔄 **Local-to-Cloud Deployment** – Push from your machine to test environments.
- 🔧 **Per-Environment Config Export** – Get configuration files to run apps locally.

## Documentation

TBD

## Motivation

Most PaaS solutions are closed-source, costly, and lock you into their ecosystem.
Treenq offers a fully open-source alternative, giving you full control over your app management and infrastructure.

Treenq is designed to prioritize Application Management over infrastructure concerns. While it provides helpful constraints, self-hosting allows full access to your Kubernetes cluster.

Many teams end up building complex **Internal Developer Platforms (IDPs)**.
Treenq isn’t meant to replace IDPs but to offer **fast, streamlined app delivery**, letting engineers focus on shipping code.

## Contributor guide

⭐ **Enjoying Treenq?**
Support the project by giving it a star on GitHub! Your support helps us grow!

### How to contribute

📢 **Want to contribute?**
We welcome all contributions! Before submitting code, please open an issue to discuss your changes.

### How to run locally

- Install [Go](https://go.dev/doc/install)
- Install Docker/Colima/Podman for running dev environment
- Mac users only: install macFUSE: `brew install macfuse`
- Run the dev environment: `make start-e2e-test-env`
- Attach remote debugger, here is example for vscode launch.json:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "go",
      "name": "debug remote service",
      "mode": "remote",
      "request": "attach",
      "port": 40000,
      "substitutePath": [
        {
          "from": "${env:HOME}/projects/treenq",
          "to": "/app"
        },
        {
          "from": "${env:HOME}/go/pkg/mod/",
          "to": "/go/pkg/mod/"
        }
      ]
    }
  ]
}
```

#### Run e2e tests

Run `make run-e2e-tests` or if you already run local setup in docker-compose then `go test -v -count=1 -race ./e2e/...`.

Alternatively you can run your dev setup for e2e tests manually.
Running a dev container locally for e2e tests require additional security options.
This tip will also help to run the service inside a container.

###### Option 1: podman

`podman run --device /dev/fuse:rw localhost/treenq`

###### Option 2: docker/colima

Docker is able running only a privileged container:
`docker run --privileged  -it treenq`

#### Docker setup explained

Dockerfile contains 4 staging:

- build stage, to actually gather all the dependencies
- build dev stage, it installs delve to be able to debug the app, compile without optimizations and starts
- build prod builds with compiler optimizations
- last stage runs it from prod build, not ready yes, but supposed to run the binary in alpine (scratch image doesn't seem possible)

docker-compose represents a base setup, directly isn't used, only linux users capable running the service locally and can use this setup
docker-compose.e2e can be used to run the setup locally, keep volumes and test it, also allows attaching a debugger
docker-compose.e2e-ci is used to run e2e tests, doesn't have volumes and debugger
docker-compose.staging is used to run a staging environment, doesn't include a database

### Run web locally

There are 2 commands to run web locally

- `npm run local` allows running a local web, it's supposed to have a backend locally as well
- `npm run dev` allows running a web only, it requires additional configuration described below

### Run locally web only

If you want to develop locally only frontend and not interested in running treenq backend locally it requires to use vite proxy server in order to support same site cookies.
Add web/.env.local file with the following content:

```sh
APP_API_HOST=http://localhost:9000/api
APP_GITHUB_APP_NAME=treenq-staging
USE_VITE_PROXY=true
```
