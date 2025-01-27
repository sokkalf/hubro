package utils

import (
	"time"

	"github.com/gosimple/slug"
)

func Slugify(s string) string {
	return slug.Make(s)
}

func ParseDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}
	}
	return t
}

func Map[A any, B any](f func(A) B, arr []A) []B {
	res := make([]B, len(arr))
	for i, a := range arr {
		res[i] = f(a)
	}
	return res
}

func Filter[A any](f func(A) bool, arr []A) []A {
	res := make([]A, 0)
	for _, a := range arr {
		if f(a) {
			res = append(res, a)
		}
	}
	return res
}

func Reject[A any](f func(A) bool, arr []A) []A {
	return Filter(func(a A) bool {
		return !f(a)
	}, arr)
}
