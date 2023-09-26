package cmd

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/env"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newConfigCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "config",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newConfigSet(ctx),
		newPrintConfigCommand(ctx),
		newDefaultConfigCommand(ctx),
	)

	return cmd
}

func newPrintConfigCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "print",
		Short:         "print config",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := config.Default().WithDriver(yaml.Driver).LoadFiles(settings.ConfigPath); err != nil {
				return err
			}

			buff := new(bytes.Buffer)

			_, err := config.DumpTo(buff, config.Yaml)
			if err != nil {
				return err
			}

			fmt.Println(buff.String())

			return nil
		},
	}

	return cmd
}

func newDefaultConfigCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "default",
		Short:         "generate default config",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := loadDefaultConfig(); err != nil {
				return err
			}

			return config.Default().DumpToFile(settings.ConfigPath, config.Yaml)
		},
	}

	return cmd
}

func newConfigSet(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "set [KEY] [VALUE]",
		Short:         "set config key",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			key := args[0]
			value := args[1]

			cfg := config.NewEmpty("file").WithDriver(yaml.Driver)

			if err := cfg.LoadFiles(settings.ConfigPath); err != nil {
				return err
			}

			t, err := defaultTypeOf(key)
			if err != nil {
				return err
			}

			switch t.Kind() {
			case reflect.Uint64:
				val, err := strconv.Atoi(value)
				if err != nil {
					return errors.Wrap(err, "can't convert value to int")
				}
				err = cfg.Set(key, val)
				if err != nil {
					return err
				}

			default:
				err := cfg.Set(key, value)
				if err != nil {
					return err
				}

			}

			fmt.Printf("key: %s\nvalue: %s\nOK\n", key, value)

			return cfg.DumpToFile(settings.ConfigPath, config.Yaml)
		},
	}

	return cmd
}

func loadDefaultConfig() error {
	return config.Default().WithDriver(yaml.Driver).LoadStrings(config.Yaml, env.DefaultConfigYaml)
}

func defaultTypeOf(key string) (reflect.Type, error) {

	if err := loadDefaultConfig(); err != nil {
		return nil, err
	}

	value, ok := config.GetValue(key)
	if !ok {
		return nil, errors.Errorf("bad key %s", key)
	}
	return reflect.TypeOf(value), nil
}
