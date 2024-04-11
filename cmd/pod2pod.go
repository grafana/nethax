package cmd

import (
	"fmt"
	"os"

	"github.com/grafana/nethax/pkg"
	"github.com/spf13/cobra"
)

func Pod2Pod() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use: "pod2remote",
			Run: Pod2PodExec,
		},
	}

	cmd.Flags().String("namespace-from", "", "Namespace to test connections from.")
	cmd.Flags().String("pod-from", "", "Pod regex to test connections from. The first pod that matches the regex will be used.")
	cmd.Flags().String("namespace-to", "", "Namespace to test connections to.")
	cmd.Flags().String("pod-to", "", "Pod regex to test connections to. The first pod that matches the regex will be used.")
	cmd.Flags().String("port", "", "Target port to connect to.")

	return cmd
}

func Pod2PodExec(cmd *cobra.Command, args []string) {
	parseFlags(cmd, args)
	podRegexFrom, err := cmd.Flags().GetString("pod-from")
	if err != nil {
		fmt.Println("pod-from must be specified", err)
		return
	}

	namespaceFrom, err := cmd.Flags().GetString("namespace-from")
	if err != nil {
		fmt.Println("namespace-from must be specified", err)
		return
	}

	podFrom, _ := getPodForWorkload(cmd.Context(), podRegexFrom, namespaceFrom)
	command := []string{"nc"}

	port, err := cmd.Flags().GetString("port")
	if err != nil || port == "" {
		fmt.Println("--port must be specified", err)
		os.Exit(3)
	}

	podRegexTo, err := cmd.Flags().GetString("pod-to")
	if err != nil {
		fmt.Println("pod-to must be specified", err)
		return
	}

	namespaceTo, err := cmd.Flags().GetString("namespace-to")
	if err != nil {
		fmt.Println("namespace-to must be specified", err)
		return
	}

	podTo, _ := getPodForWorkload(cmd.Context(), podRegexTo, namespaceTo)

	arguments := []string{"-w", "3", "-z", podTo.Status.PodIP, port}
	netshootifiedPod, ephemeralContainerName, err := pkg.LaunchEphemeralContainer(podFrom, command, arguments)
	if err != nil {
		fmt.Println("Error launching ephemeral container.", err)
		os.Exit(3)
	}
	exitStatus := pkg.PollEphemeralContainerStatus(netshootifiedPod, ephemeralContainerName)
	os.Exit(int(exitStatus))
}
