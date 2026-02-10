package common

import "github.com/nilptrderef/gogeo/internal/simplification"

//go:generate msgp -tests=false

type Counties []County

func (counties Counties) SimplifyInPlace(simplifier simplification.Simplifier, percentage float64) error {
	if simplifier == nil {
		return nil
	}

	for i := range counties {
		for j := range counties[i].Parts {
			var err error
			counties[i].Parts[j], err = simplifier.Simplify(counties[i].Parts[j], percentage)
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
	Parts       []Coordinates `msg:"coordinates"`
}

type Coordinates []float64
