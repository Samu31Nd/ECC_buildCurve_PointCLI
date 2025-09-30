package mathcurve

import "errors"

type Point struct {
	X, Y int
}

// inverso modular usando extendido de Euclides
func modInverse(a, p int) (int, error) {
	a = ((a % p) + p) % p
	if a == 0 {
		return 0, errors.New("no existe inverso modular de 0")
	}
	t, newT := 0, 1
	r, newR := p, a
	for newR != 0 {
		quot := r / newR
		t, newT = newT, t-quot*newT
		r, newR = newR, r-quot*newR
	}
	if r > 1 {
		return 0, errors.New("no existe inverso modular")
	}
	if t < 0 {
		t += p
	}
	return t, nil
}

func AddPoints(x1, y1, x2, y2, a, p int) (Point, error) {
	// Punto identidad (infinito)
	if x1 == -1 && y1 == -1 {
		return Point{x1, y1}, nil
	}
	if x2 == -1 && y2 == -1 {
		return Point{x1, y1}, nil
	}

	var lambda int
	var err error

	if x1 == x2 && y1 == y2 {
		// Caso P = Q (doblamiento)
		num := (3*x1*x1 + a) % p
		den := (2 * y1) % p
		invDen, invErr := modInverse(den, p)
		if invErr != nil {
			return Point{-1, -1}, errors.New("indeterminación: no se puede calcular lambda")
		}
		lambda = (num * invDen) % p
	} else {
		// Caso P != Q
		num := (y2 - y1) % p
		den := (x2 - x1) % p
		invDen, invErr := modInverse(den, p)
		if invErr != nil {
			return Point{-1, -1}, errors.New("indeterminación: no se puede calcular lambda")
		}
		lambda = (num * invDen) % p
	}

	// Coordenadas del nuevo punto
	x3 := (lambda*lambda - x1 - x2) % p
	y3 := (lambda*(x1-x3) - y1) % p

	// Normalizar al rango [0,p)
	if x3 < 0 {
		x3 += p
	}
	if y3 < 0 {
		y3 += p
	}

	return Point{x3, y3}, err
}
