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

type Catalog struct {
	artistsByID map[int]ArtistDetail
	artists     []ArtistDetail
}
