// Command auspex is the Auspex CLI entrypoint. Per Auspex_ADD.md
// §10.1, this package only does wiring and process exit — no business
// logic lives here or in Cobra command handlers. The real command tree
// (evaluate, decision, checkpoint, pause/resume/scheduler, status,
// doctor, hook claude ...) is composed in wire.go/adapters.go from every
// role's real, already-tested service implementation; newRootCmd below
// remains the minimal version-only fallback exercised directly by
// main_test.go.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/huaiche94/auspex/internal/buildinfo"
)

func main() {
	os.Exit(run())
}

// run holds every deferred cleanup inside its own stack frame so main can
// call os.Exit exactly once, after run (and its defers) have fully
// returned — os.Exit terminates immediately and never runs pending
// defers, so it must never appear in a function that itself defers
// cleanup.
func run() int {
	// Issue #88: cancel the root context on SIGINT/SIGTERM so a `kill
	// -TERM <auspex>` cancels every context-scoped operation — most
	// importantly the managed runner's exec.CommandContext child, which
	// was previously orphaned (terminal Ctrl-C signals the whole process
	// group, so only direct signals were affected). The runner's
	// context-cancel kill path is already tested; this wires the signal
	// into it.
	ctx, stop := rootContext()
	defer stop()

	root, closeFn, err := buildRootCmd(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auspex:", err)
		return 1
	}
	defer func() { _ = closeFn() }()

	if err := root.ExecuteContext(ctx); err != nil {
		return 1
	}
	return 0
}

// rootContext returns the process-lifetime context, cancelled on
// SIGINT/SIGTERM (issue #88). Factored out of run so the signal set is
// unit-testable — the #88 bug was precisely SIGTERM missing from it.
func rootContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

// newRootCmd builds the root Cobra command. Kept separate from main so
// tests can exercise command wiring without invoking os.Exit.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "auspex",
		Short:         "Auspex is a local-first predictive runtime guard for AI coding agents.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.AddCommand(newVersionCmd())

	return root
}

// newVersionCmd builds `auspex version`, which prints the build's
// version string to stdout.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Auspex version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), buildinfo.String())
			return err
		},
	}
}
