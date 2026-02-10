package common

import "github.com/nilptrderef/gogeo/internal/simplification"

//go:generate msgp -tests=false

type Point struct {
	X float64 `msg:"x"`
	Y float64 `msg:"y"`
}

type Rectangle struct {
	Start Point `msg:"start"`
	End   Point `msg:"end"`
}

type Range struct {
	Min float64 `msg:"min"`
	Max float64 `msg:"max"`
}

type Map struct {
	Mbr      Rectangle `msg:"minimum_bounding_rectangle"`
	Counties Counties  `msg:"counties"`
}

type Counties []County

func (m Map) SimplifyInPlace(simplifier simplification.Simplifier, percentage float64) error {
	if simplifier == nil {
		return nil
	}

	for i := range m.Counties {
		for j := range m.Counties[i].Parts {
			var err error
			m.Counties[i].Parts[j], err = simplifier.Simplify(m.Counties[i].Parts[j], percentage)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type County struct {
	Id          string        `msg:"id"`
	Name        string        `msg:"name"`
	State       string        `msg:"state"`
	InternalLat float32       `msg:"intlat"`
	InternalLon float32       `msg:"intlon"`
	Mbr         Rectangle     `msg:"minimum_bounding_rectangle"`
	Parts       []Coordinates `msg:"coordinates"`
}

type Coordinates []float64
