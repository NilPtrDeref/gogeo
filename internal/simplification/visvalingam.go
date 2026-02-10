package simplification

import (
	"container/heap"
	"errors"
	"math"
)

type VisvalingamSimplifier struct {
	// Weighting the effective area when simplifying.
	// Standard Visvalingam uses pure triangle area.
	// Weighted Visvalingam adjusts area based on the angle (cosine) of the vertex.
	// Value of 0 will be ignored. Default is usually 0.7 if enabled
	Weighting float64

	// The minimum number of points to which a polygon should be simplified. The output should always
	// be greater than this number as long as the input is greater than this number. Defaults to 4.
	MinimumPoints int
}

// Simplifies a set of points in place using the Visvalingam-Whyatt algorithm.
func (v VisvalingamSimplifier) Simplify(coordinates []float64, percentage float64) ([]float64, error) {
	if len(coordinates)%2 != 0 {
		return nil, errors.New("coordinates must be divisible by 2")
	}
	points := len(coordinates) / 2

	minimum := 4
	if v.MinimumPoints > 0 {
		minimum = v.MinimumPoints
	}

	target := max(minimum, int(float64(points)*percentage))
	if points <= minimum || target >= points {
		result := make([]float64, len(coordinates))
		copy(result, coordinates)
		return result, nil
	}

	nodes := make([]node, points)
	for i := range nodes {
		if i == 0 {
			nodes[i].Prev = points - 1
		} else {
			nodes[i].Prev = i - 1
		}

		if i == points-1 {
			nodes[i].Next = 0
		} else {
			nodes[i].Next = i + 1
		}
		nodes[i].Removed = false

		previous := nodes[i].Prev * 2
		current := i * 2
		next := nodes[i].Next * 2
		nodes[i].Area = v.CalculateMetric(coordinates[previous:previous+2], coordinates[current:current+2], coordinates[next:next+2])
	}

	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	for i := 0; i < points-1; i++ {
		heap.Push(&pq, &pqItem{Index: i, Area: nodes[i].Area})
	}

	// Progressive Removal Loop
	current := points
	for current > target {
		if pq.Len() == 0 {
			break
		}

		item := heap.Pop(&pq).(*pqItem)
		if nodes[item.Index].Removed {
			continue
		}

		// Stale check (Lazy Deletion):
		// If the area in the PQ item doesn't match the current node area,
		// it means this node was updated with a new area later in the queue.
		// We discard this "stale" entry.
		if item.Area != nodes[item.Index].Area {
			continue
		}

		// Remove node
		nodes[item.Index].Removed = true
		current--

		// Relink neighbors
		prev := nodes[item.Index].Prev
		next := nodes[item.Index].Next
		nodes[prev].Next = next
		nodes[next].Prev = prev

		// Recalculate neighbors
		// We push NEW entries to the PQ.
		// The old entries for `prev` and `next`
		// remain in the PQ but will fail the "Stale check" when popped.

		// Previous
		prevprev := nodes[prev].Prev * 2
		prevcurr := prev * 2
		prevnext := nodes[prev].Next * 2
		area := v.CalculateMetric(coordinates[prevprev:prevprev+2], coordinates[prevcurr:prevcurr+2], coordinates[prevnext:prevnext+2])
		nodes[prev].Area = area
		heap.Push(&pq, &pqItem{Index: prev, Area: area})
		// Next
		nextprev := nodes[next].Prev * 2
		nextcurr := next * 2
		nextnext := nodes[next].Next * 2
		area = v.CalculateMetric(coordinates[nextprev:nextprev+2], coordinates[nextcurr:nextcurr+2], coordinates[nextnext:nextnext+2])
		nodes[next].Area = area
		heap.Push(&pq, &pqItem{Index: next, Area: area})
	}

	result := make([]float64, target*2)
	index := 0
	for i, node := range nodes {
		if !node.Removed {
			result[index] = coordinates[i*2]
			result[index+1] = coordinates[i*2+1]
			index += 2
		}
	}

	return result, nil
}

