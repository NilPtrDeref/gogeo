package simplification

import (
	"errors"
	"math"
	"sort"
)

// Douglas-Peucker Implementation
// Simplifies a set of points in place using the Douglas-Peucker algorithm.
// This implementation assigns a "rank" (threshold) to every point, allowing
// selection by percentage.
type DouglasPeuckerSimplifier struct {
	// The minimum number of points to which a polygon should be simplified. The output should always
	// be greater than this number as long as the input is greater than this number. Defaults to 4.
	MinimumPoints int
}

// Simplify simplifies a set of coordinates (flat [x, y, x, y, ...]) using the Douglas-Peucker algorithm.
func (d DouglasPeuckerSimplifier) Simplify(coordinates []float64, percentage float64) ([]float64, error) {
	if len(coordinates)%2 != 0 {
		return nil, errors.New("coordinates must be divisible by 2")
	}

	pointCount := len(coordinates) / 2
	minimum := 4
	if d.MinimumPoints > 0 {
		minimum = d.MinimumPoints
	}

	target := max(minimum, int(float64(pointCount)*percentage))
	if pointCount <= minimum || target >= pointCount {
		result := make([]float64, len(coordinates))
		copy(result, coordinates)
		return result, nil
	}

	thresholds := make([]float64, pointCount)
	if pointCount > 0 {
		thresholds[0] = math.MaxFloat64
		thresholds[pointCount-1] = math.MaxFloat64
	}

	var processSegment func(startIdx, endIdx int, depth int, distSqPrev float64) float64
	processSegment = func(startIdx, endIdx int, depth int, distSqPrev float64) float64 {
		ax := coordinates[startIdx*2]
		ay := coordinates[startIdx*2+1]
		cx := coordinates[endIdx*2]
		cy := coordinates[endIdx*2+1]

		var maxDistSq float64 = 0
		var maxIdx int = 0

		// Find point with maximum distance from segment AC
		for i := startIdx + 1; i < endIdx; i++ {
			distSq := d.GetSqSegDist(coordinates[i*2], coordinates[i*2+1], ax, ay, cx, cy)
			if distSq >= maxDistSq {
				maxDistSq = distSq
				maxIdx = i
			}
		}

		if distSqPrev < maxDistSq {
			maxDistSq = distSqPrev
		}

		var distSqLeft float64 = 0
		var distSqRight float64 = 0
		if maxIdx-startIdx > 1 {
			distSqLeft = processSegment(startIdx, maxIdx, depth+1, maxDistSq)
		}
		if endIdx-maxIdx > 1 {
			distSqRight = processSegment(maxIdx, endIdx, depth+1, maxDistSq)
		}

		if depth == 1 && ax == cx && ay == cy {
			maxDistSq = max(distSqLeft, distSqRight)
		}

		thresholds[maxIdx] = math.Sqrt(maxDistSq)
		return maxDistSq
	}

	if pointCount > 2 {
		processSegment(0, pointCount-1, 1, math.MaxFloat64)
	}

	sortedThresholds := make([]float64, pointCount)
	copy(sortedThresholds, thresholds)
	sort.Slice(sortedThresholds, func(i, j int) bool {
		return sortedThresholds[i] > sortedThresholds[j]
	})

	var cutoff float64
	if target < len(sortedThresholds) {
		cutoff = sortedThresholds[target-1]
	} else {
		cutoff = 0
	}

	result := make([]float64, 0, target*2)
	for i := range pointCount {
		if thresholds[i] >= cutoff {
			result = append(result, coordinates[i*2], coordinates[i*2+1])
		}
	}

	return result, nil
}

