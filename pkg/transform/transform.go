package transform

import (
	"math/rand"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
)

func NSToLibreScheduledGlucoseEntry(e *nightscout.GlucoseEntry) *libreview.ScheduledContinuousGlucoseEntry {

	return &libreview.ScheduledContinuousGlucoseEntry{
		ValueInMgPerDl: e.Sgv.Float64(),
		ExtendedProperties: libreview.ExtendedProperties{
			FactoryTimestamp:       e.SysTime,
			LowOutOfRange:          e.Sgv.LowOutOfRange(nightscout.DefaultMinSVG),
			HighOutOfRange:         e.Sgv.HighOutOfRange(nightscout.DefaultMaxSVG),
			IsFirstAfterTimeChange: false,
			CanMerge:               "true",
		},
		RecordNumber: libreview.RecordNumberIncrement + e.DateString.Unix(),
		Timestamp:    e.DateString.Local(),
	}
}

func NSToLibreUnscheduledGlucoseEntry(e *nightscout.GlucoseEntry) *libreview.UnscheduledContinuousGlucoseEntry {
	var duration = time.Minute * time.Duration(rand.Intn(3))
	return &libreview.UnscheduledContinuousGlucoseEntry{
		ValueInMgPerDl: e.Sgv.Float64(),
		ExtendedProperties: libreview.UnscheduledExtendedProperties{
			FactoryTimestamp:       e.SysTime,
			LowOutOfRange:          e.Sgv.LowOutOfRange(nightscout.DefaultMinSVG),
			HighOutOfRange:         e.Sgv.HighOutOfRange(nightscout.DefaultMaxSVG),
			IsFirstAfterTimeChange: false,
			TrendArrow:             ToLibreDirection(e.Direction),
			IsActionable:           true,
		},
		RecordNumber: libreview.RecordNumberIncrementUnscheduled + e.DateString.Unix(),
		Timestamp:    e.DateString.Local().Add(duration),
	}
}

var LongActingInsulinMap = map[bool]string{
	true:  "LongActing",
	false: "RapidActing",
}

func NSToLibreInsulinEntry(t *nightscout.Treatment) *libreview.InsulinEntry {

	return &libreview.InsulinEntry{
		ExtendedProperties: libreview.TreatmentExtendedProperties{
			FactoryTimestamp: t.CreatedAt,
		},
		RecordNumber: libreview.RecordNumberIncrementInsulin + t.CreatedAt.Unix(),
		Timestamp:    t.CreatedAt.Local(),
		Units:        t.Insulin,
		InsulinType:  LongActingInsulinMap[t.InsulinInjections.IsLongActing()],
	}
}

func NSToLibreFoodEntry(e *nightscout.Treatment) *libreview.FoodEntry {
	return &libreview.FoodEntry{
		ExtendedProperties: libreview.TreatmentExtendedProperties{
			FactoryTimestamp: e.CreatedAt,
		},
		RecordNumber: libreview.RecordNumberIncrementFood + e.CreatedAt.Unix(),
		Timestamp:    e.CreatedAt.Local(),
		GramsCarbs:   int(e.Carbs),
		FoodType:     "Unknown",
	}
}

// https://github.com/nightscout/cgm-remote-monitor/blob/46418c7ff275ae80de457209c1686811e033b5dd/lib/plugins/direction.js#L53
// TODO: find out all possible values ​​for Libre TrendArrow field

func ToLibreDirection(nsDirection string) string {
	switch nsDirection {
	case "Flat":
		return "Stable"
	case "FortyFiveDown", "SingleDown", "DoubleDown", "TripleDown":
		return "Falling"
	case "FortyFiveUp", "SingleUp", "DoubleUp", "TripleUp":
		return "Rising"
	default:
		return "Stable"
	}

}
