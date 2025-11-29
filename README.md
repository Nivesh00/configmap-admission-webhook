#  Configmap Admission Webhook

Webhook is used to check keys in a configmap and can either reject the configmap or edit it to remove the forbidden keys.

Since configmaps are stored on disk, it is not recommended to store sensitive data like api keys, passwords etc. in them. These kind of data are better store in a secret, which are stored in memory.

> [!CAUTION]
> This repo is more a learning project that anything and is not intended to be actually used e.g. in production etc.

## Webhook Server

The webhook server takes 2 arguments:
- `--log-level` which is the log level. Values are [debug,info,warn,error] (default is warn)
- `--port` which is the port used by the server to listen to requests (default is 443)

The server uses environmental variables to define its behavior:

| Key  | Value  | Notes  |
|:---:|:---:|:---:|
| `FORBIDDEN_KEYS` (required)  |  Comma seperated values of type string | Webhook server looks for keys matching the forbidden keys in a configmap in the path `.data`. Users should make sure extra new lines are removed | 
| `POLICY`  | `AUTO` or `MANUAL` (default)  | Defines the behavior when the webhook server discovers a forbidden key in the configmap. `AUTO` removes keys from configmap and `MANUAL` rejects the configmap  |
| `CASE_SENSITIVE`  | `true` (default) or `false` | Checks whether upper- and lowercase letters are important for forbidden keys check  |

Example of a configmap for these values which can be mounted to the webhook pod:

```yml
# Configurations for webhook
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-conf
data:
  FORBIDDEN_KEYS: >-
    API_KEY, PASSWORD
  POLICY: AUTO
  CASE_SENSITIVE: "false"
```

### Webhook Server Docker Image

The docker image can be found in this repository and can be pulled using following command:
```dh
docker pull ghcr.io/nivesh00/configmap-admission-webhook:latest
```

## SSL

Note that SSL is managed using the Cert Manager CRDs for easier use, so the SSL certificates should be mounted under:
- `/etc/certs/tls.key`
- `/etc/certs/tls.crt`

Cert Manager can be installed suing this command:
```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.yaml
```

The certificates for the webhook configurations are then mounted using the certificate resource, read the Cert Manager docs for more

https://cert-manager.io/docs/concepts/ca-injector/#injecting-ca-data-from-a-certificate-resource


## Manifests Files
> [!WARNING]
> Manifest files are not yet fully created

Manifests files can be found in the `kustomize` directory. The manifest files are not yet created but test manifest files can be found in same directory
