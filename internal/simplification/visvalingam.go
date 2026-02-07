package simplification

import (
	"container/heap"
	"math"
)

// Visvalingam Implementation

// Options for weighting the effective area.
// Standard Visvalingam uses pure triangle area.
// Weighted Visvalingam adjusts area based on the angle (cosine) of the vertex.
type WeightOptions struct {
	// Default is usually 0.7 if enabled
	Weighting *float64
}

// Simplifies a set of points in place using the Visvalingam-Whyatt algorithm.
func VisvalingamSimplify(
	coordinates [][][]float64,
	percentage float64,
) ([][][]float64, error) {
	// TODO: Allow user to pass in weighting
	opts := WeightOptions{Weighting: new(float64)}
	*opts.Weighting = 0.7

	result := make([][][]float64, 0, len(coordinates))

	for _, part := range coordinates {
		simplified, err := visvalingamSimplifyRing(part, percentage, opts)
		if err != nil {
			return nil, err
		}
		result = append(result, simplified)
	}
	return result, nil
}

func calculateMetric(a, b, c []float64, opts WeightOptions) float64 {
	area := 0.5 * math.Abs(a[0]*(b[1]-c[1])+b[0]*(c[1]-a[1])+c[0]*(a[1]-b[1]))
	if opts.Weighting == nil {
		return area
	}
	k := *opts.Weighting

	// Weight function: -cos * k + 1
	// Sharp angles (cos ~ -1) -> Weight > 1 (Preserve)
	// Flat angles (cos ~ 1) -> Weight < 1 (Remove)
	cos := calculateCosine(a, b, c)
	return (-cos*k + 1.0) * area
}

func calculateCosine(a, b, c []float64) float64 {
	// Vector BA
	bax := a[0] - b[0]
	bay := a[1] - b[1]
	// Vector BC
	bcx := c[0] - b[0]
	bcy := c[1] - b[1]

	num := bax*bcx + bay*bcy
	den := math.Sqrt(bax*bax+bay*bay) * math.Sqrt(bcx*bcx+bcy*bcy)
	if den == 0 {
		return 0
	}
	return num / den
}

type pqItem struct {
	Index int
	Area  float64
}

// Priority Queue implementation
type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].Area < pq[j].Area }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

func (pq *priorityQueue) Push(x any) {
	item := x.(*pqItem)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func visvalingamSimplifyRing(points [][]float64, percentage float64, opts WeightOptions) ([][]float64, error) {
	minPoints := 4
	// Preserve basic shape (Triangle + Closing point)

	if len(points) <= minPoints {
		res := make([][]float64, len(points))
		copy(res, points)
		return res, nil
	}

	targetLen := max(minPoints, int(float64(len(points))*percentage))
	if targetLen >= len(points) {
		res := make([][]float64, len(points))
		copy(res, points)
		return res, nil
	}

	type node struct {
		Prev    int
		Next    int
		Area    float64
		Removed bool
	}

	nodes := make([]node, len(points))

	// Initialize Linked List
	for i := range nodes {
		if i == 0 {
			nodes[i].Prev = 0
		} else {
			nodes[i].Prev = i - 1
		}

		if i == len(points)-1 {
			nodes[i].Next = i
		} else {
			nodes[i].Next = i + 1
		}
		nodes[i].Removed = false
		nodes[i].Area = math.Inf(1)
	}

	// Calculate initial metrics for internal points
	for i := 1; i < len(points)-1; i++ {
		nodes[i].Area = calculateMetric(points[i-1], points[i], points[i+1], opts)
	}

	// Initialize Heap
	pq := make(priorityQueue, 0)
	heap.Init(&pq)

	for i := 1; i < len(points)-1; i++ {
		heap.Push(&pq, &pqItem{Index: i, Area: nodes[i].Area})
	}

	currLen := len(points)
	// Progressive Removal Loop
	for currLen > targetLen {
		if pq.Len() == 0 {
			break
		}
		item := heap.Pop(&pq).(*pqItem)
		idx := item.Index

		// Check if already removed
		if nodes[idx].Removed {
			continue
		}
		// Stale check (Lazy Deletion):
		// If the area in the PQ item doesn't match the current node area,
		// it means this node was updated with a new area later in the queue.
		// We discard this "stale" entry.
		if item.Area != nodes[idx].Area {
			continue
		}

		// Remove node
		nodes[idx].Removed = true
		currLen--

		prev := nodes[idx].Prev
		next := nodes[idx].Next

		// Relink neighbors
		nodes[prev].Next = next
		nodes[next].Prev = prev

		// Recalculate neighbors
		// We push NEW entries to the PQ.
		// The old entries for `prev` and `next`
		// remain in the PQ but will fail the "Stale check" when popped.
		if prev > 0 {
			newArea := calculateMetric(points[nodes[prev].Prev], points[prev], points[nodes[prev].Next], opts)
			nodes[prev].Area = newArea
			heap.Push(&pq, &pqItem{Index: prev, Area: newArea})
		}
		if next < len(points)-1 {
			newArea := calculateMetric(points[nodes[next].Prev], points[next], points[nodes[next].Next], opts)
			nodes[next].Area = newArea
			heap.Push(&pq, &pqItem{Index: next, Area: newArea})
		}
	}

	// Reconstruct Result
	result := make([][]float64, currLen)
	rindex := 0
	current := 0
	for {
		result[rindex] = points[current]
		rindex++
		if current == len(points)-1 {
			break
		}
		current = nodes[current].Next
	}

	return result, nil
}
