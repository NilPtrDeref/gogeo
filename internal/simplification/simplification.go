package simplification

import . "github.com/nilptrderef/gogeo/internal/common"

type Simplification int

const (
	Visvalingam Simplification = iota
	DouglasPeucker
)

func Simplify(geojson GeoJson, percentage float64, alg Simplification) error {
	for i := range geojson.Features {
		if geojson.Features[i].Geometry.Coordinates != nil {
			var err error
			switch alg {
			case Visvalingam:
				if geojson.Features[i].Geometry.Coordinates, err = VisvalingamSimplify(geojson.Features[i].Geometry.Coordinates, percentage); err != nil {
					return err
				}
			case DouglasPeucker:
				if geojson.Features[i].Geometry.Coordinates, err = DouglasPeuckerSimplify(geojson.Features[i].Geometry.Coordinates, percentage); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
