package common

import "math"

type AlbersParams struct {
	Phi1 float64
	Phi2 float64
	Phi0 float64
	Lam0 float64
}

type AlbersConstants struct {
	N     float64
	C     float64
	Rho0  float64
	Lam0r float64
	Rn    float64
}

func DegreesToRadian(d float64) float64 {
	return (d * math.Pi) / 180
}

func AlbersConstant(opts AlbersParams) AlbersConstants {
	phi1r := DegreesToRadian(opts.Phi1)
	phi2r := DegreesToRadian(opts.Phi2)
	phi0r := DegreesToRadian(opts.Phi0)
	Lam0r := DegreesToRadian(opts.Lam0)

	N := 0.5 * (math.Sin(phi1r) + math.Sin(phi2r))
	C := math.Pow(math.Cos(phi1r), 2) + 2*N*math.Sin(phi1r)
	Rn := 6378 / N
	Rho0 := Rn * math.Sqrt(C-2*N*math.Sin(phi0r))

	return AlbersConstants{N, C, Rho0, Lam0r, Rn}
}

func Albers(lat float64, lon float64, c AlbersConstants) Point {
	phir := DegreesToRadian(lat)
	lamr := DegreesToRadian(lon)

	rho := c.Rn * math.Sqrt(c.C-2*c.N*math.Sin(phir))
	theta := c.N * (lamr - c.Lam0r)
	x := rho * math.Sin(theta)
	y := c.Rho0 - rho*math.Cos(theta)
	return Point{x, y}
}
