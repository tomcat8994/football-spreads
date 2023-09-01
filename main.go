package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	nflEvents = "https://sports.core.api.espn.com/v2/sports/football/leagues/nfl/events"
	cfbEvents = "https://sports.core.api.espn.com/v2/sports/football/leagues/college-football/events"

	// Use this link if trying to get a specific week:
	// "https://sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/2022/types/2/weeks/6/events"
)

func main() {

	var NflEvent, CfbEvent Event

	resp, err := http.Get(nflEvents)
	if err != nil {
		log.Fatal("Err with NFL Event GET request: ", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&NflEvent)
	if err != nil {
		log.Fatal("Err unmarshalling: ", err)
	}

	cfgResp, err := http.Get(cfbEvents)
	if err != nil {
		log.Fatal("Err With CFB Event GET Request: ", err)
	}
	defer cfgResp.Body.Close()

	err = json.NewDecoder(cfgResp.Body).Decode(&CfbEvent)
	if err != nil {
		log.Fatal("Err decoding CfbEvent: ", err)
	}

	var nflGameLinks, cfbGameLinks []string

	for _, v := range NflEvent.Items {
		nflGameLinks = append(nflGameLinks, v.Ref)
	}

	for _, v := range CfbEvent.Items {
		cfbGameLinks = append(cfbGameLinks, v.Ref)
	}

	var ngames, cgames []GameInfo

	for k, link := range nflGameLinks {
		resp, err := http.Get(link)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Err grabbing link ", k)
			continue
		}
		var game GameInfo
		err = json.NewDecoder(resp.Body).Decode(&game)
		if err != nil {
			log.Fatal("Err decoding game: ", err)
		}
		sl := fmt.Sprintf("http://sports.core.api.espn.com/v2/sports/football/leagues/nfl/events/%s/competitions/%s/odds?lang=en&region=us", game.ID, game.ID)
		game.StatLink = sl
		ngames = append(ngames, game)
		resp.Body.Close()
	}

	for k, link := range cfbGameLinks {
		resp, err := http.Get(link)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("Err grabbing cfb link (index): ", k)
			continue
		}
		var cgame GameInfo
		err = json.NewDecoder(resp.Body).Decode(&cgame)
		if err != nil {
			log.Fatal("Err decoding cgame: ", err)
		}
		sl := fmt.Sprintf("http://sports.core.api.espn.com/v2/sports/football/leagues/college-football/events/%s/competitions/%s/odds?lang=en&region=us", cgame.ID, cgame.ID)
		cgame.StatLink = sl
		cgames = append(cgames, cgame)
		resp.Body.Close()
	}

	var nflOutput, cfbOutput []Output

	for _, v := range ngames {
		var st Stats
		resp1, err := http.Get(v.StatLink)
		if err != nil {
			log.Fatal("err getting spread link: ", err)
		}
		err = json.NewDecoder(resp1.Body).Decode(&st)
		if err != nil {
			log.Fatal("Err decoding game: ", err)
		}
		var out Output

		out.Date = FormatTime(v.Date)
		out.Shortname = v.ShortName
		out.Spread = st.Items[0].Details

		nflOutput = append(nflOutput, out)
	}

	for _, v := range cgames {
		var st Stats
		resp2, err := http.Get(v.StatLink)
		if err != nil {
			log.Fatal("error getting spread link for cfb: ", err)
		}
		err = json.NewDecoder(resp2.Body).Decode(&st)
		if err != nil {
			log.Fatal("Err decoding college game: ", err)
		}
		var out Output
		if len(st.Items) == 0 {
			continue
		}

		out.Date = FormatTime(v.Date)
		out.Shortname = v.ShortName
		out.Spread = st.Items[0].Details
		out.Name = v.Name

		cfbOutput = append(cfbOutput, out)

	}

	// Get the user home directory
	userHomeDir, _ := os.UserHomeDir()

	// Create a path to the desired file
	filePath := filepath.Join(userHomeDir, "Desktop", "NFL-Week-"+NflEvent.Meta.Parameters.Week[0]+".txt")

	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write some data to the file
	f.WriteString(fmt.Sprintf("NFL Week %v\n", NflEvent.Meta.Parameters.Week[0]))
	fmt.Printf("NFL Week %v\n", NflEvent.Meta.Parameters.Week[0])

	f.WriteString("-------------------\n")
	fmt.Println("-------------------")

	for k, v := range nflOutput {
		fmt.Println(v.Date)
		f.WriteString(v.Date + "\n")
		fmt.Println(v.Shortname)
		f.WriteString(v.Shortname + "\n")
		fmt.Println(v.Spread)
		f.WriteString(v.Spread + "\n")
		if k == len(nflOutput)-1 {
			fmt.Println()
			f.WriteString("\n")
			break
		}
		fmt.Println("----------")
		f.WriteString("----------\n")
	}
	fmt.Println("-----------------")
	f.WriteString("-----------------\n")
	fmt.Printf("College Week %v  \n", CfbEvent.Meta.Parameters.Week[0])
	f.WriteString(fmt.Sprintf("College Week %v  \n", CfbEvent.Meta.Parameters.Week[0]))
	fmt.Println("-----------------")
	f.WriteString("-----------------\n")
	fmt.Println()
	f.WriteString("\n")
	for _, v := range cfbOutput {
		fmt.Println(v.Date)
		f.WriteString(v.Date + "\n")
		fmt.Println(v.Shortname)
		fmt.Println(v.Name)
		f.WriteString(v.Shortname + "\n")
		fmt.Println(v.Spread)
		f.WriteString(v.Spread + "\n")
		fmt.Println("----------")
		f.WriteString("----------\n")
	}

}

func FormatTime(s string) string {
	layout := "2006-01-02T15:05Z"

	a, err := time.Parse(layout, s)
	if err != nil {
		log.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/New_York")
	b := a.In(loc).Format("Monday, Jan-02-06 3:04PM")
	return b
}
