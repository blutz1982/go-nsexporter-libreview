package nsgraph

import (
	"fmt"
	"io"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/transform"
	"github.com/golang/freetype/truetype"

	_ "embed"

	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

//go:embed DejaVuSerif.ttf
var fontData []byte

func dotSizeFromCount(count int) chart.SizeProvider {
	return func(xrange, yrange chart.Range, index int, x, y float64) float64 {
		switch v := count; {
		case v < 100:
			return 4
		case v >= 100 && v < 300:
			return 3
		case v >= 300 && v < 600:
			return 2
		default:
			return 1
		}
	}
}

func DrawChart(entries *nightscout.GlucoseEntries, dst io.Writer, targetLow, targetHigh float64) error {
	s := chart.TimeSeries{
		Style: chart.Style{
			StrokeWidth:      1, //chart.Disabled
			DotWidthProvider: dotSizeFromCount(entries.Len()),
			DotColorProvider: colorizeDots(targetLow, targetHigh),
		},
	}

	var lastEntry *nightscout.GlucoseEntry

	entries.Visit(func(e *nightscout.GlucoseEntry, err error) error {
		if lastEntry == nil {
			lastEntry = e
		}
		s.XValues = append(s.XValues, e.Date.Time().Local())
		s.YValues = append(s.YValues, e.Sgv.MMol())
		return nil
	})

	graph := chart.Chart{
		Width:  2048,
		Height: 800,
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeValueFormatterWithFormat("01/02 15:04"),
			Style: chart.Style{
				FontColor:   chart.ColorWhite,
				StrokeColor: chart.ColorWhite,
			},
		},
		Title:  "Nightscout",
		Canvas: chart.Style{FillColor: chart.ColorBlack},
		TitleStyle: chart.Style{
			FontColor: chart.ColorWhite,
			FontSize:  20,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   25,
				Right:  25,
				Bottom: 10,
			},
			FillColor: chart.ColorBlack,
		},
		YAxis: chart.YAxis{
			GridMajorStyle: chart.Style{
				StrokeColor:     chart.ColorAlternateGray,
				StrokeWidth:     1.0,
				StrokeDashArray: []float64{5, 5},
			},
			GridLines: []chart.GridLine{
				{Value: 3.2},
				{Value: targetLow},
				{Value: targetHigh},
				{Value: 14.5},
			},
			Style: chart.Style{
				FontColor:   chart.ColorWhite,
				StrokeColor: chart.ColorWhite,
			},

			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: 26.0,
			},
		},
		Series: []chart.Series{
			s,
			// TODO must reverse sort
			// chart.LastValueAnnotationSeries(series),
		},
	}

	if lastEntry != nil {
		glusoseWithArrow := fmt.Sprintf("%.1f %s", lastEntry.Sgv.MMol(), transform.ToArrow(lastEntry.Direction))
		graph.Elements = []chart.Renderable{
			DrawTime(lastEntry.Date.Time().Local().Format("15:04")),
			DrawLastGlucose(glusoseWithArrow, colorizeRange(lastEntry.Sgv.MMol(), targetLow, targetHigh)),
		}

	}

	err := graph.Render(chart.PNG, dst)

	return err
}

func colorizeRange(value, min, max float64) drawing.Color {
	switch v := value; {
	case v >= 14.5:
		return drawing.Color{R: 217, G: 0, B: 0, A: 255}
	case v >= max && v < 14.5:
		return chart.ColorYellow
	case v <= min:
		return drawing.Color{R: 217, G: 0, B: 0, A: 255}
	default:
		return chart.ColorGreen
	}
}

func colorizeDots(min, max float64) chart.DotColorProvider {
	return func(xrange, yrange chart.Range, index int, x, y float64) drawing.Color {
		return colorizeRange(y, min, max)
	}
}

func DrawTime(text string) chart.Renderable {
	return func(r chart.Renderer, cb chart.Box, chartDefaults chart.Style) {
		x := cb.Left + ((cb.Width() / 100) * 1)
		y := cb.Top + ((cb.Height() / 100) * 10)
		r.SetFontColor(chart.ColorLightGray)
		r.SetFontSize(40)
		r.Text(text, x, y)
	}
}

func DrawLastGlucose(text string, color drawing.Color) chart.Renderable {
	return func(r chart.Renderer, cb chart.Box, chartDefaults chart.Style) {
		x := cb.Left + ((cb.Width() / 100) * 85)
		y := cb.Top + ((cb.Height() / 100) * 10)
		r.SetFontColor(color)
		r.SetFontSize(40)
		r.SetFont(MustGetTTFont(fontData))
		r.Text(text, x, y)
	}
}

func MustGetTTFont(TTFData []byte) *truetype.Font {

	font, err := truetype.Parse(TTFData)
	if err != nil {
		panic(err)
	}

	return font
}
