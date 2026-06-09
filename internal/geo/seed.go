package geo

import _ "embed"

//go:embed seed.json
var seedJSON []byte

// SeedData returns the embedded geocode cache, mapping location slugs to
// coordinates. It lets the deployed app resolve known concert locations
// instantly without calling the geocoding service on every request.
func SeedData() []byte {
	return seedJSON
}
