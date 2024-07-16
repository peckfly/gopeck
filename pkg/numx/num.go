package numx

func CeilDiv[T int | int32 | int64](a, b T) T {
	return (a + b - 1) / b
}
