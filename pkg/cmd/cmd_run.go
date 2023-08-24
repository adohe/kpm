// Copyright 2023 The KCL Authors. All rights reserved.

package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"kcl-lang.io/kcl-go/pkg/kcl"
	"kcl-lang.io/kpm/pkg/api"
	"kcl-lang.io/kpm/pkg/errors"
	"kcl-lang.io/kpm/pkg/opt"
)

// NewRunCmd new a Command for `kpm run`.
func NewRunCmd() *cli.Command {
	return &cli.Command{
		Hidden: false,
		Name:   "run",
		Usage:  "compile kcl package.",
		Flags: []cli.Flag{
			// The entry kcl file.
			&cli.StringSliceFlag{
				Name:  FLAG_INPUT,
				Usage: "a kcl file as the compile entry file",
			},
			&cli.StringFlag{
				Name:  FLAG_TAG,
				Usage: "the tag for oci artifact",
			},
			// '--vendor' will trigger the vendor mode
			// In the vendor mode, the package search path is the subdirectory 'vendor' in current package.
			// In the non-vendor mode, the package search path is the $KCL_PKG_PATH.
			&cli.BoolFlag{
				Name:  FLAG_VENDOR,
				Usage: "run in vendor mode",
			},

			// KCL arg: --setting, -Y
			&cli.StringSliceFlag{
				Name:    FLAG_SETTING,
				Aliases: []string{"Y"},
				Usage:   "Specify the input setting file",
			},

			// KCL arg: --argument, -D
			&cli.StringSliceFlag{
				Name:    FLAG_ARGUMENT,
				Aliases: []string{"D"},
				Usage:   "Specify the top-level argument",
			},

			// KCL arg: --overrides, -O
			&cli.StringSliceFlag{
				Name:    FLAG_OVERRIDES,
				Aliases: []string{"O"},
				Usage:   "Specify the configuration override path and value",
			},

			// KCL arg: --disable_none, -n
			&cli.BoolFlag{
				Name:    FLAG_DISABLE_NONE,
				Aliases: []string{"n"},
				Usage:   "Disable dumping None values",
			},

			// KCL arg: --sort_keys -k
			&cli.BoolFlag{
				Name:    FLAG_SORT_KEYS,
				Aliases: []string{"k"},
				Usage:   "Sort result keys",
			},
		},
		Action: func(c *cli.Context) error {
			return KpmRun(c)
		},
	}
}

func KpmRun(c *cli.Context) error {
	kclOpts := CompileOptionFromCli(c)
	pkgWillBeCompiled := c.Args().First()
	// 'kpm run' compile the current package undor '$pwd'.
	if len(pkgWillBeCompiled) == 0 {
		compileResult, err := api.RunPkg(kclOpts)
		if err != nil {
			return err
		}
		fmt.Println(compileResult)
	} else {
		// 'kpm run <package source>' compile the kcl package from the <package source>.
		kclOpts.SetPkgPath(pkgWillBeCompiled)
		compileResult, err := api.RunPkgInPath(kclOpts)
		if err == errors.FailedToLoadPackage {
			compileResult, err = api.RunTar(pkgWillBeCompiled, kclOpts)
			if err == errors.InvalidKclPacakgeTar {
				compileResult, err = api.RunOci(pkgWillBeCompiled, c.String(FLAG_TAG), kclOpts)
			}
		}

		if err != nil {
			return err
		}
		fmt.Println(compileResult)
	}
	return nil
}

// CompileOptionFromCli will parse the kcl options from the cli context.
func CompileOptionFromCli(c *cli.Context) *opt.CompileOptions {
	opts := opt.DefaultCompileOptions()

	// --input
	opts.ExtendEntries(c.StringSlice(FLAG_INPUT))

	// --vendor
	opts.SetVendor(c.Bool(FLAG_VENDOR))

	// --setting, -Y
	settingsOpt := c.StringSlice(FLAG_SETTING)
	if len(settingsOpt) != 0 {
		for _, sPath := range settingsOpt {
			opts.Merge(kcl.WithSettings(sPath))
		}
		opts.SetHasSettingsYaml(true)
	}

	// --argument, -D
	argumentsOpt := c.StringSlice(FLAG_ARGUMENT)
	if len(argumentsOpt) != 0 {
		for _, arg := range argumentsOpt {
			opts.Merge(kcl.WithOptions(arg))
		}
	}

	// --overrides, -O
	overridesOpt := c.StringSlice(FLAG_OVERRIDES)
	if len(overridesOpt) != 0 {
		for _, override := range overridesOpt {
			opts.Merge(kcl.WithOverrides(override))
		}
	}

	// --disable_none, -n
	opts.Merge(kcl.WithDisableNone(c.Bool(FLAG_DISABLE_NONE)))

	// --sort_keys, -k
	opts.Merge(kcl.WithSortKeys(c.Bool(FLAG_SORT_KEYS)))

	return opts
}
