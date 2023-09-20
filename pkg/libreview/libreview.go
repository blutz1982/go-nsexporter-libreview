package libreview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	RecordNumberIncrement            = 160000000000
	RecordNumberIncrementUnscheduled = 260000000000
	RecordNumberIncrementInsulin     = 360000000000
)

var AllMeasurements = []string{
	"scheduledContinuousGlucose",
	"unscheduledContinuousGlucose",
	"insulin",
	// "food",
}

type Client interface {
	ImportMeasurements(modificators ...MeasuremenModificator) (sg, usg, ins, food int, err error)
	Auth(setDevice bool) error
	LastImported() *time.Time
}

type libreview struct {
	config      *Config
	apiEndpoint *url.URL
	client      *http.Client
	userToken   string
	lastEntryTS *time.Time
}

func NewWithConfig(config *Config) Client {
	return &libreview{
		client: http.DefaultClient,
		config: config,
	}
}

func (lv *libreview) Auth(setDevice bool) error {
	var err error

	lv.apiEndpoint, err = url.Parse(lv.config.ImportConfig.APIEndpoint)
	if err != nil {
		return err
	}

	apiAuth := &APILibreViewAuth{
		Culture:     lv.config.ImportConfig.Culture,
		DeviceId:    lv.config.ImportConfig.DevSettings.UniqueIdentifier,
		GatewayType: lv.config.ImportConfig.GatewayType,
		SetDevice:   setDevice,
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

	//// debug
	// data, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(string(data))

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

func WithScheduledGlucoseEntries(entries ScheduledContinuousGlucoseEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.ScheduledContinuousGlucoseEntries = entries

	}
}

func WithUnscheduledGlucoseEntries(entries UnscheduledContinuousGlucoseEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.UnscheduledContinuousGlucoseEntries = entries
	}
}

func WithInsulinEntries(entries InsulinEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.InsulinEntries = entries
	}
}

func (lv *libreview) ImportMeasurements(modificators ...MeasuremenModificator) (sg, usg, ins, food int, err error) {

	if len(modificators) == 0 {
		return
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
				ScheduledContinuousGlucoseEntries:   ScheduledContinuousGlucoseEntries{},
				InsulinEntries:                      InsulinEntries{},
				FoodEntries:                         []interface{}{},
				UnscheduledContinuousGlucoseEntries: UnscheduledContinuousGlucoseEntries{},
			},
		},
	}

	for _, fn := range modificators {
		fn(&m.DeviceData.MeasurementLog)
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(m); err != nil {
		return 0, 0, 0, 0, err
	}

	resp, err := lv.client.Post(lv.apiEndpoint.JoinPath("lsl", "api", "measurements").String(), "application/json", body)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer resp.Body.Close()

	// data, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(string(data))

	if resp.StatusCode != 200 {
		return 0, 0, 0, 0, fmt.Errorf("Libreview post measurements: bad http status code %d", resp.StatusCode)
	}

	exportResp := new(LibreViewExportResp)

	if err := json.NewDecoder(resp.Body).Decode(exportResp); err != nil {
		return 0, 0, 0, 0, err
	}

	if exportResp.Status != 0 {
		return 0, 0, 0, 0, fmt.Errorf("Libreview post measurements: bad status code %d", exportResp.Status)
	}

	e, ok := m.DeviceData.MeasurementLog.ScheduledContinuousGlucoseEntries.Last()
	if ok {
		lv.lastEntryTS = &e.Timestamp
	}

	// Stub
	// For some reason the API does not return exportResp.Result.MeasurementCounts
	sg = len(m.DeviceData.MeasurementLog.ScheduledContinuousGlucoseEntries)
	usg = len(m.DeviceData.MeasurementLog.UnscheduledContinuousGlucoseEntries)
	ins = len(m.DeviceData.MeasurementLog.InsulinEntries)
	food = len(m.DeviceData.MeasurementLog.FoodEntries)

	return
}

func (lv *libreview) LastImported() *time.Time {
	return lv.lastEntryTS
}
