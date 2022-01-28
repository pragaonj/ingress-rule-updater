# ingress-rule kubectl

A `kubectl` plugin to update/add/remove ingress rules on the fly.

## Quick Start

```
kubectl krew install ingress-rule
kubectl ingress-rule
```

## Usage
```
go run cmd/plugin/main.go --port 80 --add --service-name foo --namespace default --ingress-name foo --host foo.example.com



```


## Feature plans

- Allow user to specify a path (additionally to the host)
- Optional port configuration via PortName instead of PortNumber