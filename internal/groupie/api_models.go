package groupie

import "strings"

const (
	DefaultAPIBaseURL = "https://groupietrackers.herokuapp.com/api"

	apiPathArtists   = "/artists"
	apiPathLocations = "/locations"
	apiPathDates     = "/dates"
	apiPathRelation  = "/relation"
)

// Endpoints keeps every Groupie Tracker API URL in one place.
type Endpoints struct {
	BaseURL   string
	Root      string
	Artists   string
	Locations string
	Dates     string
	Relation  string
}

func NewEndpoints(baseURL string) Endpoints {
	baseURL = strings.TrimRight(baseURL, "/")
	return Endpoints{
		BaseURL:   baseURL,
		Root:      baseURL,
		Artists:   baseURL + apiPathArtists,
		Locations: baseURL + apiPathLocations,
		Dates:     baseURL + apiPathDates,
		Relation:  baseURL + apiPathRelation,
	}
}

type APILinks struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

type APIArtist struct {
	ID              int      `json:"id"`
	Image           string   `json:"image"`
	Name            string   `json:"name"`
	Members         []string `json:"members"`
	CreationDate    int      `json:"creationDate"`
	FirstAlbum      string   `json:"firstAlbum"`
	LocationsURL    string   `json:"locations"`
	ConcertDatesURL string   `json:"concertDates"`
	RelationsURL    string   `json:"relations"`
}

type APILocationsIndex struct {
	Index []APILocations `json:"index"`
}

type APILocations struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"`
}

type APIDatesIndex struct {
	Index []APIDates `json:"index"`
}

type APIDates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type APIRelationIndex struct {
	Index []APIRelation `json:"index"`
}

type APIRelation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}
