package simplification

type Simplifier interface {
	// Simplify takes a slice of float64 which must have a length divisible by 2. It is interpreted as a slice
	// of X,Y coordinates for polygon a single polygon. A simplifier's second parameter is the approximate percentage
	// of points that should be remaining when the simplifier is done.
	Simplify(coordinates []float64, percentage float64) ([]float64, error)

	SimplifyPoints(points [][]float64, percentage float64) ([][]float64, error)
}
