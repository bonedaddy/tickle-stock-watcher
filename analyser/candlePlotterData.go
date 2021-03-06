package analyser

import (
	"image/color"
	"math"

	"github.com/sdcoffey/techan"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Candles array of candle
type Candles []candle

// Len length of the array
func (c Candles) Len() int { return len(c) }

// Candle timestamp, open, close, high, low
func (c Candles) Candle(i int) (float64, float64, float64, float64, float64) {
	return c[i].Timestamp, c[i].Open, c[i].Close, c[i].High, c[i].Low
}

type candler interface {
	Len() int
	Candle(int) (float64, float64, float64, float64, float64)
}

type candle struct {
	Timestamp, Open, Close, High, Low float64
}

func copyCandles(data candler) Candles {
	cp := make(Candles, data.Len())
	for i := range cp {
		cp[i].Timestamp, cp[i].Open, cp[i].Close, cp[i].High, cp[i].Low = data.Candle(i)
	}
	return cp
}

// CandleSticks struct of candle sticks
type CandleSticks struct {
	Candles
	days               int
	timeSeries         *techan.TimeSeries
	UpColor, DownColor color.Color
	isPromising        func(int) bool
}

// NewCandleSticks factory method for candle sticks
func NewCandleSticks(cs Candles, timeSeries *techan.TimeSeries, days int, up, down color.Color) *CandleSticks {
	cp := copyCandles(cs)
	return &CandleSticks{
		Candles:     cp,
		days:        days,
		timeSeries:  timeSeries,
		UpColor:     up,
		DownColor:   down,
		isPromising: newProspectCriteriaMACD(timeSeries),
	}
}

// Plot Plot
func (cs *CandleSticks) Plot(c draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&c)

	// Plot candlestick
	for i, d := range cs.Candles {
		if i < cs.Len()-cs.days {
			continue
		}
		x0 := trX(d.Timestamp)
		x1 := trX(d.Timestamp + 24*60*60) // 24시간
		y0 := trY(d.Open)
		y1 := trY(d.Close)

		if y0 <= y1 {
			c.SetColor(cs.UpColor)
		} else {
			c.SetColor(cs.DownColor)
		}

		var p vg.Rectangle
		p.Min = vg.Point{X: x0, Y: vg.Length(math.Min(float64(y0), float64(y1)))}
		p.Max = vg.Point{X: x1, Y: vg.Length(math.Max(float64(y0), float64(y1)))}
		c.Fill(p.Path())

		x0 = trX(d.Timestamp + 8*60*60)
		x1 = trX(d.Timestamp + 16*60*60) // 8시간
		y0 = trY(d.High)
		y1 = trY(d.Low)

		var q vg.Rectangle
		q.Min = vg.Point{X: x0, Y: y0}
		q.Max = vg.Point{X: x1, Y: y1}
		c.Fill(q.Path())
	}

	for i, d := range cs.Candles {
		if i < cs.Len()-cs.days {
			continue
		}
		if !cs.isPromising(i) {
			continue
		}

		x0 := trX(d.Timestamp)
		x1 := trX(d.Timestamp + 24*60*60) // 24시간
		y0 := trY(d.Open)
		y1 := trY(d.Close)
		var q vg.Rectangle
		q.Min = vg.Point{X: x0, Y: vg.Length(math.Min(float64(y0), float64(y1)))}
		q.Max = vg.Point{X: x1, Y: vg.Length(math.Max(float64(y0), float64(y1)))}
		c.SetColor(color.RGBA{R: 255, G: 255, A: 255})
		c.Fill(q.Path())
	}
}

// DataRange DataRange
func (cs *CandleSticks) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin = cs.Candles[cs.Len()-cs.days].Timestamp
	xmax = cs.Candles[cs.Len()-1].Timestamp + 24*60*60

	ymin = cs.Candles[cs.Len()-cs.days].Low
	ymax = cs.Candles[cs.Len()-cs.days].High

	for _, d := range cs.Candles[cs.Len()-cs.days:] {
		ymin = math.Min(ymin, d.Low)
		ymax = math.Max(ymax, d.High)
	}

	return xmin, xmax, ymin, ymax
}

// GlyphBoxes GlyphBoxes
func (cs *CandleSticks) GlyphBoxes(p *plot.Plot) []plot.GlyphBox {
	boxes := make([]plot.GlyphBox, cs.Len())
	for i, d := range cs.Candles {
		boxes[i].X = p.X.Norm(d.Timestamp + 12*60*60)
		boxes[i].Y = p.Y.Norm((d.High + d.Low) * 0.5)

		h := (p.X.Norm(d.Timestamp+24*60*60) - p.X.Norm(d.Timestamp)) * 0.5
		r := (p.Y.Norm(d.High) - p.Y.Norm(d.Low)) * 0.5

		boxes[i].Rectangle = vg.Rectangle{
			Min: vg.Point{X: vg.Length(-h), Y: vg.Length(-r)},
			Max: vg.Point{X: vg.Length(h), Y: vg.Length(r)},
		}
	}
	return boxes
}
