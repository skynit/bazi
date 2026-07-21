package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"bazi/internal/service/precision"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("precision-report", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "Go module root containing internal/service/testdata")
	out := flags.String("out", "", "optional JSON output path; defaults to stdout")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	report, err := precision.BuildReport(precision.Options{RootDir: *root})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "build precision report: %v\n", err)
		return 1
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "encode precision report: %v\n", err)
		return 1
	}
	data = append(data, '\n')
	if *out == "" {
		if _, err := stdout.Write(data); err != nil {
			_, _ = fmt.Fprintf(stderr, "write precision report: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		_, _ = fmt.Fprintf(stderr, "write precision report %s: %v\n", *out, err)
		return 1
	}
	return 0
}
