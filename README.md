# nethax
A toolkit for Kubernetes network engineers.

## Usage
For full usage instructions, please run:
```
nethax --help
```

### Options
- pod-to-pod
- pod-to-remote
- (TODO) tcpdump mode (live inspection of pod traffic)

Each of these options will launch an ephemeral container inside the pod and then perform the debugging action.

### Exit codes

Possible exit codes are:
Nethack will perform the test and then return an exit code. Possible exit codes are:
```
exit 0 - success
exit 1 - connection not established
exit 2 - timeout exceeded
exit 3 - nethax error
... [TODO - clean this up]
```

### Global Options (TODO)
```
    --timeout [SECONDS] - length of time to wait for establishing a TCP connection
    --context [NAME]    - connect to specified context name from kubeconfig
```

### pod-to-pod
```
# Test TCP communication from my-app pod to other-app pod
nethax pod-to-pod \
        --namespace-from my-app --pod-from my-app \
        --namespace-to other-app --pod-to other-app \
        --port 8080
```

### pod-to-remote
Test communication from a pod to a remote TCP destination.
```
# Test communication from my-app pod to some HTTP(S) endpoint
nethax pod-to-remote \
        --namespace-from my-app --pod-from my-app \
        --remote-uri https://grafana.com

# Test communication from my-app pod to some MySQL database
nethax pod-to-remote \
        --namespace-from my-app --pod-from my-app \
        --remote-uri mysql://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples

# Test communication from my-app pod to some Postgres database
nethax pod-to-remote \
        --namespace-from my-app --pod-from my-app \
        --remote-uri postgres://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples

# Test communication from my-app pod to some Redis database
nethax pod-to-remote \
        --namespace-from my-app --pod-from my-app \
        --remote-uri redis://[DATABASE-DNS-OR-IP]:[PORT]
        # TODO - clean up these examples
```

Note: the test does not handle layer 7 protocols at all, it only establishes a TCP connection and verifies it. For example, it will not verify that authentication works to your databse, only that a connection can be made.

## Contributing
TODO - link to CONTRIBUTING.md

### Compiling locally
To compile locally:
```
make build
```

### Running tests
To run tests:
```
make test
```
