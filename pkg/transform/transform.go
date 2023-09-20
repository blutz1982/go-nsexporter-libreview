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

func NSToLibreInsulinEntry(e *nightscout.Treatment) *libreview.InsulinEntry {
	return &libreview.InsulinEntry{
		ExtendedProperties: libreview.InsulinExtendedProperties{
			FactoryTimestamp: e.CreatedAt,
		},
		RecordNumber: libreview.RecordNumberIncrementInsulin + e.CreatedAt.Unix(),
		Timestamp:    e.CreatedAt.Local(),
		Units:        e.Insulin,
		InsulinType:  "RapidActing",
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
