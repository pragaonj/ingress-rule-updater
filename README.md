# ingress-rule

A `kubectl` plugin to add/remove ingress rules on the fly.

## Description

Add/remove kubernetes ingress rules via command line.
`ingress-rule` allows the configuration of an ingress resource with command line arguments.  

When adding/deleting a backend rule the ingress will be updated.
On creation of a rule for a non-existing ingress name a new ingress will be created.
If the last rule is deleted the ingress will be deleted as well.

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
kubectl ingress-rule set my-ingress --service foo --port 80
kubectl ingress-rule set my-ingress --service foo --port 80 --host *.foo.com --namespace default
kubectl ingress-rule set my-ingress --service foo --port 80 --host foo.com --path /foo

# remove a rule
kubectl ingress-rule delete my-ingress --service foo
kubectl ingress-rule delete my-ingress --service foo --port 80
```
