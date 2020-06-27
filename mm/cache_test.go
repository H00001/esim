package buffer

import (
	"testing"
)

func TestMulLevelRoundTrip(t *testing.T) {
	ch := cacheing{}
	ch.Init(3)
	ch.Save(1, 2)
	ch.Save(2, 2)
	ch.Save(787, 9)
	ch.Save(8, 27)
	ch.Save(888, 27)

	ch.Get(787).ifPresent(func(i interface{}) {
		println(i.(int))
	})
}
