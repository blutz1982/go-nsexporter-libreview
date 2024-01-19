package libreview

import (
	"strings"
	"time"
)

type AuthResponse struct {
	Status int `json:"status"`
	Result struct {
		UserToken             string      `json:"UserToken"`
		AccountID             string      `json:"AccountId"`
		UserName              string      `json:"UserName"`
		FirstName             string      `json:"FirstName"`
		LastName              string      `json:"LastName"`
		MiddleInitial         string      `json:"MiddleInitial"`
		Email                 string      `json:"Email"`
		Country               string      `json:"Country"`
		Culture               string      `json:"Culture"`
		Timezone              interface{} `json:"Timezone"`
		DateOfBirth           string      `json:"DateOfBirth"`
		BackupFileExists      bool        `json:"BackupFileExists"`
		IsHCP                 bool        `json:"IsHCP"`
		Validated             bool        `json:"Validated"`
		NeedToAcceptPolicies  bool        `json:"NeedToAcceptPolicies"`
		CommunicationLanguage string      `json:"CommunicationLanguage"`
		UILanguage            string      `json:"UiLanguage"`
		SupportedDevices      interface{} `json:"SupportedDevices"`
		Created               string      `json:"Created"`
		GuardianLastName      string      `json:"GuardianLastName"`
		GuardianFirstName     string      `json:"GuardianFirstName"`
		DomainData            string      `json:"DomainData"`
		Consents              struct {
			RealWorldEvidence int `json:"realWorldEvidence"`
		} `json:"Consents"`
	} `json:"result"`
}

type ExtendedProperties struct {
	FactoryTimestamp       time.Time `json:"factoryTimestamp"`
	LowOutOfRange          string    `json:"lowOutOfRange"`
	HighOutOfRange         string    `json:"highOutOfRange"`
	IsFirstAfterTimeChange bool      `json:"isFirstAfterTimeChange"`
	CanMerge               string    `json:"canMerge"`
}

type UnscheduledExtendedProperties struct {
	FactoryTimestamp       time.Time `json:"factoryTimestamp"`
	LowOutOfRange          string    `json:"lowOutOfRange"`
	HighOutOfRange         string    `json:"highOutOfRange"`
	TrendArrow             string    `json:"trendArrow"`
	IsActionable           bool      `json:"isActionable"`
	IsFirstAfterTimeChange bool      `json:"isFirstAfterTimeChange"`
}

type UnscheduledContinuousGlucoseEntry struct {
	ValueInMgPerDl     float64                       `json:"valueInMgPerDl"`
	ExtendedProperties UnscheduledExtendedProperties `json:"extendedProperties"`
	RecordNumber       int64                         `json:"recordNumber"`
	Timestamp          time.Time                     `json:"timestamp"`
}

type UnscheduledContinuousGlucoseEntries []*UnscheduledContinuousGlucoseEntry

type UnscheduledVisitorFunc func(*UnscheduledContinuousGlucoseEntry, error) error

