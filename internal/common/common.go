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

type GeoJsonFeature struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties,omitempty"`
	Geometry   GeoJsonPolygon    `json:"geometry"`
}

type GeoJsonPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}
