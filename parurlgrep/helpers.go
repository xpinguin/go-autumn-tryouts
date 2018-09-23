package main

func Min(x0 int, xs ...int) int {
	r := x0
	for _, x := range xs {
		if x < r {
			r = x
		}
	}
	return r
}