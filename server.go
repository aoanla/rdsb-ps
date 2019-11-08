package main

//initally taken from gist linked by https://www.sanarias.com/blog/1214PlayingwithimagesinHTTPresponseingolang

import (
	"bytes"
	//"encoding/base64"
	"flag"
	//"html/template"
	"image"
	"image/color"
	//"image/draw"
	"image/png"
	"log"
	"net/http"
	"strconv"

	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"sort"
)

var root = flag.String("root", ".", "file system path")

//Pager is a map[string]func(bytebuffer)bool
// these map paths to the function to generate their in-memory representation
var Pager map[string]func(*bytes.Buffer)bool
var Plotter map[string]func(*imgg.ImageGraphics)bool

//thes map paths to the doublebuffer which holds their in-memory representation
var Pages map[string]*Page_buffer
var Plots map[string]*Page_buffer


func writeHTMLwithPaths(paths []string) {
	for _, p := range paths {
		var buffer bytes.Buffer
		//write to buffer
		if !Pager[p](&buffer) {
			log.Printf("error writing page with path %s", p)
		}
		Pages[p].Write(buffer.Bytes())
		Pages[p].Switch()
	}
}




func drawPlotswithPaths(paths []string) {
	for _, p := range paths {
		igr := imgg.New(1024,768, color.RGBA{0xff,0xff,0xff,0xff}, nil, nil)

		//success := Plotter[p](&igr)
		if !Plotter[p](igr) {
			log.Printf("failure encoding image path %s", p)
		}
		var img image.Image = igr.Image
		// what we should actually do is all of this stuff, including the write image, ahead ot time, into a data structure, and then just serve the contents in this handler
		var buffer bytes.Buffer
		//Plots[p].Writable() // new(bytes.Buffer) //this buffer is actually going to be part of the double-buffered type
		if err := png.Encode(&buffer, img); err != nil {
			log.Println("unable to encode image.")
		}
		Plots[p].Write(buffer.Bytes())
		Plots[p].Switch()
	}

}

func getsortedJams(map_ map[string]*JamStat) ([]string, int) {
	i := 0
	l := len(map_)
	keys := make([]string, l, l)
	for k := range map_ {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys, l
}

func drawPtsPerTeam( igr *imgg.ImageGraphics) bool {
	//igr := imgg.New(1024,768, color.RGBA{0xff,0xff,0xff,0xff}, nil, nil)
	//c is a chart.Chart type
	c := chart.ScatterChart{Title: "Total Points Per Team"}
	c.XRange.TicSetting.Mirror = 1
	c.YRange.TicSetting.Mirror = 1

	var x,y,y2 []float64
	var cc []string

	//something like
	jams, l := getsortedJams(Stats.Jams)
	if l < 1 { //we have no actual data yet, still initing probably
		x = []float64{0.0}
		y = []float64{0.0}
		y2 = []float64{0.0}
		cc = []string{"",""}
	} else {
		x = make([]float64,l,l)
		y = make([]float64,l,l)
		y2 = make([]float64,l,l)
		cc = make([]string,l+1, l+1)
		cc[0] = ""
		for i,j := range jams {
			x[i] = float64(i)
			y[i] = float64(Stats.Jams[j].TotalScores[0])
			y2[i] = float64(Stats.Jams[j].TotalScores[1])
			cc[i+1] = j
		}
	}
	// then team names are Teams[0].Name and Teams[1].Name
	// it's not clear if we can get the colours from the

	//x := []float64{1.,2.,3.,4.,5.}  //this is just an incrementing list of points (= jam number, continuing counting into p2 without resetting)
	//y := []float64{0.,0.,10.,20.,30.}
	//y2 := []float64{10.,20.,30.,30.,30.}
	//c.XRange.Category = []string{"","P1J1","P1J2","P1J3","P1J4","P2J1"} //if we stuff categories in the Axis before adding data, we can force "non-standard names" fpr the tics
	// we need the "zero" category to be empty, due to how tics are calculated for categorics, and the rest of categories need to map to X values exactly
	c.AddDataPair(Teams[0].Name, x, y, chart.PlotStyleLinesPoints, chart.Style{Symbol: '#', SymbolColor: color.NRGBA{0x00, 0xee, 0x00, 0xff}})
	c.AddDataPair(Teams[1].Name, x, y2, chart.PlotStyleLinesPoints, chart.Style{Symbol: '#', SymbolColor: color.NRGBA{0x00, 0xee, 0xff, 0xff}})
	c.Plot(igr)
	 //and switch the active pane post update
	//writeImage(w, &img)
	return true
}

func imgHandler(w http.ResponseWriter, r *http.Request) {
	// get path p
	p := r.URL.Path[5:] //the part of the path after /img/
	if _, prsnt := Plots[p]; !prsnt {
		//respond with 404
		w.WriteHeader(404)
		return
	}
	bufferb := Plots[p].Display()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(bufferb)))
	if _, err := w.Write(bufferb); err != nil {
		log.Println("unable to write image.")
	}
}

func htmlHandler(w http.ResponseWriter, r *http.Request) {
	var p string
	if r.URL.Path == "/" {
		p = "index.html"
	} else {
		p = r.URL.Path[1:] //remove leading /
	}
	if _, prsnt := Pages[p]; !prsnt {
		w.WriteHeader(404)
		return
	}
	log.Printf("Serving %s",p)
	bufferb := Pages[p].Display()
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(bufferb)))
	if _, err := w.Write(bufferb); err != nil {
		log.Println("unable to serve html")
	}
}
