# ingress-rule

A `kubectl` plugin to add/remove ingress rules on the fly.

## Description

Add/remove kubernetes ingress rules via command line.
This `ingress-rule` allows the configuration of an ingress resource with command line arguments.
It can create, update (add/remove rules) and delete ingress resources as needed.

## Quick Start

```bash
kubectl krew install ingress-rule
kubectl ingress-rule
```

## Command line options

```
kubectl ingress-rule <command> <ingress-name>

Commands:
    set         Add kubernetes ingress rules via command line. If the ingress does not exist a new ingress will be created.
    delete      Remove kubernetes ingress rules via command line. Deletes the ingress if there are no rules left.

Options:
    --port      Set backend service port by port number
    --service   Set backend service by name
    --host                  Set host (optional)
    --path                  Set path (optional)  
    --path-type             Set matching type for path (optional); Accepts: "Prefix", "Exact", "ImplementationSpecific"; Defaults to "Prefix"

From kubectl inherited options:
    -n, --namespace         Set the namespace
```

## Usage examples

```bash
# add a rule
ingress-rule set my-ingress --service foo --port 80
ingress-rule set my-ingress --service foo --port 80 --host *.foo.com --namespace default
ingress-rule set my-ingress --service foo --port 80 --host foo.com --path /foo

# remove a rule
ingress-rule delete my-ingress --service foo
ingress-rule delete my-ingress --service foo --port 80
```

## Backlog

- Optional port configuration via PortName instead of PortNumber
- Delete backend-rules by <host>
- Delete backend-rules by <host><path>
- Add option to configure ingressClassName on ingress creation

## dev

```bash
# dev usage
go run cmd/ingress-rule/main.go set foo --service foo --port 80 --host *.foo.com --namespace default

```
