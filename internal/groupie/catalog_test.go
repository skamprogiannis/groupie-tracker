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

	assertSingleSearchResult(t, catalog.Search("freddie"), "Queen")
	assertSingleSearchResult(t, catalog.Search("26-03-2001"), "Gorillaz")
	assertSingleSearchResult(t, catalog.Search("turku-finland"), "Travis Scott")
	assertSingleSearchResult(t, catalog.Search("Rami Jaffee"), "Foo Fighters")
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
