package env

import (
	"fmt"
	"os"
	"time"

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

const (
	defaultTSLayout = "2006-01-02"
)

type file struct {
	Nightscout *nightscout.Config `yaml:"nightscout"`
	Libreview  *libreview.Config  `yaml:"libreview"`
}

type EnvSettings struct {
	appName    string
	ConfigPath string
	Debug      bool
	Timezone   string
	config     *file
	listFlags  *ListFlags
}

type ListFlags struct {
	tsLayout   string
	fromDate   string
	dateOffset string
	toDate     string
	count      int
}

var defaultEnv *EnvSettings

var Default = New()

func New() *EnvSettings {
	env := &EnvSettings{
		ConfigPath: "config.yaml",
		Debug:      false,
		Timezone:   "",
		listFlags: &ListFlags{
			tsLayout: defaultTSLayout,
			count:    nightscout.MaxEnties,
		},
	}
	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddCommonFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.ConfigPath, "config", "c", s.ConfigPath, "path to config")
	fs.BoolVarP(&s.Debug, "debug", "d", s.Debug, "toggle debug")
	fs.StringVar(&s.Timezone, "timezone", s.Timezone, "override timezone")
}

func (s *EnvSettings) AddListFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.listFlags.tsLayout, "ts-layout", s.listFlags.tsLayout, "Timestamp layout for --date-from and --date-to flags. More https://go.dev/src/time/format.go")
	fs.StringVar(&s.listFlags.fromDate, "date-from", "", "Start of sampling period")
	fs.StringVar(&s.listFlags.dateOffset, "date-offset", "", "Start of sampling period with current time offset. Set in duration (e.g. 24h or 72h30m). Ignore --date-from and --date-to flags")
	fs.StringVar(&s.listFlags.toDate, "date-to", "", "End of sampling period")
	fs.IntVar(&s.listFlags.count, "max-count", s.listFlags.count, "nightscout max count entries per API request")
}

func (s *EnvSettings) DateRange() (fromDate, toDate time.Time, err error) {
	return getDateRange(s.listFlags)
}

func (s *EnvSettings) NightscoutMaxEnties() int {
	return s.listFlags.count
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

	if f.Nightscout == nil {
		f.Nightscout = new(nightscout.Config)
	}

	if f.Libreview == nil {
		f.Libreview = new(libreview.Config)
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

func getDateRange(f *ListFlags) (fromDate, toDate time.Time, err error) {

	var (
		duration time.Duration
	)

	fromDateStr := f.fromDate
	toDateStr := f.toDate

	if len(f.dateOffset) > 0 {
		fromDateStr = ""
		toDateStr = ""
		duration, err = time.ParseDuration(f.dateOffset)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if len(fromDateStr) == 0 {
		now := time.Now().Local()
		if len(f.dateOffset) > 0 {
			fromDate = now.Add(-duration)
		} else {
			fromDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		}
	} else {
		fromDate, err = time.ParseInLocation(f.tsLayout, fromDateStr, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if len(toDateStr) == 0 {
		toDate = time.Now().Local()
	} else {
		toDate, err = time.ParseInLocation(f.tsLayout, toDateStr, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return

}
