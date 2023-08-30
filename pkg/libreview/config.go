package libreview

type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DevSettings struct {
	SelectedLanguage                string `yaml:"selectedLanguage"`
	SelectedTimeFormat              string `yaml:"selectedTimeFormat"`
	SelectedCarbType                string `yaml:"selectedCarbType"`
	GlucoseTargetRangeLowInMgPerDl  int    `yaml:"glucoseTargetRangeLowInMgPerDl"`
	GlucoseTargetRangeHighInMgPerDl int    `yaml:"glucoseTargetRangeHighInMgPerDl"`
	FirmwareVersion                 string `yaml:"firmwareVersion"`
	HardwareName                    string `yaml:"hardwareName"`
	HardwareDescriptor              string `yaml:"hardwareDescriptor"`
	OSType                          string `yaml:"osType"`
	OSVersion                       string `yaml:"osVersion"`
	ModelName                       string `yaml:"modelName"`
	UniqueIdentifier                string `yaml:"uniqueIdentifier"`
}

type ImportConfig struct {
	APIEndpoint string      `yaml:"apiEndpoint"`
	Domain      string      `yaml:"domain"`
	Culture     string      `yaml:"culture"`
	GatewayType string      `yaml:"gatewayType"`
	Uom         string      `yaml:"uom"`
	DevSettings DevSettings `yaml:"deviceSettings"`
}

type Config struct {
	Auth         Auth         `yaml:"auth"`
	ImportConfig ImportConfig `yaml:"importConfig"`
}
