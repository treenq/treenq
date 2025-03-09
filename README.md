<p align=center>
    <img src="logo.jpg" />
</p>

# treenq

Platform as a Code for Kubernetes to let you build, deliver, provision cloud applications and dependencies.

## Demo

TBD

# Treenq

### **The project is in pre-alpha stage, the release is planned in 2025 Q2.**

Treenq is a Platform as a Service (as Code) solves the infrastracture and app delivery complexity in Kubernetes.

On the very early stage every team needs:

- Fast and afordable app building, shortly saying basic CI with no code
- Database or another cloud resource provision
- Support secrets and configuration for the resources
- Quick preview for the first customers

This project is created to solve it.

And finally you can install an open source platform and use it for free in order to get:

- 🚀 App delivery using your Dockerfile/Containerfile or providing build/run commands
- 🌍 Get quick 3d level domain
- 🗄️ Request a database for your service, securely store its credentials and set them to the built app
- ⚖️ Request small fractional resources for your app

Any many more planed (environments, database branches, per api handler metrics, alerting, app delivery from local machine right to the testing environment, get env config file for specific environment to run your app locally, etc.)

## Documentation

TBD

## Motivation

There plenty of similar solutions and you know those names if you came here.
All of them provide very expensive proxy to aws with a fancy UI.

- very few tools are opensource
- some still give vendor lock to a specific SDK
- some require from you a lot of configuration or manual cloud resource management

This project is born to give application mangement instead of infra.
It's supposed to give some constrainst as well, but in case of selfhosting this solution you still have access to your cluster.

The worst case a team builds a complex Internal Developer Platform.
This project is far from replacing IDP, but rather give a fast delivery and lets focus on app delivery to engineers.

## Contributors guide

⭐ If you enjoyed the project's experience then giving us a start is your first contribution step.

### How to run

- install go https://go.dev/doc/install
- install docker/colima/podman in order to run docker-compose dev environment
- for mac: install macfuse `brew install macfuse` in order to support overlafs on the host machine to run the app inside a docker container
- run the dev environment using `docker compose up` or `docker-compose up`
- run `go run cmd/server/main.go`

### How to contribute

We are happy to see an issue created before submitting and code.
Any collaboration is welcome.
