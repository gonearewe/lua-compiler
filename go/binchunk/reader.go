package binchunk

import (
	"golang.org/x/text/date"
)

type reader struct{
	data []byte
}

type (r *reader)readByte() byte{
	b:=r.data[0]
	b.data=b.data[1:]
	return b
}