// plan9asmtest exhaustively partitions one package's actual Go test inventory.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	var cfg config
	flag.StringVar(&cfg.Go, "go", "go", "Go executable")
	flag.StringVar(&cfg.Dir, "dir", ".", "package working directory")
	flag.StringVar(&cfg.Package, "package", ".", "one package, not a recursive pattern")
	flag.StringVar(&cfg.Out, "out", "_out/root-suite", "new, nonexisting artifact directory")
	flag.IntVar(&cfg.Shards, "shards", 32, "stable hash partitions, run serially")
	flag.DurationVar(&cfg.Timeout, "timeout", 45*time.Minute, "per-command wall deadline and Go test timeout")
	flag.Parse()
	cfg.Flags = flag.Args()

	ctx, stop := signal.NotifyContext(context.Background(), cancellationSignals()...)
	defer stop()
	if err := runSuite(ctx, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "plan9asmtest:", err)
		os.Exit(1)
	}
}
