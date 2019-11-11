package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
  "encoding/json"
	"sort"
	"strconv"
	"strings"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/vdobler/chart/imgg"
	"bytes"
	"net/http"
	//profiling hack - comment out when not profiling
	//"github.com/pkg/profile"
)

type reg_type struct {
  Action 	string  `json:"action"`
  Paths 	[]string `json:"paths"`
}

var addr = flag.String("addr", "localhost:8000", "http service address")
var Teams [2]Team //team global state
var Stats Stats_t //stats global state

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
	Plotter = make(map[string]func(*imgg.ImageGraphics)bool)
	Pager = make(map[string]func(*bytes.Buffer)bool)
	Plots = make(map[string]*Page_buffer)
	Pages = make(map[string]*Page_buffer)
	Plotter["scores.png"] = drawPtsPerTeam
	Plots["scores.png"] = NewPageBuffer()
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

	done := make(chan struct{})

	//channel used to signal the webserver side to rebuild its stuff
	//refreshsite := make(chan struct{})

  //build our registration request
	res := &reg_type{
    	Action: "Register",
    	Paths: []string{}}

  res.Paths = append(res.Paths, "ScoreBoard.Team(1)")
	res.Paths = append(res.Paths, "ScoreBoard.Team(2)")
  res.Paths = append(res.Paths, "ScoreBoard.Stats")
  // items of interest:

  //register all the items we're interested in, via JSON
  my_mesg, _  := json.Marshal(res)
  err = c.WriteMessage(websocket.TextMessage, my_mesg)
  if err != nil {
    	log.Println("Registration:", err)
    	return
  }


	go func() {
		defer close(done)
		for {
			var object map[string]interface{}
			_, message, err := c.ReadMessage()
			json.Unmarshal(message, &object)
			if err != nil {
				log.Println("read:", err)
				return
			}
			//grab a timestamp because we want to actually record when star passes and leads happen within a jam!
			update_time := time.Now()
			//extract the actual keys, if it's safe to do so (we have a state update, and it is not empty - so we know its type)
			if state,prs := object["state"]; prs && (state != nil) {
				i := 0
				updates := object["state"].(map[string]interface{})
				l := len(updates)
				keys := make([]string, l, l)
				for k := range updates {
					keys[i] = k
					i++
				}
				//we sort the keys so similar items are grouped, and list items are ordered, which makes a lot of update processing easier
				sort.Strings(keys)
				//and now we iterate through the keys, parsing them for updates (we split the keys at this point as it makes parsing easier)
				//this is going to be a deep parse tree, so we should probably hive it off into functions before it nests too much.

				for _, k := range keys {
						path := strings.Split(k,".") //path[0] will always be "Scoreboard"

						switch path[1] {
						case "Stats":
							switch path[2][0:3]{
							default: //don't know what this is so don't do anything
								continue
							case "Per": //Period / Jam sorted data, this should be Period(n).Jam(m).stuff
								periodnum, _ := strconv.Atoi(path[2][7:len(path[2])-1])
								jamnum, _  := strconv.Atoi(path[3][4:len(path[3])-1])
								jamid := fmt.Sprintf("P%dJ%d",periodnum,jamnum)
								if _, prsnt := Stats.Jams[jamid]; !prsnt {
									//make this jamid record before we do anything else - it needs to know its own place
									Stats.Jams[jamid] = &JamStat{}
									Stats.Jams[jamid]._init_(jamnum,periodnum)
								}
								switch path[4][0:3]{
								case "Tea": //team specific info for this jam - lineup positions, lead jammer status, etc
									teamnum, _ := strconv.Atoi(path[4][5:len(path[4])-1])
									teamnum -= 1
									switch path[5][:7] {
									case "LeadJam":
										//log.Printf("%v %v",jamid, teamnum)
										if updates[k].(string) == "Lead"	{ //or however we do that string comparison
										//set or unset lead jammer status
										//	log.Printf("Set Lead")
											Stats.Jams[jamid].Lead = teamnum
											Stats.Jams[jamid].LeadTime = update_time
										}
									case "JamScor":
										Stats.Jams[jamid].ScoreDeltas[teamnum] = int(updates[k].(float64))
									case "TotalSc":
										Stats.Jams[jamid].TotalScores[teamnum] = int(updates[k].(float64))
									case "Skater(": //we only care about Lineup here
										skaterid := path[5][7:43]
										if _,prsnt := Teams[teamnum].Skaters[skaterid]; !prsnt {
											//make a new Skater field for this new skater, if they didn't exist
											Teams[teamnum].Skaters[skaterid] = &Skater{} //is that how this works?
											Teams[teamnum].Skaters[skaterid]._init_()
										}
										switch path[6]	{
										case "Position":
											Stats.Jams[jamid].Lineups[teamnum][skaterid], _ = updates[k].(string)
										case "StarPass":
											Stats.Jams[jamid].StarPass[teamnum] = update_time
										}

									default: //we don't care about anything else
										continue
									}
								default: //else its timing data for the jam, so put it in the stats structure neat
									//ADDTIMEDATA(periodnum,jamnum,type) = updates[k].(int)
								}
							}
						default: //its a team, so we nest into that
							teamnum, _ := strconv.Atoi(path[1][5:len(path[1])-1])
							teamnum -= 1
							switch path[2][0:2] {
							case "Sk": //skater record
								//eg 87d25fe6-c914-434b-bcc2-8c02f3da9cae (36 char uuid)
								skaterid := path[2][7:43]
								if _,prsnt := Teams[teamnum].Skaters[skaterid]; !prsnt {
									//make a new Skater field for this new skater, if they didn't exist
									Teams[teamnum].Skaters[skaterid] = &Skater{} //is that how this works?
									Teams[teamnum].Skaters[skaterid]._init_()
								}
								switch path[3][0:3] {//parse skater stuff
								case "Pen": //penalty data
									if path[3][7:8] != "(" {
										continue
									}
									pennum, _ := strconv.Atoi(path[3][8:len(path[3])-1])
									pennum -= 1
									//log.Printf("%v",path)
									switch path[4]{
									case "Code":
										//log.Printf("Penalty: %v %v %v",pennum,skaterid,teamnum)
										Teams[teamnum].Skaters[skaterid].Penalties[pennum].Symbol = updates[k].(string)[0]
									case "Period":
										Teams[teamnum].Skaters[skaterid].Penalties[pennum].Period = int(updates[k].(float64))
									case "Jam":
										Teams[teamnum].Skaters[skaterid].Penalties[pennum].Jam = int(updates[k].(float64))
									default:
										continue
									}
								case "Num":
									//log.Printf("Skater Num %v %v %v", teamnum, skaterid, updates[k].(string))
									Teams[teamnum].Skaters[skaterid].Number, _ = updates[k].(string)
								case "Nam":
									Teams[teamnum].Skaters[skaterid].Name, _ = updates[k].(string)
								default: //don't care about other possibilities
									continue
								}
							case "Na": // name
								Teams[teamnum].Name, _ = updates[k].(string)
							case "Score":
								//Teams[teamnum].Score = updates[k].(string)
							}
						}

				}
				//log.Printf("recv: %v", keys)
			}

      //messages are JSON formatted like:
      // {stateID: num , "state": {"item":value, "item":value etc}}

			//push an interrupt down the channel to the webserver to make it rebuild its stuff
			// we should set bits appropriately and only rebuild the things we need to here, but this bulk process will do for a start
			writeHTMLwithPaths(getPageBuffers(Pages))
			drawPlotswithPaths(getPageBuffers(Plots))
		}
	}()

	//***********************************************************************************************
	//goroutine for WEBSERVER
	go func() {
		//
		//http.Handle("/img",http.StripPrefix("/img/",http.))
		http.HandleFunc("/img/", imgHandler)
		//http.HandleFunc("/red/", redHandler)
		http.HandleFunc("/", htmlHandler)
		log.Println("Listening on 80")
		err := http.ListenAndServe(":80", nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
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
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}
