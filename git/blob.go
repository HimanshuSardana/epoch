package git

import (
	"fmt"
)

type Blob struct {
	Data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{Data: data}
}

func (b *Blob) Type() ObjectType {
	return ObjectBlob
}

func (b *Blob) Serialize() []byte {
	return b.Data
}

func (b *Blob) Parse(data []byte) error {
	b.Data = data
	return nil
}

func (b *Blob) String() string {
	return fmt.Sprintf("Blob{%d bytes}", len(b.Data))
}
