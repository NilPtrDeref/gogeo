package simplification

import (
	"math"
	"sort"
)

// Douglas-Peucker Implementation

// Simplifies a set of points in place using the Douglas-Peucker algorithm.
// This implementation assigns a "rank" (threshold) to every point, allowing
// selection by percentage.
func DouglasPeuckerSimplify(
	coordinates [][][]float64,
	percentage float64,
) ([][][]float64, error) {
	result := make([][][]float64, 0, len(coordinates))

	for _, part := range coordinates {
		simplified, err := douglasPeuckerSimplifyRing(part, percentage)
		if err != nil {
			return nil, err
		}
		result = append(result, simplified)
	}
	return result, nil
}

func douglasPeuckerSimplifyRing(points [][]float64, percentage float64) ([][]float64, error) {
	minPoints := 4
	if len(points) <= minPoints {
		res := make([][]float64, len(points))
		copy(res, points)
		return res, nil
	}

	// Calculate Thresholds (Epsilon ranks) for all points
	thresholds := make([]float64, len(points))

	// Initialize endpoints to Infinity so they are always kept
	thresholds[0] = math.Inf(1)
	thresholds[len(points)-1] = math.Inf(1)

	if len(points) > 2 {
		// Recursive calculation
		_ = processSegment(points, thresholds, 0, len(points)-1, 1, math.MaxFloat64)
	}

	// Determine the Cutoff Value to select by percentage, we sort the thresholds to find the N-th largest value.
	sortedThresholds := make([]float64, len(points))
	copy(sortedThresholds, thresholds)
	// Sort descending
	sort.Slice(sortedThresholds, func(i, j int) bool {
		return sortedThresholds[i] > sortedThresholds[j]
	})

	targetLen := max(minPoints, int(float64(len(points))*percentage))

	// The value at the target index is our "cutoff" (epsilon).
	// Any point with a threshold >= cutoff is kept.
	var cutoff float64
	if targetLen < len(sortedThresholds) {
		cutoff = sortedThresholds[targetLen-1]
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
func processSegment(points [][]float64, dest []float64, startIdx, endIdx int, depth int, distSqPrev float64) float64 {
	ax := points[startIdx][0]
	ay := points[startIdx][1]
	cx := points[endIdx][0]
	cy := points[endIdx][1]

	var maxDistSq float64 = 0
	var maxIdx int = 0

	// Find point with maximum distance from segment AC
	for i := startIdx + 1; i < endIdx; i++ {
		distSq := getSqSegDist(points[i][0], points[i][1], ax, ay, cx, cy)
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
		distSqLeft = processSegment(points, dest, startIdx, maxIdx, depth+1, maxDistSq)
	}
	// Recurse Right
	if endIdx-maxIdx > 1 {
		distSqRight = processSegment(points, dest, maxIdx, endIdx, depth+1, maxDistSq)
	}

	// Constraint 2: Island Polygon (Ring) Preservation
	// If we are at the top level (depth 1) and it's a closed loop (start == end),
	// we force the max point to keep the max threshold of its children.
	// This prevents a small triangle from collapsing into a single point too early.
	if depth == 1 && ax == cx && ay == cy {
		maxDistSq = maxFloat64(distSqLeft, distSqRight)
	}

	dest[maxIdx] = math.Sqrt(maxDistSq)
	return maxDistSq
}

// Squared distance from point (px, py) to line segment (ax, ay)-(bx, by)
func getSqSegDist(px, py, ax, ay, bx, by float64) float64 {
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

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
