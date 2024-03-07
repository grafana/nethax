package cmd

import (
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/grafana/nethack/pkg"
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

func getPodForWorkload(cmd *cobra.Command) (*k8s.Pod, error) {
	podRegex, err := cmd.Flags().GetString("pod-from")
	if err != nil {
		fmt.Println("pod-from must be specified", err)
		return nil, err
	}

	namespace, err := cmd.Flags().GetString("namespace-from")
	if err != nil {
		fmt.Println("namespace-from must be specified", err)
		return nil, err
	}

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
	pod, err := k.Client.CoreV1().Pods(namespace).Get(cmd.Context(), matchPod, v1.GetOptions{})
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
	pod, _ := getPodForWorkload(cmd) // TODO
	command := []string{"nc"}

	// TODO --- fix error - unknown flag: --timeout
	// timeout, err := cmd.Flags().GetString("timeout")
	// if err != nil {
	// 	fmt.Println("timeout must be specified", err)
	// 	os.Exit(3)
	// }
	uri, err := cmd.Flags().GetString("remote-uri")
	if err != nil || uri == "" {
		fmt.Println("remote-uri must be specified", err)
		os.Exit(3)
	}
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		fmt.Println("remote-uri is invalid", err)
		os.Exit(3)
	}
	arguments := []string{"-w", "3", "-z", url.Host, getPortFromUrl(url)}
	netshootifiedPod, ephemeralContainerName, err := pkg.LaunchEphemeralContainer(pod, command, arguments)
	if err != nil {
		fmt.Println("Error launching ephemeral container.", err)
		os.Exit(3)
	}
	exitStatus := pkg.PollEphemeralContainerStatus(netshootifiedPod, ephemeralContainerName)
	os.Exit(int(exitStatus))
}
