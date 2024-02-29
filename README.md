# nethack
A toolkit for Kubernetes network engineers.

## Usage
For full usage instructions, please run:
```
nethack --help
```

To launch the TUI, use:
```
nethack

nethack --context context-name
```

This will launch a TUI that displays namespaces. You can drill down from a namespace into a specific pod and then launch into these network debugging modes:
- pod-to-pod
- pod-to-remote
- pod iteractive
- tcpdump mode (live inspection of pod traffic)

Each of these options will launch an ephemeral container inside the pod and then perform the debugging action.

## CLI Usage (e.g. for CI/CD)
nethack can also be run in a non-interactive mode:

### Exit codes
In non-interactive mode, nethack will perform the test and then return an exit code.

Possible exit codes are:
```
exit 0 - success
exit 1 - connection not established
exit 2 - timeout exceeded
exit 3 - nethack error
... [TODO - clean this up]
```

### Global Options
```
    --timeout [SECONDS] - length of time to wait for establishing a TCP connection
    --context [NAME]    - connect to specified context name from kubeconfig
```

### pod-to-pod
```
# Test TCP communication from my-app deployment to other-app daemonset
nethack pod-to-pod \
        --namespace-from my-app --workload-from deployment/my-app \
        --namespace-to other-app --workload-to daemonset/other-app \
        --port 8080

# Test communication from my-app deployment to my-app statefulset
nethack pod-to-pod \
        --namespace-from my-app --workload-from deployment/my-app \
        --namespace-to my-app --workload-to statefulset/other-app \
        --remote-port 8080
```

### pod-to-remote
Test communication from a workload to a remote TCP destination.
```
# Test communication from my-app deployment to some HTTP(S) endpoint
nethack pod-to-remote \
        --namespace-from my-app --workload-from deployment/my-app \
        --remote-uri https://grafana.com

# Test communication from my-app deployment to some MySQL database
nethack pod-to-remote \
        --namespace-from my-app --workload-from deployment/my-app \
        --remote-uri mysql://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples

# Test communication from my-app deployment to some Postgres database
nethack pod-to-remote \
        --namespace-from my-app --workload-from deployment/my-app \
        --remote-uri postgres://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples

# Test communication from my-app deployment to some Redis database
nethack pod-to-remote \
        --namespace-from my-app --workload-from deployment/my-app \
        --remote-uri redis://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples
```

Note: the test does not handle layer 7 protocols at all, it only establishes a TCP connection and verifies it. For example, it will not verify that authentication works to your databse, only that a connection can be made.

## Contributing
TODO - link to CONTRIBUTING.md

### Compiling locally
To compile locally:
```
make run
```

### Running tests
To run tests:
```
make test
```