package nightscout

type Config struct {
	URL       string `yaml:"url"`
	APIToken  string `yaml:"apiToken"`
	APISecret string `yaml:"apiSecret"`
}
