package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"image/color"
	"github.com/gorilla/websocket"
	"github.com/aoanla/chart/imgg"
	"bytes"
	//profiling hack - comment out when not profiling
	//"github.com/pkg/profile"
)

type reg_type struct {
  Action 	string  `json:"action"`
  Paths 	[]string `json:"paths"`
}

var addr = flag.String("addr", "localhost:8000", "URL of scoreboard")
var port = flag.String("port", ":80", "port for local web server")
var Teams [2]Team //team global state
var Stats Stats_t //stats global state
var ImgBuff *imgg.ImageGraphics //image buffer to reduce memory footprint + allocations

func getPageBuffers(map_ map[string]*Page_buffer) []string {
	i := 0
	l := len(map_)
	keys := make([]string, l, l)
	for k := range map_ {
		keys[i] = k
		i++
	}
	//sort.Strings(keys)
	return keys
}

func main() {
	//profiling
	//defer profile.Start(profile.MemProfile).Stop()
	//profiling
	flag.Parse()
	log.SetFlags(0)

	//init our mappings
	ImgBuff = imgg.New(1024,768, color.RGBA{0xff,0xff,0xff,0xff}, nil, nil)
	Plotter = make(map[string]func(*imgg.ImageGraphics)bool)
	Pager = make(map[string]func(*bytes.Buffer)bool)
	Plots = make(map[string]*Page_buffer)
	Pages = make(map[string]*Page_buffer)
	Plotter["scores.png"] = drawPtsPerTeam
	Plots["scores.png"] = NewPageBuffer()
	Plotter["dscores.png"] = drawDeltaPtsPerTeam
	Plots["dscores.png"] = NewPageBuffer()
	Plotter["lead.png"] = drawLeadsPerTeam
	Plots["lead.png"] = NewPageBuffer()
	Pager["whiteboard.html"] = page_PenaltyWhiteboard
	Pages["whiteboard.html"] = NewPageBuffer()
	Pager["index.html"] = page_Index
	Pages["index.html"] = NewPageBuffer()
	writeHTMLwithPaths(getPageBuffers(Pages))
	drawPlotswithPaths(getPageBuffers(Plots))


	Teams[0].Skaters = make(map[string]*Skater)
	Teams[1].Skaters = make(map[string]*Skater)
	Stats.Jams = make(map[string]*JamStat)


	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/WS/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	//*************************************************************************************
	//go routine for client, needs the websocket connection to work
	go scoreboard_client(c)

	//***********************************************************************************************
	//goroutine for WEBSERVER
	go web_server(port)

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		//case <-done:
		//	return
		case <-ticker.C:
      ping := []byte(`{"action":"Ping","data":0}`)
			err := c.WriteMessage(websocket.TextMessage, ping)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			//select {
			//case <-done:
			//case <-time.After(time.Second):
			//}
			//return
		}
	}

}