// Simplifies a set of points in place using the Visvalingam-Whyatt algorithm.
func (v VisvalingamSimplifier) SimplifyPoints(points [][]float64, percentage float64) ([][]float64, error) {
	for _, point := range points {
		if len(point) != 2 {
			return nil, errors.New("points must be all be of length 2")
		}
	}

	minimum := 4
	if v.MinimumPoints > 0 {
		minimum = v.MinimumPoints
	}

	target := max(minimum, int(float64(len(points))*percentage))
	if len(points) <= minimum || target >= len(points) {
		result := make([][]float64, len(points))
		for i, point := range points {
			result[i] = append(result[i], point[0], point[1])
		}
		return result, nil
	}

	nodes := make([]node, len(points))
	for i := range nodes {
		if i == 0 {
			nodes[i].Prev = len(points) - 1
		} else {
			nodes[i].Prev = i - 1
		}

		if i == len(points)-1 {
			nodes[i].Next = 0
		} else {
			nodes[i].Next = i + 1
		}
		nodes[i].Removed = false
		nodes[i].Area = v.CalculateMetric(points[nodes[i].Prev], points[i], points[nodes[i].Next])
	}

	pq := make(priorityQueue, 0)
	heap.Init(&pq)
	for i := 0; i < len(points)-1; i++ {
		heap.Push(&pq, &pqItem{Index: i, Area: nodes[i].Area})
	}

	// Progressive Removal Loop
	current := len(points)
	for current > target {
		if pq.Len() == 0 {
			break
		}

		item := heap.Pop(&pq).(*pqItem)
		if nodes[item.Index].Removed {
			continue
		}

		// Stale check (Lazy Deletion):
		// If the area in the PQ item doesn't match the current node area,
		// it means this node was updated with a new area later in the queue.
		// We discard this "stale" entry.
		if item.Area != nodes[item.Index].Area {
			continue
		}

		// Remove node
		nodes[item.Index].Removed = true
		current--

		// Relink neighbors
		prev := nodes[item.Index].Prev
		next := nodes[item.Index].Next
		nodes[prev].Next = next
		nodes[next].Prev = prev

		// Recalculate neighbors
		// We push NEW entries to the PQ.
		// The old entries for `prev` and `next`
		// remain in the PQ but will fail the "Stale check" when popped.

		// Previous
		area := v.CalculateMetric(points[nodes[prev].Prev], points[prev], points[nodes[prev].Next])
		nodes[prev].Area = area
		heap.Push(&pq, &pqItem{Index: prev, Area: area})
		// Next
		area = v.CalculateMetric(points[nodes[next].Prev], points[next], points[nodes[next].Next])
		nodes[next].Area = area
		heap.Push(&pq, &pqItem{Index: next, Area: area})
	}

	result := make([][]float64, target)
	index := 0
	for i, node := range nodes {
		if !node.Removed {
			result[index] = append(result[index], points[i][0], points[i][1])
			index++
		}
	}

	return result, nil
}

func (v VisvalingamSimplifier) CalculateMetric(a, b, c []float64) float64 {
	area := 0.5 * math.Abs(a[0]*(b[1]-c[1])+b[0]*(c[1]-a[1])+c[0]*(a[1]-b[1]))
	if v.Weighting == 0 {
		return area
	}

	// Weight function: -cos * k + 1
	// Sharp angles (cos ~ -1) -> Weight > 1 (Preserve)
	// Flat angles (cos ~ 1) -> Weight < 1 (Remove)
	cos := v.CalculateCosine(a, b, c)
	return (-cos*v.Weighting + 1.0) * area
}

func (v VisvalingamSimplifier) CalculateCosine(a, b, c []float64) float64 {
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

type node struct {
	Prev    int
	Next    int
	Area    float64
	Removed bool
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
