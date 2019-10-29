package number

import "math"

// 在lua里，除法都是是向下取整的，而在Go 和C 中，都是向0取整的
func IFloorDiv(a, b int64) int64 {
	if a > 0 && b > 0 || a < 0 && b < 0 || a%b == 0 {
		return a / b
	} else {
		return a/b - 1
	}
}

func FFloorDiv(a, b float64) float64 {
	return math.Floor(a / b)
}

func IMod(a, b int64) int64 {
	return a - IFloorDiv(a, b)*b
}

func FMod(a, b float64) float64 {
	return a - math.Floor(a/b)*b
}

func ShiftLeft(a, n int64) int64 {
	if n >= 0 {
		// in Go, right operand of << must be an unsigned int
		return a << uint64(n)
	} else {
		return ShiftRight(a, -n)
	}
}

func ShiftRight(a, n int64) int64 {
	if n >= 0 {
		// uint64(a) makes sure it moves unsigned int
		return int64(uint64(a) >> uint64(n))
	} else {
		return ShiftLeft(a, -n)
	}
}
