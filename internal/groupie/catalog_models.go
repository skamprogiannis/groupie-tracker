package groupie

type ArtistSummary struct {
	ID            int
	Name          string
	Image         string
	CreationDate  int
	FirstAlbum    string
	MemberSummary string
	MemberCount   int
	LocationCount int
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
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	CreationDate  int    `json:"creationDate"`
	FirstAlbum    string `json:"firstAlbum"`
	MemberSummary string `json:"memberSummary"`
	MemberCount   int    `json:"memberCount"`
	LocationCount int    `json:"locationCount"`
	MatchedDetail string `json:"matchedDetail"`
}

// SearchQuery carries the free-text query plus every filter and sort option the
// home page can send. Zero values mean "no constraint", so an empty SearchQuery
// returns the whole catalog in default order.
type SearchQuery struct {
	Text       string
	MinYear    int
	MaxYear    int
	MinMembers int
	MaxMembers int
	Country    string
	Sort       string // "name" (default), "newest", "oldest", "members"
}

// FilterOptions describes the bounds of the dataset so the UI can build filter
// controls (year range, member range, the list of countries) without hardcoding
// anything.
type FilterOptions struct {
	MinYear    int
	MaxYear    int
	MinMembers int
	MaxMembers int
	Countries  []string
}

type Catalog struct {
	artistsByID map[int]ArtistDetail
	artists     []ArtistDetail
}
