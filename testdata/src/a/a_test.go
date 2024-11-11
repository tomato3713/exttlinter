package a

import (
	"testing"
)

func Test_a(t *testing.T) {
	assert1 := func() {
		t.Error("no match") // want "should not use external testing object."
	}

	assert2 := func(t *testing.T, a int) {
		t.Error("no match")
	}

	assert3 := func(a int, tt *testing.T) {
		tt.Error("no match")
		t.Error("no match") // want "should not use external testing object."
	}

	t.Run("sub test", func(t *testing.T) {
		assert1()
		assert2(t, 1)
		assert3(1, t)
	})
}
