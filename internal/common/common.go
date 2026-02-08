package common

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

func (geojson GeoJson) ToCounties() Counties {
	var c Counties
	for _, feature := range geojson.Features {
		var county County
		county.Id = feature.Properties["GEOID"]
		county.Name = feature.Properties["NAMELSAD"]
		county.State = StateFips[feature.Properties["STATEFP"]]
		for _, part := range feature.Geometry.Coordinates {
			coordinates := make([]float32, len(part)*2)
			for i, point := range part {
				coordinates[i*2] = float32(point[0])
				coordinates[i*2+1] = float32(point[1])
			}
			county.Parts = append(county.Parts, coordinates)
		}
		c = append(c, county)
	}
	return c
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

var StateFips = map[string]string{
	"02": "AK, ALASKA",
	"28": "MS, MISSISSIPPI",
	"01": "AL, ALABAMA",
	"30": "MT, MONTANA",
	"05": "AR, ARKANSAS",
	"37": "NC, NORTH CAROLINA",
	"60": "AS, AMERICAN SAMOA",
	"38": "ND, NORTH DAKOTA",
	"04": "AZ, ARIZONA",
	"31": "NE, NEBRASKA",
	"06": "CA, CALIFORNIA",
	"33": "NH, NEW HAMPSHIRE",
	"08": "CO, COLORADO",
	"34": "NJ, NEW JERSEY",
	"09": "CT, CONNECTICUT",
	"35": "NM, NEW MEXICO",
	"11": "DC, DISTRICT OF COLUMBIA",
	"32": "NV, NEVADA",
	"10": "DE, DELAWARE",
	"36": "NY, NEW YORK",
	"12": "FL, FLORIDA",
	"39": "OH, OHIO",
	"13": "GA, GEORGIA",
	"40": "OK, OKLAHOMA",
	"66": "GU, GUAM",
	"41": "OR, OREGON",
	"15": "HI, HAWAII",
	"42": "PA, PENNSYLVANIA",
	"19": "IA, IOWA",
	"72": "PR, PUERTO RICO",
	"16": "ID, IDAHO",
	"44": "RI, RHODE ISLAND",
	"17": "IL, ILLINOIS",
	"45": "SC, SOUTH CAROLINA",
	"18": "IN, INDIANA",
	"46": "SD, SOUTH DAKOTA",
	"20": "KS, KANSAS",
	"47": "TN, TENNESSEE",
	"21": "KY, KENTUCKY",
	"48": "TX, TEXAS",
	"22": "LA, LOUISIANA",
	"49": "UT, UTAH",
	"25": "MA, MASSACHUSETTS",
	"51": "VA, VIRGINIA",
	"24": "MD, MARYLAND",
	"78": "VI, VIRGIN ISLANDS",
	"23": "ME, MAINE",
	"50": "VT, VERMONT",
	"26": "MI, MICHIGAN",
	"53": "WA, WASHINGTON",
	"27": "MN, MINNESOTA",
	"55": "WI, WISCONSIN",
	"29": "MO, MISSOURI",
	"54": "WV, WEST VIRGINIA",
	"56": "WY, WYOMING",
}