// SimplifyPoints simplifies a set of points ([][x, y]) using the Douglas-Peucker algorithm.
func (d DouglasPeuckerSimplifier) SimplifyPoints(points [][]float64, percentage float64) ([][]float64, error) {
	for _, point := range points {
		if len(point) != 2 {
			return nil, errors.New("points must be all be of length 2")
		}
	}

	minimum := 4
	if d.MinimumPoints > 0 {
		minimum = d.MinimumPoints
	}

	target := max(minimum, int(float64(len(points))*percentage))
	if len(points) <= minimum || target >= len(points) {
		result := make([][]float64, len(points))
		for i, point := range points {
			result[i] = []float64{point[0], point[1]}
		}
		return result, nil
	}

	// Calculate Thresholds (Epsilon ranks) for all points
	thresholds := make([]float64, len(points))

	// Ensure the first and last points are always kept by giving them infinite threshold
	if len(points) > 0 {
		thresholds[0] = math.MaxFloat64
		thresholds[len(points)-1] = math.MaxFloat64
	}

	if len(points) > 2 {
		// Recursive calculation
		_ = d.ProcessSegment(points, thresholds, 0, len(points)-1, 1, math.MaxFloat64)
	}

	// Determine the Cutoff Value to select by percentage, we sort the thresholds to find the N-th largest value.
	sortedThresholds := make([]float64, len(points))
	copy(sortedThresholds, thresholds)
	// Sort descending
	sort.Slice(sortedThresholds, func(i, j int) bool {
		return sortedThresholds[i] > sortedThresholds[j]
	})

	// The value at the target index is our "cutoff" (epsilon).
	// Any point with a threshold >= cutoff is kept.
	var cutoff float64
	if target < len(sortedThresholds) {
		cutoff = sortedThresholds[target-1]
	} else {
		cutoff = 0
	}

	// Filter Points
	resultList := make([][]float64, 0)
	for i, p := range points {
		if thresholds[i] >= cutoff {
			resultList = append(resultList, p)
		}
	}

	return resultList, nil
}

// Recursive function to calculate maximum distance and assign thresholds
func (d DouglasPeuckerSimplifier) ProcessSegment(points [][]float64, dest []float64, startIdx, endIdx int, depth int, distSqPrev float64) float64 {
	ax := points[startIdx][0]
	ay := points[startIdx][1]
	cx := points[endIdx][0]
	cy := points[endIdx][1]

	var maxDistSq float64 = 0
	var maxIdx int = 0

	// Find point with maximum distance from segment AC
	for i := startIdx + 1; i < endIdx; i++ {
		distSq := d.GetSqSegDist(points[i][0], points[i][1], ax, ay, cx, cy)
		if distSq >= maxDistSq {
			maxDistSq = distSq
			maxIdx = i
		}
	}

	// Constraint 1: Parent Threshold Cap
	// Ensure child nodes never have a higher threshold than their parent segment.
	// This ensures that as you slide the epsilon slider, points disappear hierarchically.
	if distSqPrev < maxDistSq {
		maxDistSq = distSqPrev
	}

	var distSqLeft float64 = 0
	var distSqRight float64 = 0
	// Recurse Left
	if maxIdx-startIdx > 1 {
		distSqLeft = d.ProcessSegment(points, dest, startIdx, maxIdx, depth+1, maxDistSq)
	}
	// Recurse Right
	if endIdx-maxIdx > 1 {
		distSqRight = d.ProcessSegment(points, dest, maxIdx, endIdx, depth+1, maxDistSq)
	}

	// Constraint 2: Island Polygon (Ring) Preservation
	// If we are at the top level (depth 1) and it's a closed loop (start == end),
	// we force the max point to keep the max threshold of its children.
	// This prevents a small triangle from collapsing into a single point too early.
	if depth == 1 && ax == cx && ay == cy {
		maxDistSq = max(distSqLeft, distSqRight)
	}

	dest[maxIdx] = math.Sqrt(maxDistSq)
	return maxDistSq
}

// Squared distance from point (px, py) to line segment (ax, ay)-(bx, by)
func (d DouglasPeuckerSimplifier) GetSqSegDist(px, py, ax, ay, bx, by float64) float64 {
	dx := ax - bx
	dy := ay - by

	if dx != 0 || dy != 0 {
		// Project point onto line, clamped to segment [0, 1]
		t := ((px-ax)*-dx + (py-ay)*-dy) / (dx*dx + dy*dy)
		if t > 1 {
			dx = px - bx
			dy = py - by
		} else if t > 0 {
			dx = px - (ax - dx*t)
			dy = py - (ay - dy*t)
		} else {
			dx = px - ax
			dy = py - ay
		}
	} else {
		// Segment is a single point
		dx = px - ax
		dy = py - ay
	}

	return dx*dx + dy*dy
}
