package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/grafana/nethax/pkg"
	"github.com/spf13/cobra"
	k8s "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Pod2Remote() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use: "pod2remote",
			Run: Pod2RemoteExec,
		},
	}

	cmd.Flags().String("namespace-from", "", "Namespace to test connections from.")
	cmd.Flags().String("pod-from", "", "Pod regex to test connections from. The first pod that matches the regex will be used.")
	cmd.Flags().String("remote-uri", "", "Remote URI to connect to. (SCHEME://[[USER:PASS]@]HOST[:PORT]])")

	return cmd
}

func getPodForWorkload(ctx context.Context, podRegex string, namespace string) (*k8s.Pod, error) {
	podNames := pkg.GetPods(namespace)
	var matchPod string
	for _, podName := range podNames {
		match, err := regexp.MatchString(podRegex, podName)
		if err != nil {
			fmt.Println("pod name regex invalid", err)
			return nil, err
		}

		if match {
			matchPod = podName
			break
		}
	}

	k := pkg.GetKubernetes()
	pod, err := k.Client.CoreV1().Pods(namespace).Get(ctx, matchPod, v1.GetOptions{})
	if err != nil {
		fmt.Println(err)
		return pod, err
	}

	return pod, nil
}

func parseFlags(cmd *cobra.Command, args []string) {
	err := cmd.Flags().Parse(args)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO validate flag usage
}

func getPortFromUrl(url *url.URL) string {
	if url.Port() != "" {
		return url.Port()
	} else if url.Scheme == "http" {
		return "80"
	} else if url.Scheme == "https" {
		return "443"
	} // TODO add more schemes

	return ""
}

func Pod2RemoteExec(cmd *cobra.Command, args []string) {
	parseFlags(cmd, args)
	podRegexFrom, err := cmd.Flags().GetString("pod-from")
	if err != nil {
		fmt.Println("--pod-from must be specified", "stacktrace:", err)
		return
	}
	namespaceFrom, err := cmd.Flags().GetString("namespace-from")
	if err != nil {
		fmt.Println("--namespace-from must be specified", "stacktrace:", err)
		return
	}
	uri, err := cmd.Flags().GetString("remote-uri")
	if err != nil || uri == "" {
		fmt.Println("--remote-uri must be specified", "stacktrace:", err)
		os.Exit(3)
	}
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		fmt.Println("--remote-uri is invalid", "stacktrace:", err)
		os.Exit(3)
	}
	timeout, err := cmd.Flags().GetInt("timeout")
	if err != nil {
		fmt.Println("--timeout is invalid", "stacktrace:", err)
		os.Exit(3)
	}
	expectFail, err := cmd.Flags().GetBool("expect-fail")
	if err != nil {
		fmt.Println("--expect-fail is invalid", "stacktrace:", err)
		os.Exit(3)
	}

	podFrom, _ := getPodForWorkload(cmd.Context(), podRegexFrom, namespaceFrom)

	// nc -w $timeout -z $host $port
	command := []string{"nc"}
	arguments := []string{"-w", strconv.Itoa(timeout), "-z", url.Host, getPortFromUrl(url)}
	netshootifiedPod, ephemeralContainerName, err := pkg.LaunchEphemeralContainer(podFrom, command, arguments)
	if err != nil {
		fmt.Println("--Error launching ephemeral container.", "stacktrace:", err)
		os.Exit(3)
	}

	exitStatus := pkg.PollEphemeralContainerStatus(netshootifiedPod, ephemeralContainerName)
	os.Exit(pkg.ExitNethax(int(exitStatus), expectFail))
}
