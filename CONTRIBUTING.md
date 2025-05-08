# Developer Setup

Ensure you have `kubectl`, `kind`, and some kind of `docker` command installed.

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

## TODO: Implement Kyverno E2E Testing
This would automate the test process to make developer onboarding easier.
