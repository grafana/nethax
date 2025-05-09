package main

import (
	"strings"
	"testing"
)

const exampleTestPlan = `testPlan:
  name: "Otel Demo Test Plan"
  description: "Test plan for the opentelemetry demo application: https://github.com/open-telemetry/opentelemetry-demo"
  testTargets:
  - name: "basic connectivity tests"
    podSelector: "app.kubernetes.io/instance=opentelemetry-demo"
    podSelection:
      mode: "random"  # options: "all", "random"
    tests:
    - name: "Check internet access"
      endpoint: "https://grafana.com"
      statusCode: 200
      timeout: 5s
    - name: "Check internal service"
      endpoint: "http://otel-collector.otel-demo.svc.cluster.local:8888/metrics"
      statusCode: 200
      timeout: 3s
  - name: "coredns connectivity tests"
    podSelector: "k8s-app=kube-dns"
    namespace: kube-system
    podSelection:
      mode: "all"  # will test all pods matching the selector and namespace
    tests:
    - name: "Ensure fake service call fails"
      endpoint: "http://fake-service.fake.svc.cluster.local/fake/healthz"
      statusCode: 0 # 0 expects a connection failure; useful to test network policies
      timeout: 3s
    - name: "TCP service call"
      endpoint: "fake-service.fake.svc.cluster.local:9001"
      type: tcp
      expectFail: true # TCP connection failure is expected; useful to test network policies
      timeout: 3s
`

func TestParseTestPlan(t *testing.T) {
	r := strings.NewReader(exampleTestPlan)

	tp, err := ParseTestPlan(r)
	if err != nil {
		t.Fatal(err)
	}

	if e, g := 2, len(tp.TestTargets); e != g {
		t.Fatalf("expecting %d test targets, got %d", e, g)
	}

	t.Run("default to HTTP(S)", func(t *testing.T) {
		tt := tp.TestTargets[0]

		for i, tt := range tt.Tests {
			if tt.Type != "HTTP(S)" {
				t.Errorf("expecting test target 0, test %d to be of type HTTP(S), got %q", i, tt.Type)
			}
		}
	})

	t.Run("don't override type", func(t *testing.T) {
		tt := tp.TestTargets[1].Tests[1]

		if e, g := "tcp", tt.Type; e != g {
			t.Fatalf("expecting type to be %q, got %q", e, g)
		}
	})
}
