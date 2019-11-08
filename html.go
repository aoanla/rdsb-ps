package main

//html page writing stuff, including some utility functions
import "bytes"
import "sort"
import "strings"
import "fmt"
import "strconv"

// html writers have this signature
//func(buff bytes.Buffer)bool

//Skater customised version of getsortedkeys - we should probably just make this generic by letting someone pass a sort function
func getsortedSkaters(map_ map[string]*Skater) ([]string, int) {
	i := 0
	l := len(map_)
	keys := make([]string, l, l)
	for k := range map_ {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i,j int) bool {return map_[keys[i]].Name < map_[keys[j]].Name})
	return keys, l
}

func page_Index(buff *bytes.Buffer) bool {
  buff.WriteString("<html><body><a href='whiteboard.html'>Penalty Whiteboard</a><br /><a href='img/scores.png'>Total Scores Graph</a></body></html>")
  return true
}


func page_PenaltyWhiteboard( buff *bytes.Buffer) bool {
  var s strings.Builder
  s.WriteString("<html><style>body {width: 100%; font-weight: bold;} div {background-color: purple; color: white;} table { border-collapse: collapse; width: 100%; font-weight: bold; text-align: center; } thead { background-color: thistle; } table tr:nth-child(even) { background-color: #F7F7FF; } tr { border: 3px solid purple; } td { border: 1px solid grey; } .v1 { color: #009933; } .v2 { color: #009933; } .v3 { color: #009933; } .v4 { color: #996600; } .v5 { color: #996600; } .v6 { color: #990000; } .v7 { color: #990000; } .v8 { color: #990000; } .v9 { color: #990000; }  </style><body>")
  var s1,s2 strings.Builder //for the parallel strings here in penalties
  for i := 0; i<2; i++ {
    skaters,l := getsortedSkaters(Teams[i].Skaters)
    s.WriteString("<div>Penalties: ")
    s.WriteString(Teams[i].Name)
    s.WriteString("</div><table><thead><tr><td>S#</td><td>P1</td><td>P2</td><td>P3</td><td>P4</td><td>P5</td><td>P6</td><td>P7</td><td>P8</td><td>P9</td><td>p#</td></tr></thead>")

    if l < 1 { continue } //avoid iterating empty list if this is in setup

    for _, ss := range skaters {
      s1.WriteString("'><td>") //fragment because we need to customise the tr class
      s1.WriteString(Teams[i].Skaters[ss].Number)
      s1.WriteString("</td><td>")
      s2.WriteString("'><td>")
      s2.WriteString(Teams[i].Skaters[ss].Name)
      s2.WriteString("</td><td>")
      count := 0
      for _, sp := range Teams[i].Skaters[ss].Penalties {
        if sp.Symbol != '\000' {
          s1.WriteByte(sp.Symbol)
          s2.WriteString(fmt.Sprintf("%d:%d",sp.Period,sp.Jam))
          count += 1
        }
        s1.WriteString("</td><td>")
        s2.WriteString("</td><td>")
      }
      s1.WriteString(strconv.Itoa(count)) //penaty count on last line
      s1.WriteString("</td>")
      s2.WriteString("</td>")
      s.WriteString("<tr class='top v") //topline class for customisation
      s.WriteString(strconv.Itoa(count))
      s.WriteString(s1.String())
      s.WriteString("</tr>")
      s.WriteString("<tr class='bot v") //bottomline class for customisation
      s.WriteString(strconv.Itoa(count))
      s.WriteString(s2.String())
      s.WriteString("</tr>")
      s1.Reset()
      s2.Reset()
    }
    //end of skaters, close table
    s.WriteString("</table>")
  }
  //end of teams, close html
  s.WriteString("</body></html>")
  buff.WriteString(s.String())
  return true
}
