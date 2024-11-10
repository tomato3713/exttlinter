package a

import (
	"testing"
)

func Test_a(t *testing.T) {
	assert1 := func() {
		t.Error("no match") // want "should not use external testing object."
	}

	assert2 := func(t *testing.T) {
		t.Error("no match")
	}

	t.Run("sub test", func(t *testing.T) {
		assert1()
		assert2(t)
	})
}
