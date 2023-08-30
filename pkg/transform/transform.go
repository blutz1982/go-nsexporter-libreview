package transform

import (
	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
)

func NSToLibreGlucoseEntry(e *nightscout.GlucoseEntry) *libreview.GlucoseEntry {

	return &libreview.GlucoseEntry{
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
