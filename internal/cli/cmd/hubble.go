// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cmd

import (
	"context"
	"os"

	"github.com/cilium/cilium-cli/defaults"
	"github.com/cilium/cilium-cli/hubble"

	"github.com/spf13/cobra"
)

func newCmdHubble() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hubble",
		Short: "Hubble observability",
		Long:  ``,
	}

	cmd.AddCommand(
		newCmdPortForwardCommand(),
		newCmdUI(),
	)
	if os.Getenv("CILIUM_CLI_MODE") == "helm" {
		cmd.AddCommand(
			newCmdHubbleEnableWithHelm(),
			newCmdHubbleDisableWithHelm(),
		)
	} else {
		cmd.AddCommand(
			newCmdHubbleEnable(),
			newCmdHubbleDisable(),
		)
	}
	return cmd
}

func newCmdHubbleEnable() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable Hubble observability",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Namespace = namespace
			ctx := context.Background()
			h, err := hubble.NewK8sHubble(ctx, k8sClient, params)
			if err != nil {
				fatalf("Unable to enable Hubble: %s", err)
			}
			if err := h.Enable(ctx); err != nil {
				fatalf("Unable to enable Hubble: %s", err)
			}
			return nil
		},
	}

	addCommonHubbleEnableFlags(cmd, &params)
	// It can be deprecated since we have a helm option for it
	cmd.Flags().StringVar(&params.RelayImage, "relay-image", "", "Image path to use for Relay")
	// It can be deprecated since we have a helm option for it
	cmd.Flags().StringVar(&params.RelayVersion, "relay-version", "", "Version of Relay to deploy")
	// It can be deprecated since there is not a helm option for it and
	cmd.Flags().StringVar(&params.RelayServiceType, "relay-service-type", "ClusterIP", "Type of Kubernetes service to expose Hubble Relay")
	cmd.Flags().MarkDeprecated("relay-service-type", "value is no longer used for relay-service")

	// It can be deprecated since we have a helm option for it
	cmd.Flags().StringVar(&params.UIImage, "ui-image", "", "Image path to use for UI")
	// It can be deprecated since we have a helm option for it
	cmd.Flags().StringVar(&params.UIBackendImage, "ui-backend-image", "", "Image path to use for UI backend")
	// It can be deprecated since we have a helm option for it
	cmd.Flags().StringVar(&params.UIVersion, "ui-version", "", "Version of UI to deploy")
	cmd.Flags().BoolVar(&params.CreateCA, "create-ca", true, "Automatically create CA if needed")
	cmd.Flags().BoolVar(&params.Wait, "wait", true, "Wait for status to report success (no errors)")
	cmd.Flags().DurationVar(&params.WaitDuration, "wait-duration", defaults.StatusWaitDuration, "Maximum time to wait for status")

	cmd.Flags().StringVar(&params.K8sVersion, "k8s-version", "", "Kubernetes server version in case auto-detection fails")
	cmd.Flags().StringVar(&params.HelmChartDirectory, "chart-directory", "", "Helm chart directory")
	cmd.Flags().StringSliceVar(&params.HelmOpts.ValueFiles, "helm-values", []string{}, "Specify helm values in a YAML file or a URL (can specify multiple)")
	cmd.Flags().StringArrayVar(&params.HelmOpts.Values, "helm-set", []string{}, "Set helm values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&params.HelmOpts.StringValues, "helm-set-string", []string{}, "Set helm STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&params.HelmOpts.FileValues, "helm-set-file", []string{}, "Set helm values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	cmd.Flags().StringVar(&params.HelmGenValuesFile, "helm-auto-gen-values", "", "Write an auto-generated helm values into this file")
	cmd.Flags().StringVar(&params.HelmValuesSecretName, "helm-values-secret-name", defaults.HelmValuesSecretName, "Secret name to store the auto-generated helm values file. The namespace is the same as where Cilium will be installed")
	cmd.Flags().BoolVar(&params.RedactHelmCertKeys, "redact-helm-certificate-keys", true, "Do not print in the terminal any certificate keys generated by helm. (Certificates will always be stored unredacted in the secret defined by 'helm-values-secret-name')")

	for flagName := range hubble.FlagsToHelmOpts {
		// TODO(aanm) Do not mark the flags has deprecated for now.
		// msg := fmt.Sprintf("use --helm-set=%s<=value> instead", helmOpt)
		// err := cmd.Flags().MarkDeprecated(flagName, msg)
		// if err != nil {
		// 	panic(err)
		// }
		hubble.FlagValues[flagName] = cmd.Flags().Lookup(flagName).Value
	}

	return cmd
}

