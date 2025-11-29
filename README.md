# $${\color{white}Configmap \space Admission \space Webhook}$$

Webhook is used to check keys in a configmap and can either reject the configmap or edit it to remove the forbidden keys.

Since configmaps are stored on disk, it is not recommended to store sensitive data like api keys, passwords etc. in them. These kind of data are better store in a secret, which are stored in memory.

> [!CAUTION]
> This repo is more a learning project that anything and is not intended to be actually used e.g. in production etc.

## Webhook Server

The webhook server takes 2 arguments:
- `--log-level` which is the log level. Values are [debug,info,warn,error] (default is warn)
- `--port` which is the port used by the server to listen to requests (default is 443)

The server uses environmental variables to define its behavior:
- `FORBIDDEN_KEYS` is a string of comma-seperated values. The webhook will then look for keys matching the values in a configmap in the path `.data` (this is a required parameter)
  - users should pay special attention to remove any newline
- `POLICY` defines the bahavior when the webhook discovers a forbidden key in the configmap
  - `AUTO` removes the forbidden keys from the configmap
  - `MANUAL` rejects the configmap (this is the dafault value)
- `CASE_SENSITIVE` is set to `true` if the check should be case sensitive and `false` if not (default is true)

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


## manifests files
> [!WARNING]
> Manifest files are not yet fully created

Manifests files can be found in the `kustomize` directory. The manifest files are not yet created but test manifest files can be found in same directory
