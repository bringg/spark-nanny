# spark-nanny

A simple app to monitor spark app driver pods and restart them in case of failure

## Prerequisites

- Helm >= 3

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
| Bringg DevOps | devops@bringg.com |  |
