package groupie

import (
	"fmt"
	"sort"
	"strings"
)

func NewCatalog(data APIData) (*Catalog, error) {
	locationsByID, err := indexLocations(data.Locations)
	if err != nil {
		return nil, err
	}
	datesByID, err := indexDates(data.Dates)
	if err != nil {
		return nil, err
	}
	relationsByID, err := indexRelations(data.Relations)
	if err != nil {
		return nil, err
	}

	artists := make([]ArtistDetail, 0, len(data.Artists))
	artistsByID := make(map[int]ArtistDetail, len(data.Artists))
	for _, artist := range data.Artists {
		if _, exists := artistsByID[artist.ID]; exists {
			return nil, fmt.Errorf("artist %d appears more than once", artist.ID)
		}

		detail := ArtistDetail{
			ID:             artist.ID,
			Name:           artist.Name,
			Image:          artist.Image,
			Members:        cloneStrings(artist.Members),
			CreationDate:   artist.CreationDate,
			FirstAlbum:     artist.FirstAlbum,
			Locations:      cloneStrings(locationsByID[artist.ID].Locations),
			Dates:          cloneStrings(datesByID[artist.ID].Dates),
			DatesLocations: cloneDatesLocations(relationsByID[artist.ID].DatesLocations),
		}
		artists = append(artists, detail)
		artistsByID[artist.ID] = detail
	}

	sort.Slice(artists, func(i, j int) bool {
		return artists[i].ID < artists[j].ID
	})

	return &Catalog{
		artistsByID: artistsByID,
		artists:     artists,
	}, nil
}

func (c *Catalog) Artists() []ArtistSummary {
	if c == nil {
		return nil
	}

	summaries := make([]ArtistSummary, 0, len(c.artists))
	for _, artist := range c.artists {
		summaries = append(summaries, summarize(artist))
	}
	return summaries
}

func summarize(artist ArtistDetail) ArtistSummary {
	return ArtistSummary{
		ID:            artist.ID,
		Name:          artist.Name,
		Image:         artist.Image,
		CreationDate:  artist.CreationDate,
		FirstAlbum:    artist.FirstAlbum,
		MemberSummary: strings.Join(artist.Members, ", "),
		MemberCount:   len(artist.Members),
		LocationCount: len(artist.Locations),
	}
}

// Facets reports the dataset bounds used to build the filter controls: the
// creation-year range, the first-album-year range, the member-count range, the
// sorted list of distinct concert countries, and every concert location slug.
func (c *Catalog) Facets() FilterOptions {
	if c == nil || len(c.artists) == 0 {
		return FilterOptions{}
	}

	firstAlbum := firstAlbumYear(c.artists[0].FirstAlbum)
	opts := FilterOptions{
		MinYear:      c.artists[0].CreationDate,
		MaxYear:      c.artists[0].CreationDate,
		MinAlbumYear: firstAlbum,
		MaxAlbumYear: firstAlbum,
		MinMembers:   len(c.artists[0].Members),
		MaxMembers:   len(c.artists[0].Members),
	}
	countries := make(map[string]struct{})
	locations := make(map[string]struct{})
	for _, artist := range c.artists {
		if artist.CreationDate < opts.MinYear {
			opts.MinYear = artist.CreationDate
		}
		if artist.CreationDate > opts.MaxYear {
			opts.MaxYear = artist.CreationDate
		}
		if year := firstAlbumYear(artist.FirstAlbum); year > 0 {
			if opts.MinAlbumYear == 0 || year < opts.MinAlbumYear {
				opts.MinAlbumYear = year
			}
			if year > opts.MaxAlbumYear {
				opts.MaxAlbumYear = year
			}
		}
		if n := len(artist.Members); n < opts.MinMembers {
			opts.MinMembers = n
		} else if n > opts.MaxMembers {
			opts.MaxMembers = n
		}
		for _, location := range artist.Locations {
			locations[location] = struct{}{}
			if country := countryOf(location); country != "" {
				countries[country] = struct{}{}
			}
		}
	}

	opts.Countries = sortedKeys(countries)
	opts.Locations = sortedKeys(locations)
	return opts
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// countryOf extracts the country slug from a "city-country" location string,
// e.g. "los_angeles-usa" -> "usa". It returns the whole string when there is no
// separator.
func countryOf(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	if idx := strings.LastIndex(location, "-"); idx >= 0 && idx < len(location)-1 {
		return location[idx+1:]
	}
	return location
}

func (c *Catalog) ArtistByID(id int) (ArtistDetail, bool) {
	if c == nil {
		return ArtistDetail{}, false
	}

	artist, ok := c.artistsByID[id]
	if !ok {
		return ArtistDetail{}, false
	}
	return cloneArtistDetail(artist), true
}

func indexLocations(locations []APILocations) (map[int]APILocations, error) {
	byID := make(map[int]APILocations, len(locations))
	for _, location := range locations {
		if _, exists := byID[location.ID]; exists {
			return nil, fmt.Errorf("locations %d appears more than once", location.ID)
		}
		byID[location.ID] = location
	}
	return byID, nil
}

func indexDates(dates []APIDates) (map[int]APIDates, error) {
	byID := make(map[int]APIDates, len(dates))
	for _, date := range dates {
		if _, exists := byID[date.ID]; exists {
			return nil, fmt.Errorf("dates %d appears more than once", date.ID)
		}
		byID[date.ID] = date
	}
	return byID, nil
}

func indexRelations(relations []APIRelation) (map[int]APIRelation, error) {
	byID := make(map[int]APIRelation, len(relations))
	for _, relation := range relations {
		if _, exists := byID[relation.ID]; exists {
			return nil, fmt.Errorf("relation %d appears more than once", relation.ID)
		}
		byID[relation.ID] = relation
	}
	return byID, nil
}

func cloneArtistDetail(artist ArtistDetail) ArtistDetail {
	return ArtistDetail{
		ID:             artist.ID,
		Name:           artist.Name,
		Image:          artist.Image,
		Members:        cloneStrings(artist.Members),
		CreationDate:   artist.CreationDate,
		FirstAlbum:     artist.FirstAlbum,
		Locations:      cloneStrings(artist.Locations),
		Dates:          cloneStrings(artist.Dates),
		DatesLocations: cloneDatesLocations(artist.DatesLocations),
	}
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneDatesLocations(values map[string][]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string][]string, len(values))
	for location, dates := range values {
		cloned[location] = cloneStrings(dates)
	}
	return cloned
}
