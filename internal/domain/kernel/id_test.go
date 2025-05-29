package kernel_test

import (
	"testing"

	"github.com/alnah/fla/internal/domain/kernel"
)

// TestEntity is a dummy type for testing generic IDs
type TestEntity struct{}

func TestNewID(t *testing.T) {
	t.Run("creates ID with valid input", func(t *testing.T) {
		input := "test-123"
		got, err := kernel.NewID[TestEntity](input)

		assertNoError(t, err)

		if got.String() != input {
			t.Errorf("got %q, want %q", got, input)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		input := "  test-123  "
		want := "test-123"

		got, err := kernel.NewID[TestEntity](input)

		assertNoError(t, err)

		if got.String() != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns error for empty string", func(t *testing.T) {
		_, err := kernel.NewID[TestEntity]("")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("returns error for whitespace only", func(t *testing.T) {
		_, err := kernel.NewID[TestEntity]("   ")

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestID_String(t *testing.T) {
	want := "test-id"
	id := kernel.ID[TestEntity](want)

	got := id.String()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestID_Validate(t *testing.T) {
	t.Run("valid ID passes validation", func(t *testing.T) {
		id := kernel.ID[TestEntity]("valid-id")

		err := id.Validate()

		assertNoError(t, err)
	})

	t.Run("empty ID fails validation", func(t *testing.T) {
		id := kernel.ID[TestEntity]("")

		err := id.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("whitespace-only ID fails validation", func(t *testing.T) {
		id := kernel.ID[TestEntity]("   ")

		err := id.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestID_TypeSafety(t *testing.T) {
	// This test verifies that IDs of different types are not interchangeable
	// It should not compile if uncommented:
	// type Entity1 struct{}
	// type Entity2 struct{}
	// var id1 kernel.ID[Entity1] = "id1"
	// var id2 kernel.ID[Entity2] = id1 // This would cause a compile error

	t.Run("IDs maintain type safety", func(t *testing.T) {
		// This test passes if the code compiles
		type Entity1 struct{}
		type Entity2 struct{}

		id1 := kernel.ID[Entity1]("test")
		id2 := kernel.ID[Entity2]("test")

		// These are different types even with same value
		if id1.String() != id2.String() {
			t.Errorf("string values should be equal")
		}
	})
}
