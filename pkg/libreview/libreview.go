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
	RecordNumberIncrementFood        = 460000000000
	RecordNumberIncrementGeneric     = 560000000000
)

var AllMeasurements = []string{
	"scheduledContinuousGlucose",
	"unscheduledContinuousGlucose",
	"insulin",
	"food",
	// "generic",
}

type Client interface {
	ImportMeasurements(modificators ...MeasuremenModificator) (*LibreViewExportResp, error)
	Auth(setDevice bool) error
	LastImported() *time.Time
	Token() string
	SetToken(token string)
	NewSensor(serial string) error
}

type libreview struct {
	config      *Config
	apiEndpoint *url.URL
	client      *http.Client
	userToken   string
	lastEntryTS *time.Time
}

func (lv *libreview) Token() string {
	return lv.userToken
}

func (lv *libreview) SetToken(token string) {
	lv.userToken = token
}

func NewWithConfig(config *Config) (Client, error) {

	u, err := url.Parse(config.ImportConfig.APIEndpoint)
	if err != nil {
		return nil, err
	}

	return &libreview{
		client:      http.DefaultClient,
		config:      config,
		apiEndpoint: u,
	}, nil
}

func (lv *libreview) Auth(setDevice bool) error {

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

	// debug
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

func WithFoodEntries(entries FoodEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.FoodEntries = entries
	}
}

func WithGenericEntries(entries GenericEntries) MeasuremenModificator {
	return func(l *MeasurementLog) {
		l.GenericEntries = entries
	}
}

func (lv *libreview) NewSensor(serial string) error {

	// Mock exit
	// if 0 == 0 {
	// 	return fmt.Errorf("Not implemented yet")
	// }

	s := &Sensor{
		Domain:      lv.config.ImportConfig.Domain,
		DomainData:  fmt.Sprintf("{\"activeSensor\":\"%s\"}", serial),
		GatewayType: lv.config.ImportConfig.GatewayType,
		UserToken:   lv.userToken,
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(s); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, lv.apiEndpoint.JoinPath("lsl", "api", "nisperson").String(), body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := lv.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Debug
	// data, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(string(data))

	if resp.StatusCode != 200 {
		return fmt.Errorf("Libreview post new sensor: bad http status code %d", resp.StatusCode)
	}

	return nil
}

func (lv *libreview) ImportMeasurements(modificators ...MeasuremenModificator) (exportResp *LibreViewExportResp, err error) {

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
				GenericEntries:                      GenericEntries{},
				KetoneEntries:                       []interface{}{},
				ScheduledContinuousGlucoseEntries:   ScheduledContinuousGlucoseEntries{},
				InsulinEntries:                      InsulinEntries{},
				FoodEntries:                         FoodEntries{},
				UnscheduledContinuousGlucoseEntries: UnscheduledContinuousGlucoseEntries{},
			},
		},
	}

	for _, fn := range modificators {
		fn(&m.DeviceData.MeasurementLog)
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(m); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, lv.apiEndpoint.JoinPath("lsl", "api", "measurements").String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	// debug
	// data, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(string(data))

	resp, err := lv.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// data, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(string(data))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Libreview post measurements: bad http status code %d", resp.StatusCode)
	}

	exportResp = new(LibreViewExportResp)

	if err := json.NewDecoder(resp.Body).Decode(exportResp); err != nil {
		return nil, err
	}

	if exportResp.Status != 0 {
		return nil, fmt.Errorf("Libreview post measurements: bad status code %d", exportResp.Status)
	}

	e, ok := m.DeviceData.MeasurementLog.ScheduledContinuousGlucoseEntries.Last()
	if ok {
		lv.lastEntryTS = &e.Timestamp
	}

	// Stub
	// For some reason the API does not return exportResp.Result.MeasurementCounts
	exportResp.Result.MeasurementCounts.ScheduledGlucoseCount = len(m.DeviceData.MeasurementLog.ScheduledContinuousGlucoseEntries)
	exportResp.Result.MeasurementCounts.UnScheduledGlucoseCount = len(m.DeviceData.MeasurementLog.UnscheduledContinuousGlucoseEntries)
	exportResp.Result.MeasurementCounts.InsulinCount = len(m.DeviceData.MeasurementLog.InsulinEntries)
	exportResp.Result.MeasurementCounts.FoodCount = len(m.DeviceData.MeasurementLog.FoodEntries)

	return
}

func (lv *libreview) LastImported() *time.Time {
	return lv.lastEntryTS
}
