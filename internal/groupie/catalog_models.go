package groupie

type ArtistSummary struct {
	ID           int
	Name         string
	Image        string
	CreationDate int
	FirstAlbum   string
}

type ArtistDetail struct {
	ID             int
	Name           string
	Image          string
	Members        []string
	CreationDate   int
	FirstAlbum     string
	Locations      []string
	Dates          []string
	DatesLocations map[string][]string
}

type SearchResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Image        string `json:"image"`
	CreationDate int    `json:"creationDate"`
	FirstAlbum   string `json:"firstAlbum"`
}

type Catalog struct {
	artistsByID map[int]ArtistDetail
	artists     []ArtistDetail
}
