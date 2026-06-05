package groupie

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const DefaultClientTimeout = 10 * time.Second

type APIData struct {
	Links     APILinks
	Artists   []APIArtist
	Locations []APILocations
	Dates     []APIDates
	Relations []APIRelation
}

type Client struct {
	endpoints  Endpoints
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	if timeout <= 0 {
		timeout = DefaultClientTimeout
	}
	return &Client{
		endpoints: NewEndpoints(baseURL),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func NewClientWithHTTP(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultClientTimeout}
	}
	return &Client{
		endpoints:  NewEndpoints(baseURL),
		httpClient: httpClient,
	}
}

func (c *Client) Endpoints() Endpoints {
	return c.endpoints
}

func (c *Client) FetchAll(ctx context.Context) (APIData, error) {
	var data APIData
	if err := c.getJSON(ctx, c.endpoints.Root, &data.Links); err != nil {
		return APIData{}, err
	}

	if err := c.getJSON(ctx, c.endpoints.Artists, &data.Artists); err != nil {
		return APIData{}, err
	}

	var locations APILocationsIndex
	if err := c.getJSON(ctx, c.endpoints.Locations, &locations); err != nil {
		return APIData{}, err
	}
	data.Locations = locations.Index

	var dates APIDatesIndex
	if err := c.getJSON(ctx, c.endpoints.Dates, &dates); err != nil {
		return APIData{}, err
	}
	data.Dates = dates.Index

	var relations APIRelationIndex
	if err := c.getJSON(ctx, c.endpoints.Relation, &relations); err != nil {
		return APIData{}, err
	}
	data.Relations = relations.Index

	return data, nil
}

func (c *Client) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request for %s: %w", url, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}
