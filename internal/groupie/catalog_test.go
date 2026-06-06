package groupie

import "testing"

func TestCatalogListsAllArtistsInIDOrder(t *testing.T) {
	catalog := newTestCatalog(t)

	artists := catalog.Artists()
	if len(artists) != 4 {
		t.Fatalf("len(Artists()) = %d, want 4", len(artists))
	}
	wantIDs := []int{1, 2, 4, 30}
	for i, wantID := range wantIDs {
		if artists[i].ID != wantID {
			t.Fatalf("artists[%d].ID = %d, want %d", i, artists[i].ID, wantID)
		}
	}
}

func TestCatalogReturnsArtistByID(t *testing.T) {
	catalog := newTestCatalog(t)

	artist, ok := catalog.ArtistByID(2)
	if !ok {
		t.Fatal("ArtistByID(2) ok = false, want true")
	}
	if artist.Name != "Gorillaz" || artist.FirstAlbum != "26-03-2001" {
		t.Fatalf("artist = %#v, want Gorillaz detail", artist)
	}
}

func TestCatalogReturnsFalseForMissingArtist(t *testing.T) {
	catalog := newTestCatalog(t)

	_, ok := catalog.ArtistByID(999)
	if ok {
		t.Fatal("ArtistByID(999) ok = true, want false")
	}
}

func TestCatalogIncludesAuditData(t *testing.T) {
	catalog := newTestCatalog(t)

	queen, ok := catalog.ArtistByID(1)
	if !ok {
		t.Fatal("Queen is missing from catalog")
	}
	assertContainsAll(t, queen.Members, []string{
		"Freddie Mercury",
		"Brian May",
		"John Daecon",
		"Roger Meddows-Taylor",
		"Mike Grose",
		"Barry Mitchell",
		"Doug Fogie",
	})

	gorillaz, ok := catalog.ArtistByID(2)
	if !ok {
		t.Fatal("Gorillaz is missing from catalog")
	}
	if gorillaz.FirstAlbum != "26-03-2001" {
		t.Fatalf("Gorillaz first album = %q, want 26-03-2001", gorillaz.FirstAlbum)
	}

	travisScott, ok := catalog.ArtistByID(30)
	if !ok {
		t.Fatal("Travis Scott is missing from catalog")
	}
	assertContainsAll(t, travisScott.Locations, []string{
		"santiago-chile",
		"sao_paulo-brazil",
		"los_angeles-usa",
		"houston-usa",
		"atlanta-usa",
		"new_orleans-usa",
		"philadelphia-usa",
		"london-uk",
		"frauenfeld-switzerland",
		"turku-finland",
	})

	fooFighters, ok := catalog.ArtistByID(4)
	if !ok {
		t.Fatal("Foo Fighters are missing from catalog")
	}
	assertContainsAll(t, fooFighters.Members, []string{
		"Dave Grohl",
		"Nate Mendel",
		"Taylor Hawkins",
		"Chris Shiflett",
		"Pat Smear",
		"Rami Jaffee",
	})
}

func TestCatalogSearchMatchesJoinedFields(t *testing.T) {
	catalog := newTestCatalog(t)

	assertSingleSearchResult(t, catalog.Search(SearchQuery{Text: "freddie"}), "Queen")
	assertSingleSearchResult(t, catalog.Search(SearchQuery{Text: "26-03-2001"}), "Gorillaz")
	assertSingleSearchResult(t, catalog.Search(SearchQuery{Text: "turku-finland"}), "Travis Scott")
	assertSingleSearchResult(t, catalog.Search(SearchQuery{Text: "Rami Jaffee"}), "Foo Fighters")
}

func TestCatalogSearchSortsByName(t *testing.T) {
	catalog := newTestCatalog(t)

	results := catalog.Search(SearchQuery{})
	wantOrder := []string{"Foo Fighters", "Gorillaz", "Queen", "Travis Scott"}
	if len(results) != len(wantOrder) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(wantOrder))
	}
	for i, want := range wantOrder {
		if results[i].Name != want {
			t.Fatalf("results[%d].Name = %q, want %q", i, results[i].Name, want)
		}
	}
}

func TestCatalogSortNewestAndMembers(t *testing.T) {
	catalog := newTestCatalog(t)

	newest := catalog.Search(SearchQuery{Sort: "newest"})
	if newest[0].Name != "Travis Scott" {
		t.Fatalf("newest[0] = %q, want Travis Scott (2008)", newest[0].Name)
	}

	byMembers := catalog.Search(SearchQuery{Sort: "members"})
	if byMembers[0].Name != "Queen" {
		t.Fatalf("members[0] = %q, want Queen (7 members)", byMembers[0].Name)
	}
	if byMembers[0].MemberCount != 7 {
		t.Fatalf("Queen MemberCount = %d, want 7", byMembers[0].MemberCount)
	}
}

func TestCatalogFilters(t *testing.T) {
	catalog := newTestCatalog(t)

	// Year range: Queen (1970) and Foo Fighters (1994) formed by 1995; Gorillaz
	// (1998) and Travis Scott (2008) are excluded.
	byYear := catalog.Search(SearchQuery{MaxYear: 1995})
	assertResultNames(t, byYear, []string{"Foo Fighters", "Queen"})

	// Member count: bands with at least 6 members are Queen (7) and Foo Fighters (6).
	byMembers := catalog.Search(SearchQuery{MinMembers: 6})
	assertResultNames(t, byMembers, []string{"Foo Fighters", "Queen"})

	// Country: only Travis Scott played in Finland.
	byCountry := catalog.Search(SearchQuery{Country: "finland"})
	assertResultNames(t, byCountry, []string{"Travis Scott"})

	// Filters compose with text search.
	none := catalog.Search(SearchQuery{Text: "queen", MinYear: 2000})
	if len(none) != 0 {
		t.Fatalf("Queen formed in 1970 should be excluded by MinYear 2000, got %#v", none)
	}
}

