# spark-nanny

A simple app to monitor spark app driver pods and restart them in case of failure

![Version: 0.3.1](https://img.shields.io/badge/Version-0.3.1-informational?style=flat-square) ![AppVersion: v0.2.1](https://img.shields.io/badge/AppVersion-v0.2.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

## Prerequisites

- Helm >= 3

## Installing the chart

```shell
# add the spark-nanny repo

$ helm repo add bringg-spark-nanny https://bringg.github.io/spark-nanny

$ helm install spark-nanny bringg-spark-nanny/spark-nanny
```

This will create a release of `spark-nanny` in the default namespace. To install in a different one:

```shell
$ helm install -n spark my-release bringg-spark-nanny/spark-nanny
```

Note that `helm` will fail to install if the namespace doesn't exist. Either create the namespace beforehand or pass the `--create-namespace` flag to the `helm install` command.

## Uninstalling the chart

To uninstall `my-release`:

```shell
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod assignment |
| args | list | `[]` | List of additional args to pass, e.g. - --interval=45 - --debug - --dry-run |
| fullnameOverride | string | `""` | String to override release name |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"bringg/spark-nanny"` | Image repository |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Image pull secrets |
| listenPort | int | `9164` | Listen port for health checks and metrics, change this if you're changing the `listen-address` argument |
| nameOverride | string | `""` | String to partially override `spark-nanny.fullname` template (will maintain the release name) |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| podAnnotations | object | `{}` | Additional annotations to add to the pod |
| resources | object | `{}` | Pod resource requests and limits |
| serviceAccount.annotations | object | `{}` | Additional annotations to add to the service account |
| sparkApps | string | `""` | Comma separated list of spark apps to watch, e.g. spark-app1,spark-app2. Required |
| tolerations | list | `[]` | List of node taints to tolerate |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Bringg DevOps | <devops@bringg.com> |  |
