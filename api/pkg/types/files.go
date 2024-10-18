package types

import "io"

type File struct {
	Name   string
	Data   io.Reader
	Size   int64
	Format string
}
