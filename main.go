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
	NFL       = "nfl"
	CFB       = "cfb"
	nflEvents = "https://sports.core.api.espn.com/v2/sports/football/leagues/nfl/events"
	cfbEvents = "https://sports.core.api.espn.com/v2/sports/football/leagues/college-football/events?groups=80&limit=500"

	// Use this link if trying to get a specific week:
	// "https://sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/2022/types/2/weeks/6/events"
)

func main() {

	nflEvent, err := fetchAndDecodeEvent(nflEvents)
	if err != nil {
		log.Fatal("Error fetching NFL event:", err)
	}

	cfbEvent, err := fetchAndDecodeEvent(cfbEvents)
	if err != nil {
		log.Fatal("Error fetching CFB event:", err)
	}

	nflEvent.EventType = NFL
	cfbEvent.EventType = CFB

	nflGames := fetchGames(nflEvent)
	cfbGames := fetchGames(cfbEvent)

	nflOutput := ProcessGame(nflGames)
	cfbOutput := ProcessGame(cfbGames)

	fileName := "NFL-Week-" + nflEvent.Meta.Parameters.Week[0] + ".txt"
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	filePath := filepath.Join(dir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println("failed to create file:", err)
		return
	}

	f.WriteString(fmt.Sprintf("NFL Week %v\n", nflEvent.Meta.Parameters.Week[0]))
	fmt.Printf("NFL Week %v\n", nflEvent.Meta.Parameters.Week[0])

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
	fmt.Printf("College Week %v  \n", cfbEvent.Meta.Parameters.Week[0])
	f.WriteString(fmt.Sprintf("College Week %v  \n", cfbEvent.Meta.Parameters.Week[0]))
	fmt.Println("-----------------")
	f.WriteString("-----------------\n")
	fmt.Println()
	f.WriteString("\n")
	for k, v := range cfbOutput {
		fmt.Println("Game ", k)
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
		log.Fatal("failed to parse time:", err)
	}
	loc, _ := time.LoadLocation("America/New_York")
	b := a.In(loc).Format("Monday, Jan-02-06 3:04PM")
	return b
}

func fetchAndDecodeEvent(url string) (Event, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Event{}, err
	}
	defer resp.Body.Close()

	var event Event
	err = json.NewDecoder(resp.Body).Decode(&event)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func fetchGames(event Event) []GameInfo {
	var games []GameInfo

	for _, item := range event.Items {
		resp, err := http.Get(item.Ref)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Fatal("Error fetching game link:", err)
		}
		defer resp.Body.Close()

		var game GameInfo
		err = json.NewDecoder(resp.Body).Decode(&game)
		if err != nil {
			log.Fatal("Error decoding game:", err)
		}

		var sl string
		if event.EventType == NFL {
			sl = fmt.Sprintf("http://sports.core.api.espn.com/v2/sports/football/leagues/nfl/events/%s/competitions/%s/odds?lang=en&region=us", game.ID, game.ID)
		} else if event.EventType == CFB {
			sl = fmt.Sprintf("http://sports.core.api.espn.com/v2/sports/football/leagues/college-football/events/%s/competitions/%s/odds?lang=en&region=us", game.ID, game.ID)
		}

		game.StatLink = sl

		games = append(games, game)
	}

	return games
}

func fetchGameStats(statURL string) (Stats, error) {
	resp, err := http.Get(statURL)
	if err != nil {
		fmt.Println("Can't find statURL: ", statURL)
		return Stats{}, fmt.Errorf("error fetching statURL: %w", err)
	}
	defer resp.Body.Close()

	var stats Stats
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return Stats{}, fmt.Errorf("error decoding statURL: %w", err)
	}

	return stats, nil
}

func ProcessGame(games []GameInfo) []Output {
	var output []Output
	for _, game := range games {
		stats, err := fetchGameStats(game.StatLink)
		if err != nil {
			log.Fatal("Failed to fetch game stats: ", err)
		}

		if len(stats.Items) == 0 {
			fmt.Println("No spread given. Skipping ", game.ShortName)
			continue

		}
		out := Output{
			Date:      FormatTime(game.Date),
			Shortname: game.ShortName,
			Spread:    stats.Items[0].Details,
			Name:      game.Name,
		}

		output = append(output, out)
	}
	return output
}
