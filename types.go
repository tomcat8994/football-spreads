package main

type EventType string

type Output struct {
	Date      string
	Shortname string
	Spread    string
	Name      string
}

type Stats struct {
	Items []struct {
		Ref     string `json:"$ref"`
		Details string `json:"details"`
	} `json:"items"`
}

type GameInfo struct {
	Ref       string `json:"$ref"`
	ID        string `json:"id"`
	Date      string `json:"date"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	StatLink  string
}

type Event struct {
	Meta struct {
		Parameters struct {
			Week        []string `json:"week"`
			Season      []string `json:"season"`
			Seasontypes []string `json:"seasontypes"`
		} `json:"parameters"`
	} `json:"$meta"`
	Items []struct {
		Ref string `json:"$ref"`
	} `json:"items"`
	EventType EventType
}
