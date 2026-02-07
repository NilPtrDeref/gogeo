package shapefile

import (
	"encoding/binary"
	"io"

	. "github.com/nilptrderef/gogeo/internal/common"
	"github.com/nilptrderef/gogeo/internal/dbase"
	"github.com/nilptrderef/gogeo/internal/simplification"
)

type Shapefile struct {
	Header  Header
	Records []Record
}

func Parse(r io.Reader) (*Shapefile, error) {
	shp := &Shapefile{}
	if err := shp.Header.Parse(r); err != nil {
		return nil, err
	}

	for {
		var rh RecordHeader
		err := binary.Read(r, binary.BigEndian, &rh)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var st ShapeType
		if err := binary.Read(r, binary.LittleEndian, &st); err != nil {
			return nil, err
		}

		switch st {
		case PolygonType:
			poly := &Polygon{}
			if err := poly.Parse(r); err != nil {
				return nil, err
			}
			shp.Records = append(shp.Records, Record{Polygon: poly})
		default:
			// Discard remaining bytes in record. Total bytes = rh.Len * 2.
			// We read 4 bytes (ShapeType) already.
			if _, err := io.CopyN(io.Discard, r, int64(rh.Len*2-4)); err != nil {
				return nil, err
			}
		}
	}
	return shp, nil
}

func (s *Shapefile) LoadAttributes(r io.Reader) error {
	db, err := dbase.Parse(r)
	if err != nil {
		return err
	}

	for i := 0; i < int(db.Header.RecordCount); i++ {
		if i >= len(s.Records) {
			break
		}
		record := &s.Records[i]
		if record.Attrs == nil {
			record.Attrs = make(map[string]string)
		}

		var delFlag uint8
		if err := binary.Read(r, binary.LittleEndian, &delFlag); err != nil {
			return err
		}

		for _, field := range db.Fields {
			val, err := field.Read(r)
			if err != nil {
				return err
			}
			record.Attrs[field.GetName()] = val
		}
	}
	return nil
}

func (s *Shapefile) ToGeoJson() GeoJson {
	gj := GeoJson{
		Type:     "FeatureCollection",
		Features: make([]GeoJsonFeature, len(s.Records)),
	}

	for i, r := range s.Records {
		gj.Features[i] = GeoJsonFeature{
			Type:       "Feature",
			Properties: r.Attrs,
			Geometry:   r.Polygon.ToGeoJsonPolygon(),
		}
	}

	return gj
}

func (s *Shapefile) Simplify(percentage float64, alg simplification.Simplification) error {
	for i := range s.Records {
		if s.Records[i].Polygon != nil {
			if err := s.Records[i].Polygon.Simplify(percentage, alg); err != nil {
				return err
			}
		}
	}
	return nil
}

type Header struct {
	File  FileInfo
	Shape ShapeInfo
}

type FileInfo struct {
	FileCode uint32
	Reserved [5]uint32
	// Count of 16-bit words, including header
	FileLength uint32
}

type ShapeInfo struct {
	Version uint32
	Type    ShapeType
	Mbr     Rectangle
	Zrange  Range
	Mrange  Range
}

type ShapeType uint32

const (
	Null        ShapeType = 0
	PointType   ShapeType = 1
	Polyline    ShapeType = 3
	PolygonType ShapeType = 5
	MultiPoint  ShapeType = 8
	PointZ      ShapeType = 11
	PolylineZ   ShapeType = 13
	PolygonZ    ShapeType = 15
	MultiPointZ ShapeType = 18
	PointM      ShapeType = 21
	PolylineM   ShapeType = 23
	PolygonM    ShapeType = 25
	MultiPointM ShapeType = 28
	MultiPatch  ShapeType = 31
)

func (h *Header) Parse(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &h.File); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Shape); err != nil {
		return err
	}
	return nil
}

type Record struct {
	Polygon *Polygon
	Attrs   map[string]string
}

type RecordHeader struct {
	Index uint32
	// Count of 16-bit words, including header
	Len uint32
}

type Polygon struct {
	Header PolygonHeader
	Parts  []uint32
	Points []Point
}

type PolygonHeader struct {
	Mbr        Rectangle
	PartCount  uint32
	PointCount uint32
}

// If this fails, the caller CANNOT assume that the points/parts are in a valid state.
// They must still be freed, but it's possible that there are missing/invalid points.
func (p *Polygon) Simplify(percentage float64, alg simplification.Simplification) error {
	switch alg {
	case simplification.Visvalingam:
		return simplification.VisvalingamSimplify(&p.Points, &p.Parts, percentage)
	case simplification.DouglasPeucker:
		return simplification.DouglasPeuckerSimplify(&p.Points, &p.Parts, percentage)
	}
	p.Header.PointCount = uint32(len(p.Points))
	return nil
}

func (p *Polygon) Parse(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &p.Header); err != nil {
		return err
	}

	p.Parts = make([]uint32, p.Header.PartCount)
	if err := binary.Read(r, binary.LittleEndian, &p.Parts); err != nil {
		return err
	}

	p.Points = make([]Point, p.Header.PointCount)
	if err := binary.Read(r, binary.LittleEndian, &p.Points); err != nil {
		return err
	}

	return nil
}

func (p *Polygon) ToGeoJsonPolygon() GeoJsonPolygon {
	out := GeoJsonPolygon{
		Type:        "Polygon",
		Coordinates: [][][]float64{},
	}
	var currentPart []([]float64)

	// Helper to check if index is a part start
	isPart := func(idx int) bool {
		for _, partIdx := range p.Parts {
			if int(partIdx) == idx {
				return true
			}
		}
		return false
	}

	for i, pt := range p.Points {
		if i != 0 && isPart(i) {
			out.Coordinates = append(out.Coordinates, currentPart)
			currentPart = []([]float64){}
		}
		// Rounding can be applied here if needed, but standard float64 used
		currentPart = append(currentPart, []float64{pt.X, pt.Y})
	}
	if len(currentPart) > 0 {
		out.Coordinates = append(out.Coordinates, currentPart)
	}

	return out
}
