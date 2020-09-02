package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	flags, opts := setupFlags(args[0])
	err := flags.Parse(args[1:])
	switch {
	case err == flag.ErrHelp:
		return nil
	case err != nil:
		return err
	}
	return runMog(*opts)
}

type options struct {
	source string
}

func setupFlags(name string) (*flag.FlagSet, *options) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	opts := &options{}

	// TODO: make this a positional arg, set a Usage func to document it
	flags.StringVar(&opts.source, "source", ".", "package path for source structs")
	return flags, opts
}

func runMog(opts options) error {
	if opts.source == "" {
		return fmt.Errorf("missing required source package")
	}

	sources, err := loadSourceStructs(opts.source)
	if err != nil {
		return fmt.Errorf("failed to load source from %s: %w", opts.source, err)
	}

	cfg, err := configsFromAnnotations(sources)
	if err != nil {
		return fmt.Errorf("failed to parse annotations: %w", err)
	}

	log.Printf("Generated code for %d structs", len(cfg.Structs))
	targets, err := loadTargetStructs(targetPackages(cfg.Structs))
	if err != nil {
		return fmt.Errorf("failed to load targets: %w", err)
	}

	return generateFiles(cfg, targets)
}

func targetPackages(cfgs []structConfig) []string {
	result := make([]string, 0, len(cfgs))
	for _, cfg := range cfgs {
		if cfg.Target.Package == "" {
			continue
		}
		result = append(result, cfg.Target.Package)
	}
	return result
}