func TestCatalogFacets(t *testing.T) {
	catalog := newTestCatalog(t)

	facets := catalog.Facets()
	if facets.MinYear != 1970 || facets.MaxYear != 2008 {
		t.Fatalf("year range = %d-%d, want 1970-2008", facets.MinYear, facets.MaxYear)
	}
	if facets.MaxMembers != 7 {
		t.Fatalf("MaxMembers = %d, want 7", facets.MaxMembers)
	}
	if !containsString(facets.Countries, "finland") || !containsString(facets.Countries, "usa") {
		t.Fatalf("countries = %#v, want finland and usa", facets.Countries)
	}
	for i := 1; i < len(facets.Countries); i++ {
		if facets.Countries[i-1] > facets.Countries[i] {
			t.Fatalf("countries not sorted: %#v", facets.Countries)
		}
	}
}

func TestCatalogReturnsDefensiveCopies(t *testing.T) {
	catalog := newTestCatalog(t)

	artist, ok := catalog.ArtistByID(1)
	if !ok {
		t.Fatal("Queen is missing from catalog")
	}
	artist.Members[0] = "mutated"
	artist.DatesLocations["london-uk"][0] = "mutated"

	artist, ok = catalog.ArtistByID(1)
	if !ok {
		t.Fatal("Queen is missing from catalog after mutation")
	}
	if artist.Members[0] == "mutated" {
		t.Fatal("ArtistByID returned shared member slice")
	}
	if artist.DatesLocations["london-uk"][0] == "mutated" {
		t.Fatal("ArtistByID returned shared relation map")
	}
}

func newTestCatalog(t *testing.T) *Catalog {
	t.Helper()

	catalog, err := NewCatalog(testAPIData())
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}
	return catalog
}

func assertSingleSearchResult(t *testing.T, results []SearchResult, wantName string) {
	t.Helper()

	if len(results) != 1 {
		t.Fatalf("len(Search()) = %d, want 1 for %q: %#v", len(results), wantName, results)
	}
	if results[0].Name != wantName {
		t.Fatalf("Search()[0].Name = %q, want %q", results[0].Name, wantName)
	}
}

func assertResultNames(t *testing.T, results []SearchResult, want []string) {
	t.Helper()

	got := make([]string, len(results))
	for i, r := range results {
		got[i] = r.Name
	}
	if len(got) != len(want) {
		t.Fatalf("result names = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("result names = %v, want %v", got, want)
		}
	}
}

func assertContainsAll(t *testing.T, values []string, wants []string) {
	t.Helper()

	for _, want := range wants {
		if !containsString(values, want) {
			t.Fatalf("%q missing from %#v", want, values)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func testAPIData() APIData {
	return APIData{
		Artists: []APIArtist{
			{
				ID:           30,
				Name:         "Travis Scott",
				Image:        "travis.jpeg",
				Members:      []string{"Travis Scott"},
				CreationDate: 2008,
				FirstAlbum:   "04-09-2015",
			},
			{
				ID:           1,
				Name:         "Queen",
				Image:        "queen.jpeg",
				Members:      []string{"Freddie Mercury", "Brian May", "John Daecon", "Roger Meddows-Taylor", "Mike Grose", "Barry Mitchell", "Doug Fogie"},
				CreationDate: 1970,
				FirstAlbum:   "14-12-1973",
			},
			{
				ID:           4,
				Name:         "Foo Fighters",
				Image:        "foo.jpeg",
				Members:      []string{"Dave Grohl", "Nate Mendel", "Taylor Hawkins", "Chris Shiflett", "Pat Smear", "Rami Jaffee"},
				CreationDate: 1994,
				FirstAlbum:   "04-07-1995",
			},
			{
				ID:           2,
				Name:         "Gorillaz",
				Image:        "gorillaz.jpeg",
				Members:      []string{"Damon Albarn", "Jamie Hewlett"},
				CreationDate: 1998,
				FirstAlbum:   "26-03-2001",
			},
		},
		Locations: []APILocations{
			{ID: 1, Locations: []string{"london-uk"}},
			{ID: 2, Locations: []string{"paris-france"}},
			{ID: 4, Locations: []string{"seattle-usa"}},
			{ID: 30, Locations: []string{"santiago-chile", "sao_paulo-brazil", "los_angeles-usa", "houston-usa", "atlanta-usa", "new_orleans-usa", "philadelphia-usa", "london-uk", "frauenfeld-switzerland", "turku-finland"}},
		},
		Dates: []APIDates{
			{ID: 1, Dates: []string{"14-12-1973"}},
			{ID: 2, Dates: []string{"26-03-2001"}},
			{ID: 4, Dates: []string{"04-07-1995"}},
			{ID: 30, Dates: []string{"05-07-2019"}},
		},
		Relations: []APIRelation{
			{ID: 1, DatesLocations: map[string][]string{"london-uk": {"14-12-1973"}}},
			{ID: 2, DatesLocations: map[string][]string{"paris-france": {"26-03-2001"}}},
			{ID: 4, DatesLocations: map[string][]string{"seattle-usa": {"04-07-1995"}}},
			{ID: 30, DatesLocations: map[string][]string{"turku-finland": {"05-07-2019"}}},
		},
	}
}