func newCmdHubbleDisable() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable Hubble observability",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Namespace = namespace
			ctx := context.Background()

			h, err := hubble.NewK8sHubble(ctx, k8sClient, params)
			if err != nil {
				fatalf("Unable to disable Hubble:  %s", err)
			}
			if err := h.Disable(ctx, false); err != nil {
				fatalf("Unable to disable Hubble:  %s", err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&params.RedactHelmCertKeys, "redact-helm-certificate-keys", true, "Do not print in the terminal any certificate keys generated by helm. (Certificates will always be stored unredacted in the secret defined by 'helm-values-secret-name')")
	cmd.Flags().StringVar(&params.HelmChartDirectory, "chart-directory", "", "Helm chart directory")
	cmd.Flags().StringVar(&params.HelmValuesSecretName, "helm-values-secret-name", defaults.HelmValuesSecretName, "Secret name to store the auto-generated helm values file. The namespace is the same as where Cilium will be installed")

	return cmd
}

func newCmdPortForwardCommand() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "port-forward",
		Short: "Forward the relay port to the local machine",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Context = contextName
			params.Namespace = namespace
			ctx := context.Background()

			if err := params.RelayPortForwardCommand(ctx, k8sClient); err != nil {
				fatalf("Unable to port forward: %s", err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&params.RedactHelmCertKeys, "redact-helm-certificate-keys", true, "Do not print in the terminal any certificate keys generated by helm. (Certificates will always be stored unredacted in the secret defined by 'helm-values-secret-name')")
	cmd.Flags().StringVar(&params.HelmValuesSecretName, "helm-values-secret-name", defaults.HelmValuesSecretName, "Secret name to store the auto-generated helm values file. The namespace is the same as where Cilium will be installed")
	cmd.Flags().IntVar(&params.PortForward, "port-forward", 4245, "Local port to forward to")

	return cmd
}

func newCmdUI() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Open the Hubble UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Context = contextName
			params.Namespace = namespace

			if err := params.UIPortForwardCommand(context.Background()); err != nil {
				fatalf("Unable to port forward: %s", err)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&params.UIPortForward, "port-forward", 12000, "Local port to use for the port forward")
	cmd.Flags().BoolVar(&params.UIOpenBrowser, "open-browser", true, "When --open-browser=false is supplied, cilium Hubble UI will not open the browser")

	return cmd
}

// addCommonUninstallFlags adds uninstall command flags that are shared between classic and helm mode.
func addCommonHubbleEnableFlags(cmd *cobra.Command, params *hubble.Parameters) {
	cmd.Flags().BoolVar(&params.Relay, "relay", true, "Deploy Hubble Relay")
	cmd.Flags().BoolVar(&params.UI, "ui", false, "Enable Hubble UI")
}

func newCmdHubbleEnableWithHelm() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable Hubble observability using Helm",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Namespace = namespace
			ctx := context.Background()
			if err := hubble.EnableWithHelm(ctx, k8sClient, params); err != nil {
				fatalf("Unable to enable Hubble: %s", err)
			}
			return nil
		},
	}

	addCommonHubbleEnableFlags(cmd, &params)
	return cmd
}

func newCmdHubbleDisableWithHelm() *cobra.Command {
	var params = hubble.Parameters{
		Writer: os.Stdout,
	}

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable Hubble observability using Helm",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			params.Namespace = namespace
			ctx := context.Background()
			if err := hubble.DisableWithHelm(ctx, k8sClient, params); err != nil {
				fatalf("Unable to disable Hubble:  %s", err)
			}
			return nil
		},
	}

	return cmd
}
