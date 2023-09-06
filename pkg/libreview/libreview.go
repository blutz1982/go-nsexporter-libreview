package libreview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	RecordNumberIncrement            = 160000000000
	RecordNumberIncrementUnscheduled = 260000000000
)

type Client interface {
	ImportMeasurements(dryRun bool, modificators ...MeasuremenModificator) error
	Auth() error
}

type libreview struct {
	config      *Config
	apiEndpoint *url.URL
	client      *http.Client
	userToken   string
}

func NewWithConfig(config *Config) Client {
	return &libreview{
		client: http.DefaultClient,
		config: config,
	}
}

func (lv *libreview) Auth() error {
	var err error

	lv.apiEndpoint, err = url.Parse(lv.config.ImportConfig.APIEndpoint)
	if err != nil {
		return err
	}

	apiAuth := &APILibreViewAuth{
		Culture:     lv.config.ImportConfig.Culture,
		DeviceId:    lv.config.ImportConfig.DevSettings.UniqueIdentifier,
		GatewayType: lv.config.ImportConfig.GatewayType,
		SetDevice:   false,
		UserName:    lv.config.Auth.Username,
		Domain:      lv.config.ImportConfig.Domain,
		Password:    lv.config.Auth.Password,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(apiAuth); err != nil {
		return err
	}

	resp, err := lv.client.Post(lv.apiEndpoint.JoinPath("lsl", "api", "nisperson", "getauthentication").String(), "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Libreview auth: bad status code %d", resp.StatusCode)
	}

	authResponse := new(AuthResponse)

	if err := json.NewDecoder(resp.Body).Decode(authResponse); err != nil {
		return err
	}

	if len(authResponse.Result.UserToken) == 0 {
		return errors.New("cant get token")
	}

	lv.userToken = authResponse.Result.UserToken

	return nil

}

type MeasuremenModificator func(*MeasurementLog)

func WithScheduledGlucoseEntries(entries ScheduledGlucoseEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.ScheduledContinuousGlucoseEntries = entries

	}
}

func WithUnscheduledGlucoseEntries(entries UnscheduledContinuousGlucoseEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.UnscheduledContinuousGlucoseEntries = entries
	}
}

func (lv *libreview) ImportMeasurements(dryRun bool, modificators ...MeasuremenModificator) error {

	if len(modificators) == 0 {
		return nil
	}

	m := &Measurements{
		UserToken:   lv.userToken,
		GatewayType: lv.config.ImportConfig.GatewayType,
		Domain:      lv.config.ImportConfig.Domain,
		DeviceData: DeviceData{
			DeviceSettings: DeviceSettings{
				FactoryConfig: FactoryConfig{
					Uom: lv.config.ImportConfig.Uom,
				},
				FirmwareVersion: lv.config.ImportConfig.DevSettings.FirmwareVersion,
				Miscellaneous: Miscellaneous{
					SelectedLanguage:                     lv.config.ImportConfig.DevSettings.SelectedLanguage,
					ValueGlucoseTargetRangeLowInMgPerDl:  lv.config.ImportConfig.DevSettings.GlucoseTargetRangeLowInMgPerDl,
					ValueGlucoseTargetRangeHighInMgPerDl: lv.config.ImportConfig.DevSettings.GlucoseTargetRangeHighInMgPerDl,
					SelectedTimeFormat:                   lv.config.ImportConfig.DevSettings.SelectedTimeFormat,
					SelectedCarbType:                     lv.config.ImportConfig.DevSettings.SelectedCarbType,
				},
			},
			Header: DeviceDataHeader{
				Device: Device{
					HardwareDescriptor: lv.config.ImportConfig.DevSettings.HardwareDescriptor,
					OsVersion:          lv.config.ImportConfig.DevSettings.OSVersion,
					ModelName:          lv.config.ImportConfig.DevSettings.ModelName,
					OsType:             lv.config.ImportConfig.DevSettings.OSType,
					UniqueIdentifier:   lv.config.ImportConfig.DevSettings.UniqueIdentifier,
					HardwareName:       lv.config.ImportConfig.DevSettings.HardwareName,
				},
			},
			MeasurementLog: MeasurementLog{
				Capabilities: []string{
					"scheduledContinuousGlucose",
					"unscheduledContinuousGlucose",
					"bloodGlucose",
					"insulin",
					"food",
					"generic-com.abbottdiabetescare.informatics.exercise",
					"generic-com.abbottdiabetescare.informatics.customnote",
					"generic-com.abbottdiabetescare.informatics.ondemandalarm.low",
					"generic-com.abbottdiabetescare.informatics.ondemandalarm.high",
					"generic-com.abbottdiabetescare.informatics.ondemandalarm.projectedlow",
					"generic-com.abbottdiabetescare.informatics.ondemandalarm.projectedhigh",
					"generic-com.abbottdiabetescare.informatics.sensorstart",
					"generic-com.abbottdiabetescare.informatics.error",
					"generic-com.abbottdiabetescare.informatics.isfGlucoseAlarm",
					"generic-com.abbottdiabetescare.informatics.alarmSetting",
				},
				BloodGlucoseEntries:                 []interface{}{},
				GenericEntries:                      []interface{}{},
				KetoneEntries:                       []interface{}{},
				ScheduledContinuousGlucoseEntries:   []*ScheduledContinuousGlucoseEntry{},
				InsulinEntries:                      []interface{}{},
				FoodEntries:                         []interface{}{},
				UnscheduledContinuousGlucoseEntries: []*UnscheduledContinuousGlucoseEntry{},
			},
		},
	}

	for _, fn := range modificators {
		fn(&m.DeviceData.MeasurementLog)
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(m); err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	resp, err := lv.client.Post(lv.apiEndpoint.JoinPath("lsl", "api", "measurements").String(), "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	if resp.StatusCode != 200 {
		return fmt.Errorf("Libreview post measurements: bad http status code %d", resp.StatusCode)
	}

	exportResp := new(LibreViewExportResp)

	if err := json.NewDecoder(resp.Body).Decode(exportResp); err != nil {
		return err
	}

	if exportResp.Status != 0 {
		return fmt.Errorf("Libreview post measurements: bad status code %d", exportResp.Status)
	}

	return nil
}
