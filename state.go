package main
//utility stuff

//import "bytes"
import "time"
//import "log"

//type colourpair struct{
//	BG color
//	FG color
//}

type Penalty struct {
  Symbol byte
  Jam int
  Period int
}

type Skater struct {
  Name string
  Number string
  Penalties []Penalty
}

func (s *Skater) _init_(){
  //s.Name = name
  //s.Number = number
  s.Penalties = make([]Penalty,9,10)
  //for i := 0 ; i<9; i++ {
  //  s.Penalties[i].Code = ''
  //  s.Penalties[i].Jam = 0
  //  s.Penalties[i].Period = 0
  //}
}

type Team struct {
	Skaters 	map[string]*Skater
	//Colours 	colorpair
	Name			string
}

//container for stats per jam
type JamStat struct{
  Jam int
  Period int
  Start time.Time //the recorded timestamp of Start, for reference for offsets
  StartClock int //the PeriodClock returned offset of Start (the scoreboard's value)
  End time.Time
  EndClock int
  Lead  int //team with lead, -1 if none
  LeadTime time.Time //time offset of Lead, from the time we got the update
  ScoreDeltas []int //score deltas [scored in *this jam*] for each team
  TotalScores []int //total scores [score at end of this jam] for each team
  Lineups     []map[string]string  //map from position ("Jammer","Pivot" etc) to Skateruuid for each team
  StarPass    []time.Time //times of registered star pass, per team (nil if none happened for that team)
}

func (js *JamStat) _init_(jamnum,periodnum int) {
  js.Jam = jamnum
  js.Period = periodnum
  js.Start = time.Now() //is this a reasonable assumption?
  js.Lead = -1
  js.ScoreDeltas = make([]int, 2, 2)
  js.TotalScores = make([]int, 2, 2)
  js.Lineups = make([]map[string]string, 2, 2)
	js.Lineups[0] = make(map[string]string)
	js.Lineups[1] = make(map[string]string)
  js.StarPass = make([]time.Time, 2, 2)
}

type Stats_t struct {
  Jams map[string]*JamStat
}


//webserver "in memory" types: these are all created as double-buffers, and we render to the off-screen one to remove need for mutex
// we toggle active to the "clean" page each time
type Page_buffer struct {
 	page [2][]byte
	active int
}

func NewPageBuffer() *Page_buffer {
	var pb = &Page_buffer{};
	pb.page[0] = make([]byte,10,10)
	pb.page[1] = make([]byte,10,10)
	return pb
}

//get a writable buffer (the non-visible one)
func (pg *Page_buffer) Write(input []byte) {
  pg.page[1 - pg.active] = input
}

//switch the buffers
func (pg *Page_buffer) Switch() bool {
		pg.active = 1 - pg.active
		return true
}

//get the current readable buffer (the visible one)
func (pg *Page_buffer) Display() []byte {
	//log.Printf("Serving active page: %v", pg.active)
	//log.Printf("Page state: %v", pg.page)
	return pg.page[pg.active] //need to init these guys with make or similar (and also make them before they're first looked at)
}
