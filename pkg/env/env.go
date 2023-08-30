package env

import (
	"os"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type file struct {
	Nightscout *nightscout.Config `yaml:"nightscout"`
	Libreview  *libreview.Config  `yaml:"libreview"`
}

type EnvSettings struct {
	version    string
	ConfigPath string
	Debug      bool
	config     *file
}

var defaultEnv *EnvSettings

var Default = New()

func New() *EnvSettings {
	env := &EnvSettings{
		ConfigPath: "config.yaml",
		Debug:      false,
	}

	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddCommonFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.ConfigPath, "config", "c", s.ConfigPath, "path to config")
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "toggle debug")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func (s *EnvSettings) LoadConfig() error {
	f := new(file)

	b, err := os.ReadFile(s.ConfigPath)
	if err != nil {
		return errors.Wrapf(err, "couldn't load config file (%s)", s.ConfigPath)
	}

	if err := yaml.Unmarshal(b, f); err != nil {
		return err
	}

	s.config = f

	return nil

}

func (s *EnvSettings) Version() string {
	return s.version
}

func (s *EnvSettings) SetVersion(version string) {
	s.version = version
}

func (s *EnvSettings) Nightscout() *nightscout.Config {
	return s.config.Nightscout
}

func (s *EnvSettings) Libreview() *libreview.Config {
	return s.config.Libreview
}
