package common

import (
	"strconv"

	"github.com/nilptrderef/gogeo/internal/simplification"
)

type Point struct {
	X float64
	Y float64
}

type Rectangle struct {
	Start Point
	End   Point
}

type Range struct {
	Min float64
	Max float64
}

type GeoJson struct {
	Type     string           `json:"type"`
	Features []GeoJsonFeature `json:"features"`
}

// TODO: Consider adding a minimum bounding rectangle, could be useful on the frontend.
func (geojson GeoJson) ToCounties() Counties {
	var c Counties
	for _, feature := range geojson.Features {
		var county County
		county.Id = feature.Properties["GEOID"]
		county.Name = feature.Properties["NAMELSAD"]
		county.State = StateAbbrFips[feature.Properties["STATEFP"]]

		lat, _ := strconv.ParseFloat(feature.Properties["INTPTLAT"], 32)
		county.InternalLat = float32(lat)
		lon, _ := strconv.ParseFloat(feature.Properties["INTPTLON"], 32)
		county.InternalLon = float32(lon)

		for _, part := range feature.Geometry.Coordinates {
			coordinates := make([]float64, len(part)*2)
			for i, point := range part {
				coordinates[i*2] = point[0]
				coordinates[i*2+1] = point[1]
			}
			county.Parts = append(county.Parts, coordinates)
		}
		c = append(c, county)
	}
	return c
}

func (geojson *GeoJson) SimplifyInPlace(simplifier simplification.Simplifier, percentage float64) error {
	if simplifier == nil {
		return nil
	}

	for i := range geojson.Features {
		for j := range geojson.Features[i].Geometry.Coordinates {
			var err error
			geojson.Features[i].Geometry.Coordinates[j], err = simplifier.SimplifyPoints(geojson.Features[i].Geometry.Coordinates[j], percentage)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type GeoJsonFeature struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Geometry   GeoJsonPolygon    `json:"geometry"`
}

type GeoJsonPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

var StateAbbrFips = map[string]string{
	"02": "AK",
	"28": "MS",
	"01": "AL",
	"30": "MT",
	"05": "AR",
	"37": "NC",
	"60": "AS",
	"38": "ND",
	"04": "AZ",
	"31": "NE",
	"06": "CA",
	"33": "NH",
	"08": "CO",
	"34": "NJ",
	"09": "CT",
	"35": "NM",
	"11": "DC",
	"32": "NV",
	"10": "DE",
	"36": "NY",
	"12": "FL",
	"39": "OH",
	"13": "GA",
	"40": "OK",
	"66": "GU",
	"41": "OR",
	"15": "HI",
	"42": "PA",
	"19": "IA",
	"72": "PR",
	"16": "ID",
	"44": "RI",
	"17": "IL",
	"45": "SC",
	"18": "IN",
	"46": "SD",
	"20": "KS",
	"47": "TN",
	"21": "KY",
	"48": "TX",
	"22": "LA",
	"49": "UT",
	"25": "MA",
	"51": "VA",
	"24": "MD",
	"78": "VI",
	"23": "ME",
	"50": "VT",
	"26": "MI",
	"53": "WA",
	"27": "MN",
	"55": "WI",
	"29": "MO",
	"54": "WV",
	"56": "WY",
}

var StateFips = map[string]string{
	"02": "ALASKA",
	"28": "MISSISSIPPI",
	"01": "ALABAMA",
	"30": "MONTANA",
	"05": "ARKANSAS",
	"37": "NORTH CAROLINA",
	"60": "AMERICAN SAMOA",
	"38": "NORTH DAKOTA",
	"04": "ARIZONA",
	"31": "NEBRASKA",
	"06": "CALIFORNIA",
	"33": "NEW HAMPSHIRE",
	"08": "COLORADO",
	"34": "NEW JERSEY",
	"09": "CONNECTICUT",
	"35": "NEW MEXICO",
	"11": "DISTRICT OF COLUMBIA",
	"32": "NEVADA",
	"10": "DELAWARE",
	"36": "NEW YORK",
	"12": "FLORIDA",
	"39": "OHIO",
	"13": "GEORGIA",
	"40": "OKLAHOMA",
	"66": "GUAM",
	"41": "OREGON",
	"15": "HAWAII",
	"42": "PENNSYLVANIA",
	"19": "IOWA",
	"72": "PUERTO RICO",
	"16": "IDAHO",
	"44": "RHODE ISLAND",
	"17": "ILLINOIS",
	"45": "SOUTH CAROLINA",
	"18": "INDIANA",
	"46": "SOUTH DAKOTA",
	"20": "KANSAS",
	"47": "TENNESSEE",
	"21": "KENTUCKY",
	"48": "TEXAS",
	"22": "LOUISIANA",
	"49": "UTAH",
	"25": "MASSACHUSETTS",
	"51": "VIRGINIA",
	"24": "MARYLAND",
	"78": "VIRGIN ISLANDS",
	"23": "MAINE",
	"50": "VERMONT",
	"26": "MICHIGAN",
	"53": "WASHINGTON",
	"27": "MINNESOTA",
	"55": "WISCONSIN",
	"29": "MISSOURI",
	"54": "WEST VIRGINIA",
	"56": "WYOMING",
}
