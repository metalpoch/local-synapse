package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type osmResponse struct {
	Type     string `json:"type"`
	Licence  string `json:"licence"`
	Features []struct {
		Type       string `json:"type"`
		Properties struct {
			PlaceID     int     `json:"place_id"`
			OsmType     string  `json:"osm_type"`
			OsmID       int     `json:"osm_id"`
			PlaceRank   int     `json:"place_rank"`
			Category    string  `json:"category"`
			Type        string  `json:"type"`
			Importance  float64 `json:"importance"`
			Addresstype string  `json:"addresstype"`
			Name        string  `json:"name"`
			DisplayName string  `json:"display_name"`
			Address     struct {
				Municipality string `json:"municipality"`
				County       string `json:"county"`
				State        string `json:"state"`
				ISO31662Lvl4 string `json:"ISO3166-2-lvl4"`
				Country      string `json:"country"`
				CountryCode  string `json:"country_code"`
			} `json:"address"`
		} `json:"properties"`
		Bbox     []float64 `json:"bbox"`
		Geometry struct {
			Type        string        `json:"type"`
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

type osmLocation struct {
	Country             string        `json:"country"`
	CountryCode         string        `json:"country_code"`
	State               string        `json:"state"`
	Municipality        string        `json:"municipality"`
	County              string        `json:"county"`
	GeometryCoordinates [][][]float64 `json:"coordinates"`
}

func OSMBoundaryZone() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool("osm-boundary-fetcher",
		mcp.WithDescription("Fetches boundary data (bounding box and polygon coordinates) for a geographic zone using OpenStreetMap's Nominatim API"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Geographic location to fetch boundaries for (e.g., 'Madrid, Spain', 'Paris')"),
		),
	)

	return tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := osmBoundary(query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get boundary zone: %v", err)), nil
		}

		return mcp.NewToolResultText(data), nil
	}
}

func osmBoundary(query string) (string, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   "nominatim.openstreetmap.org",
		Path:   url.PathEscape("search"),
	}
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "geojson")
	params.Add("addressdetails", "1")
	params.Add("polygon_geojson", "1")

	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading the response: %v", err)
	}

	var res osmResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	var locations []osmLocation
	for _, feat := range res.Features {
		if feat.Geometry.Type != "Polygon" {
			continue
		}

		addr := feat.Properties.Address
		locations = append(locations, osmLocation{
			Country:             addr.Country,
			CountryCode:         addr.CountryCode,
			State:               addr.State,
			Municipality:        addr.Municipality,
			County:              addr.Country,
			GeometryCoordinates: feat.Geometry.Coordinates,
		})
	}

	jsonLocations, err := json.Marshal(locations)
	if err != nil {
		return "", fmt.Errorf("error marshaling: %v", err)
	}

	return string(jsonLocations), nil
}
