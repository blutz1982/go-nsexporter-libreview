package env

import (
	"fmt"
	"os"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const DefaultConfigYaml string = `
libreview:
  auth:
    password: ""
    username: ""
  importConfig:
    apiEndpoint: https://api.libreview.ru
    culture: ru-RU
    deviceSettings:
      firmwareVersion: 2.8.2
      glucoseTargetRangeHighInMgPerDl: 144
      glucoseTargetRangeLowInMgPerDl: 90
      hardwareDescriptor: Redmi Note 8 Pro
      hardwareName: Xiaomi
      modelName: com.freestylelibre.app.ru
      osType: Android
      osVersion: "29"
      selectedCarbType: grams of carbs
      selectedLanguage: ru_RU
      selectedTimeFormat: 24hr
      uniqueIdentifier: ""
    domain: Libreview
    gatewayType: FSLibreLink.Android
    uom: mmol/L
nightscout:
  apiToken: ""
  url: ""
`

type file struct {
	Nightscout *nightscout.Config `yaml:"nightscout"`
	Libreview  *libreview.Config  `yaml:"libreview"`
}

type EnvSettings struct {
	appName    string
	version    string
	ConfigPath string
	Debug      bool
	Timezone   string
	config     *file
}

var defaultEnv *EnvSettings

var Default = New()

func New() *EnvSettings {
	env := &EnvSettings{
		ConfigPath: "config.yaml",
		Debug:      false,
		Timezone:   "",
	}
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddCommonFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.ConfigPath, "config", "c", s.ConfigPath, "path to config")
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "toggle debug")
	fs.StringVar(&s.Timezone, "timezone", s.Timezone, "override timezone")
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
		fmt.Printf("bad yaml: %s\n---\n%s\n", s.ConfigPath, string(b))
		return err
	}

	s.config = f

	return nil

}

func (s *EnvSettings) SaveConfig() error {

	data, err := yaml.Marshal(s.config)
	if err != nil {
		return err
	}

	return os.WriteFile(s.ConfigPath, data, 0644)
}

func (s *EnvSettings) Version() string {
	return s.version
}

func (s *EnvSettings) SetVersion(version string) {
	s.version = version
}

func (s *EnvSettings) AppName() string {
	return s.appName
}

func (s *EnvSettings) SetAppName(name string) {
	s.appName = name
}

func (s *EnvSettings) Nightscout() *nightscout.Config {
	return s.config.Nightscout
}

func (s *EnvSettings) Libreview() *libreview.Config {
	return s.config.Libreview
}

func (s *EnvSettings) SetNightscout(cfg *nightscout.Config) {
	s.config.Nightscout = cfg
}
