// Command kolang-linter is a standalone linter for the Kolang (کلنگ)
// programming language. It reads Kolang source from a file argument or stdin
// and emits JSON diagnostics to stdout.
//
// Usage:
//
//	kolang-linter [file]
//	kolang-linter -format json [file]
//
// Exit code is 0 on a successful lint run (even when diagnostics are
// reported) and 1 on an internal error (unreadable input, unsupported flag).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/faralidev/kolang-linter/internal/diag"
	"github.com/faralidev/kolang-linter/internal/rules"
)

func main() {
	format := flag.String("format", "json", "output format (only \"json\" for v1)")
	strict := flag.Bool("strict", false, "treat warnings as errors in the exit code (reserved; not used in v1)")
	flag.Parse()

	if *format != "json" {
		fmt.Fprintf(os.Stderr, "kolang-linter: unsupported format %q\n", *format)
		os.Exit(1)
	}
	_ = strict // accepted for CLI compatibility; no behavioral change in v1

	src, err := readInput(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "kolang-linter: %v\n", err)
		os.Exit(1)
	}

	diagnostics := rules.NewRegistry().Run(string(src))
	out, err := diag.Marshal(diagnostics)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kolang-linter: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}

// readInput reads source from the file argument if given, otherwise from stdin.
func readInput(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}