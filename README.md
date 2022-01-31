# ingress-rule kubectl

A `kubectl` plugin to update/add/remove ingress rules on the fly.

## Quick Start

```
kubectl krew install ingress-rule
kubectl ingress-rule
```

## Usage

```bash
go run cmd/plugin/main.go --port 80 --add --service-name foo --namespace default --ingress-name foo --host foo.example.com

go run cmd/plugin/main.go set foo --service foo --port 80
go run cmd/plugin/main.go set foo --service foo --port 80 --host *.foo.com --namespace default

```

## Command line options

```
kubectl ingress-rule <command>

Commands:
set <ingress-name>      Adds a backend rule to the ingress
delete <ingress-name>   Deletes a backend rule from the ingress

Options:
-p, --port              Set backend service port by port number
-s, --service           Set backend service by name
--host                  Set host (optional)
--path                  Set path (optional)  
--path-type             Set matching type for path (optional); Accepts: "Prefix", "Exact", "ImplementationSpecific"; Defaults to "Prefix"

From kubectl inherited options:
-n, --namespace         Set the namespace
```

## Feature plans

- Allow user to specify a path (additionally to the host)
- Optional port configuration via PortName instead of PortNumber
- Extract namespace from context if no namespace options is provided
- Delete backend-rules by <service>
- Delete backend-rules by <service>:<port>
- Delete backend-rules by <host>
- Delete backend-rules by <host><path>
- Add option to configure ingressClassName on ingress creation
