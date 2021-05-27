# spark-nanny

Who watches the watchers?

`spark-nanny` is a simple app designed to monitor the health of `spark` apps installed using [spark-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator), and restart the driver pod in case it's unresponsive.

The main motivation comes from two main caveats in the way `spark` (and `spark-operator`) work when run on `kubernetes`:
* `Spark` doesn't have any configuration to provide health and readiness checks for pods
* The `spark-operator` doesn't provide any mechanism that supports this (e.g. by form of mutating webhook), see [here](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/issues/969)

While it is possible to use pod templates to launch `spark` apps (as of `spark 3.0.0`) and define `probes` there, this isn't supported by the `spark-operator` (yet, maybe in the future, see [this](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/issues/1176))

## Installation

The simplest way to install `spark-nanny` is to use the provided chart:

```shell
# Requires helm3

$ helm install spark-nanny --namespace spark charts/spark-nanny --set sparkApps="spark-app1\,spark-app2"
```

Note that the comma must be escaped (with `\`)

See the chart's `values` file for more details and possible configuration options

## Configuration

`spark-nanny` is configured using command line flags passed to the executable, the following flags are supported:

| key | default | description |
|---|---|---|
| `apps` | "" | comma separated list of `spark` app names to watch, e.g. `spark-app1,spark-app2` (required) |
| `interval` | `30` | time in seconds between checks |
| `timeout` | `10` | timeout in seconds to wait for a response from the driver pod |
| `namespace` | `spark` | `spark` apps namespace |
| `dry-run` | `false` | preforms all the checks and logic, but won't actually delete the pod |
| `debug` | `false` | set to `true` to enable more verbose logging |

## How it Works

`spark-nanny` does the following for each `spark` app passed via the `--apps` flag:
1. Get the pod ip from the kubernetes api server
2. Make sure the pod isn't in terminating phase, all containers are in `running` state and have been running for at least 60 seconds
3. Issue a `GET` request on the driver application endpoint `http://it<pod-ip>:4040/api/v1/applications`
4. If the request times out, the connection is refused or a non 200 status code is returned, retry 2 more times
5. If after 3 retries the driver pod still doesn't return a 200 status code, delete the pod
6. Rinse and repeat for every interval period defined

Once the driver pod is deleted, any executors owned by it will also be deleted and the `spark` app will be rescheduled be the operator

## Development

### Getting Started

> Requires go 1.16+

Clone the repo and run `make install-tools`, this will download the project dependencies and required tools.

Use `make build` to build locally and test. Note that because `spark-nanny` issues `http` requests to the driver pod, the pod needs to be accessible from `spark-nanny`

### Releasing a New Version

To release a new version of `spark-nanny` do the following:

1. Increment the `TAG` variable in the `Makefile`
2. Run `make push-image`
3. Update the `appVersion` field in `charts/spark-nanny/chart.yaml`
4. Deploy the updated version
