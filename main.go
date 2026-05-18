// Command kubectl-crd-sample is a kubectl plugin that generates an example
// YAML manifest for a CRD.
//
//	kubectl crd-sample <crd-name>
//
// Standard kubectl flags such as --kubeconfig, --context and --namespace are
// honoured via k8s.io/cli-runtime/pkg/genericclioptions.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/hegerdes/kubectl-crd-sample/internal/client"
	"github.com/hegerdes/kubectl-crd-sample/internal/generator"
)

// Build-time metadata. Populated by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const usage = `Usage:
  kubectl crd-sample <crd-name> [flags]

Generates an example YAML manifest from the storage version of the named
CustomResourceDefinition. The manifest is annotated with comments describing
each field (description, required/optional, allowed enum values).

Examples:
  kubectl crd-sample certificates.cert-manager.io
  kubectl crd-sample widgets.example.com --context staging > widget.yaml

Flags:
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "kubectl-crd-sample:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr *os.File) error {
	flags := pflag.NewFlagSet("kubectl-crd-sample", pflag.ContinueOnError)
	flags.SetOutput(stderr)

	kubeFlags := genericclioptions.NewConfigFlags(true)
	kubeFlags.AddFlags(flags)

	showHelp := flags.BoolP("help", "h", false, "show usage and exit")
	showVersion := flags.Bool("version", false, "print version information and exit")

	flags.Usage = func() {
		fmt.Fprint(stderr, usage)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}
		return err
	}
	if *showHelp {
		flags.Usage()
		return nil
	}
	if *showVersion {
		fmt.Fprintf(stdout, "kubectl-crd-sample %s (commit %s, built %s)\n", version, commit, date)
		return nil
	}

	positional := flags.Args()
	if len(positional) != 1 {
		flags.Usage()
		return fmt.Errorf("exactly one positional argument (CRD name) is required, got %d", len(positional))
	}
	crdName := positional[0]

	ctx := context.Background()
	crd, err := client.FetchCRD(ctx, kubeFlags, crdName)
	if err != nil {
		return fmt.Errorf("fetching CRD %q: %w", crdName, err)
	}

	out, err := generator.Generate(crd)
	if err != nil {
		return fmt.Errorf("generating manifest: %w", err)
	}

	if _, err := stdout.Write(out); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return nil
}
