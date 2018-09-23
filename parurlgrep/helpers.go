package main

import (
	"io"
	//"log"
)

func Min(x0 int, xs ...int) int {
	r := x0
	for _, x := range xs {
		if x < r {
			r = x
		}
	}
	return r
}

type stringReader struct {
	string
	_pos int
}

func NewStringReader(s string) *stringReader {
	return &stringReader{s, 0}
}

func (s *stringReader) Read(out_s []byte) (int, error) {
	// TODO: rewrite the mess
	s_pos := &(s._pos)
	if *s_pos >= len(s.string) {
		return 0, io.EOF
	}
	s_pos_end := Min(*s_pos+len(out_s), len(s.string))
	s_bytes := []byte(s.string[*s_pos:s_pos_end])
	//log.Println(len(s_bytes), *s_pos, s_pos_end, len(out_s), string(s_bytes))

	copy(out_s, s_bytes)
	//log.Println(len(s_bytes), *s_pos, s_pos_end, len(out_s), out_s)

	*s_pos = s_pos_end
	if len(s_bytes) > 0 {
		return len(s_bytes), nil
	}
	return 0, io.EOF
}
