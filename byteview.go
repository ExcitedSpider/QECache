// A structure represent the cached value
// use byte for its universality
package qecache

type ByteView struct {
	value []byte
}

func (v ByteView) Len() int {
	return len(v.value)
}

// The slice return a copy of value
func (v ByteView) ByteSlice() []byte {
	clone := make([]byte, v.Len())
	copy(clone, v.value)
	return clone
}

func (v ByteView) String() string {
	return string(v.value)
}