func (es UnscheduledContinuousGlucoseEntries) Visit(fn UnscheduledVisitorFunc) error {
	var err error
	for _, entry := range es {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}

func (es UnscheduledContinuousGlucoseEntries) Last() (*UnscheduledContinuousGlucoseEntry, bool) {
	var entry *UnscheduledContinuousGlucoseEntry
	es.Visit(func(e *UnscheduledContinuousGlucoseEntry, _ error) error {
		if entry == nil {
			entry = e
		} else {
			if e.Timestamp.After(entry.Timestamp) {
				entry = e
			}
		}
		return nil
	})
	return entry, (entry != nil)
}

func (es *UnscheduledContinuousGlucoseEntries) Append(e *UnscheduledContinuousGlucoseEntry) {
	*es = append(*es, e)
}

func NewUnscheduledFromScheduledContinuousGlucoseEntry(s *ScheduledContinuousGlucoseEntry, trendArrow string) *UnscheduledContinuousGlucoseEntry {
	entry := &UnscheduledContinuousGlucoseEntry{
		ValueInMgPerDl: s.ValueInMgPerDl,
		ExtendedProperties: UnscheduledExtendedProperties{
			FactoryTimestamp:       s.ExtendedProperties.FactoryTimestamp,
			LowOutOfRange:          s.ExtendedProperties.LowOutOfRange,
			HighOutOfRange:         s.ExtendedProperties.HighOutOfRange,
			TrendArrow:             trendArrow,
			IsActionable:           true,
			IsFirstAfterTimeChange: s.ExtendedProperties.IsFirstAfterTimeChange,
		},
		RecordNumber: s.RecordNumber, // TODO May change
		Timestamp:    s.Timestamp,
	}

	return entry
}

type GenericExtendedProperties struct {
	FactoryTimestamp time.Time `json:"factoryTimestamp"`
	Gmin             string    `json:"gmin"`
	Gmax             string    `json:"gmax"`
	WearDuration     string    `json:"wearDuration"`
	WarmupTime       string    `json:"warmupTime"`
	ProductType      string    `json:"productType"`
}

type GenericEntry struct {
	Type               string                    `json:"type"`
	ExtendedProperties GenericExtendedProperties `json:"extendedProperties"`
	RecordNumber       int64                     `json:"recordNumber"`
	Timestamp          time.Time                 `json:"timestamp"`
}

type GenericEntries []*GenericEntry

func (r *GenericEntries) Append(e *GenericEntry) {
	*r = append(*r, e)
}

type ScheduledContinuousGlucoseEntry struct {
	ValueInMgPerDl     float64            `json:"valueInMgPerDl"`
	ExtendedProperties ExtendedProperties `json:"extendedProperties"`
	RecordNumber       int64              `json:"recordNumber"`
	Timestamp          time.Time          `json:"timestamp"`
}

type ScheduledContinuousGlucoseEntries []*ScheduledContinuousGlucoseEntry

type VisitorFunc func(*ScheduledContinuousGlucoseEntry, error) error

func (es ScheduledContinuousGlucoseEntries) Last() (*ScheduledContinuousGlucoseEntry, bool) {
	var entry *ScheduledContinuousGlucoseEntry
	es.Visit(func(e *ScheduledContinuousGlucoseEntry, _ error) error {
		if entry == nil {
			entry = e
		} else {
			if e.Timestamp.After(entry.Timestamp) {
				entry = e
			}
		}
		return nil
	})

	return entry, (entry != nil)
}

func (es ScheduledContinuousGlucoseEntries) Visit(fn VisitorFunc) error {
	var err error
	for _, entry := range es {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}

func (es ScheduledContinuousGlucoseEntries) Filter(fn func(*ScheduledContinuousGlucoseEntry) bool) (result ScheduledContinuousGlucoseEntries) {
	es.Visit(func(e *ScheduledContinuousGlucoseEntry, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}

func (r *ScheduledContinuousGlucoseEntries) Append(e *ScheduledContinuousGlucoseEntry) {
	*r = append(*r, e)
}

type TreatmentExtendedProperties struct {
	FactoryTimestamp time.Time `json:"factoryTimestamp"`
}

type FoodEntry struct {
	ExtendedProperties TreatmentExtendedProperties `json:"extendedProperties"`
	RecordNumber       int64                       `json:"recordNumber"`
	Timestamp          time.Time                   `json:"timestamp"`
	GramsCarbs         int                         `json:"gramsCarbs"`
	FoodType           string                      `json:"foodType"`
}

type FoodEntries []*FoodEntry

func (fes *FoodEntries) Append(e *FoodEntry) {
	*fes = append(*fes, e)
}

type InsulinEntry struct {
	ExtendedProperties TreatmentExtendedProperties `json:"extendedProperties"`
	RecordNumber       int64                       `json:"recordNumber"`
	Timestamp          time.Time                   `json:"timestamp"`
	Units              float64                     `json:"units"`
	InsulinType        string                      `json:"insulinType"`
}

type InsulinEntries []*InsulinEntry

func (ies *InsulinEntries) Append(e *InsulinEntry) {
	*ies = append(*ies, e)
}

type FactoryConfig struct {
	Uom string `json:"UOM"`
}

type Miscellaneous struct {
	SelectedLanguage                     string `json:"selectedLanguage"`
	ValueGlucoseTargetRangeLowInMgPerDl  int    `json:"valueGlucoseTargetRangeLowInMgPerDl"`
	ValueGlucoseTargetRangeHighInMgPerDl int    `json:"valueGlucoseTargetRangeHighInMgPerDl"`
	SelectedTimeFormat                   string `json:"selectedTimeFormat"`
	SelectedCarbType                     string `json:"selectedCarbType"`
}

type DeviceSettings struct {
	FactoryConfig   FactoryConfig `json:"factoryConfig"`
	FirmwareVersion string        `json:"firmwareVersion"`
	Miscellaneous   Miscellaneous `json:"miscellaneous"`
}

type DeviceDataHeader struct {
	Device Device `json:"device"`
}

type Device struct {
	HardwareDescriptor string `json:"hardwareDescriptor"`
	OsVersion          string `json:"osVersion"`
	ModelName          string `json:"modelName"`
	OsType             string `json:"osType"`
	UniqueIdentifier   string `json:"uniqueIdentifier"`
	HardwareName       string `json:"hardwareName"`
}

type MeasurementLog struct {
	Capabilities                        []string                            `json:"capabilities"`
	BloodGlucoseEntries                 []any                               `json:"bloodGlucoseEntries"`
	GenericEntries                      GenericEntries                      `json:"genericEntries"`
	KetoneEntries                       []any                               `json:"ketoneEntries"`
	ScheduledContinuousGlucoseEntries   ScheduledContinuousGlucoseEntries   `json:"scheduledContinuousGlucoseEntries"`
	InsulinEntries                      InsulinEntries                      `json:"insulinEntries"`
	FoodEntries                         FoodEntries                         `json:"foodEntries"`
	UnscheduledContinuousGlucoseEntries UnscheduledContinuousGlucoseEntries `json:"unscheduledContinuousGlucoseEntries"`
}

type DeviceData struct {
	DeviceSettings DeviceSettings   `json:"deviceSettings"`
	Header         DeviceDataHeader `json:"header"`
	MeasurementLog MeasurementLog   `json:"measurementLog"`
}

type Measurements struct {
	UserToken   string     `json:"UserToken"`
	GatewayType string     `json:"GatewayType"`
	Domain      string     `json:"Domain"`
	DeviceData  DeviceData `json:"DeviceData"`
}

type Sensor struct {
	Domain      string     `json:"Domain"`
	DomainData  string     `json:"DomainData"`
	GatewayType string     `json:"GatewayType"`
	UserToken   string     `json:"UserToken"`
	DeviceData  DeviceData `json:"DeviceData"`
}

type APILibreViewAuth struct {
	Culture     string `json:"Culture"`
	DeviceId    string `json:"DeviceId"`
	GatewayType string `json:"GatewayType"`
	SetDevice   bool   `json:"SetDevice"`
	UserName    string `json:"UserName"`
	Domain      string `json:"Domain"`
	Password    string `json:"Password"`
}

type LibreViewExportResp struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
	Result struct {
		UploadID          string `json:"UploadId"`
		Status            int    `json:"Status"`
		MeasurementCounts struct {
			ScheduledGlucoseCount   int `json:"ScheduledGlucoseCount"`
			UnScheduledGlucoseCount int `json:"UnScheduledGlucoseCount"`
			BloodGlucoseCount       int `json:"BloodGlucoseCount"`
			InsulinCount            int `json:"InsulinCount"`
			GenericCount            int `json:"GenericCount"`
			FoodCount               int `json:"FoodCount"`
			KetoneCount             int `json:"KetoneCount"`
			TotalCount              int `json:"TotalCount"`
		} `json:"MeasurementCounts"`
		ItemCount       int             `json:"ItemCount"`
		CreatedDateTime LibreExportTime `json:"CreatedDateTime"`
		SerialNumber    string          `json:"SerialNumber"`
	} `json:"result"`
}

const LibreExportTimeLayout = "2006-01-02T15:04:05"

type LibreExportTime struct {
	time.Time
}

func (t *LibreExportTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return
	}

	t.Time, err = time.Parse(LibreExportTimeLayout, s)

	return
}
