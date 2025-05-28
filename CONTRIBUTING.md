# Developer Setup

Ensure you have `go`, `kubectl`, `kind`, and some kind of `docker` command installed.

You can initialize a new kind cluster for testing and run the example test plan against it using this command:
```Console
# make kind-init-oteldemo run-example-test-plan
```

This make target will create new containers of the runner and the probe and upload them to your kind cluster:
```Console
# make docker
```

## Running the OtelDemo Example Test Plan
The example test plan depends on the OtelDemo app, which is a comprehensive demonstration of a decently sized cloud-native application.

The easiest way to install this is with `kubectl apply`:
https://opentelemetry.io/docs/demo/kubernetes-deployment/#install-using-kubectl

Once you've installed the OpenTelemetry Demo, you can execute the example test plan with the runner:
```
make run-example-test-plan
```
# Dependencies
Renovate is used for dependency management.

Some non-default things we do with Renovate:
- Run `god mod tidy` after updating Go modules
- Automerge is enabled -- PRs must be approved by a CODEOWNER and pass checks to be merged

See the Renovate config for the latest config. :)
