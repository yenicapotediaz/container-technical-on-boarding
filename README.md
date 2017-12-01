# Samsung CNCT Technical On Boarding container
[![pipeline status](https://git.cnct.io/common-tools/samsung-cnct_container-technical-on-boarding/badges/master/pipeline.svg)](https://git.cnct.io/common-tools/samsung-cnct_container-technical-on-boarding/commits/master)
[![Docker Repository on Quay](https://quay.io/repository/samsung_cnct/technical-on-boarding-container/status "Docker Repository on Quay")](https://quay.io/repository/samsung_cnct/technical-on-boarding-container)

This repo contains a Dockerfile to build and test the technical on-boarding container.

This container is used by the chart-technical-on-boarding repo, and is part of the onboarding process for new Samsung CNCT employees.

## Goals

- Automate the preparation of a new CNCT employee's onboarding, as a list of tasks, represented as issues in GitHub.
- Provide a functional example of a Go program, integrating with the GitHub API.

### Functional Requirements

- Loads a template of "tasks" to be assigned to a new-hire. 
- Creates a Milestone and Project in GitHub. 
- Creates Issues in GitHub to represent tasks, and links them to Milestone and Project.
- Assigns those Issues to the new-hire.

## Usage

This application is designed to be hosted and used [here](http://technical-on-boarding.kubeme.io).
To run locally see the [Development and Testing](#development-and-testing) section below.

## Development and Testing

### Running in Container

Before you start you'll need:

- Docker
- Make
- A local environment file

To make an environment file that contains the Github client credentials you can copy the `template.env`
file like so:
```shell
cp template.env .env
```

To build and run the application in a container execute the following:

```shell
make -f Makefile.docker up
```

This will start a local HTTP server, at [127.0.0.1:9000](http://127.0.0.1:9000/). Open this URL
in your browser, click the *Authorize* button, and log into GitHub _as yourself_. Once authenticated, 
the application will direct you to the workload screen, start the task generateion job, and 
provide progress on the job.

To support [Revel's][4] _hot code reload_ in container, use the `run-dev` option to map 
your working directory into the container like so:

```shell
make -f Makefile.docker run-dev
```

#### Alternate listening IP addresses.

*NOTE*: In some Docker configurations this may listen on a different IP address.
Typically this is most easily found via:

```shell
env | grep DOCKER_HOST  # hopefully not empty...
```

Or potentially:

```shell
docker run --net=host codenvy/che-ip 
```

If either of the above indicate another IP address, try reaching that on port 9000.

### Running Locally (in your desktop)

If somehow you prefer to build and run this in Go on your desktop environment,
rather than a preconfigured Docker container, here's how...

Please ensure that [Go is properly set up](./SETTINGUPGO.md) first.

```shell
make setup   # installs dependencies, etc
make test    # runs Golang unit tests/etc
make build   # prepares the web app

# app credentials required
ONBOARD_CLIENT_ID="{clientid}" \ 
    ONBOARD_CLIENT_SECRET="{clientsecret}" \
    ONBOARD_REPO="technical-on-boarding" \
    ONBOARD_ORG="samsung-cnct" \
    VERSION="1.1.0" \
    BUILD="local" \
    revel run github.com/samsung-cnct/technical-on-boarding/onboarding
```

This workload relies heavily on the GitHub API, which also requires valid appliation tokens.

To facilitate testing this projects includes a fairly robust mock of the GitHub API client, and relies on
interfaces and proxy methods in several other points to allow the business logic to operate against a local
testing environment without reaching GitHub's API service.

Per golang's convention, tests are found in files ending with `_test.go`.

[2]: https://github.com/settings/applications/new
[3]: https://github.com/settings/apps
[4]: https://revel.github.io/
