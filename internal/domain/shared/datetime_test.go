package shared_test

import (
	"testing"
	"time"

	"github.com/alnah/fla/internal/domain/kernel"
	"github.com/alnah/fla/internal/domain/shared"
)

func TestNewDatetime(t *testing.T) {
	t.Run("creates datetime with past time", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)

		got, err := shared.NewDatetime(pastTime)

		assertNoError(t, err)
		if !got.Time().Equal(pastTime.UTC()) {
			t.Errorf("got %v, want %v", got.Time(), pastTime.UTC())
		}
	})

	t.Run("creates datetime with current time", func(t *testing.T) {
		now := time.Now()

		got, err := shared.NewDatetime(now)

		assertNoError(t, err)
		// Allow small time difference due to execution time
		diff := got.Time().Sub(now.UTC())
		if diff > time.Second || diff < -time.Second {
			t.Errorf("time difference too large: %v", diff)
		}
	})

	t.Run("converts to UTC", func(t *testing.T) {
		// Create time in different timezone
		loc, _ := time.LoadLocation("America/New_York")
		nyTime := time.Date(2024, 1, 1, 12, 0, 0, 0, loc)

		got, err := shared.NewDatetime(nyTime)

		assertNoError(t, err)
		if got.Time().Location() != time.UTC {
			t.Errorf("expected UTC location, got %v", got.Time().Location())
		}
		if !got.Time().Equal(nyTime.UTC()) {
			t.Errorf("time not properly converted to UTC")
		}
	})

	t.Run("rejects future time", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)

		_, err := shared.NewDatetime(futureTime)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("rejects zero time", func(t *testing.T) {
		var zeroTime time.Time

		_, err := shared.NewDatetime(zeroTime)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestNewDatetimeNow(t *testing.T) {
	t.Run("creates datetime with current time", func(t *testing.T) {
		before := time.Now()

		got, err := shared.NewDatetimeNow()

		after := time.Now()

		assertNoError(t, err)

		// Verify time is between before and after
		if got.Time().Before(before.UTC()) || got.Time().After(after.UTC()) {
			t.Errorf("datetime not within expected range")
		}
	})

	t.Run("returns UTC time", func(t *testing.T) {
		got, err := shared.NewDatetimeNow()

		assertNoError(t, err)
		if got.Time().Location() != time.UTC {
			t.Errorf("expected UTC location, got %v", got.Time().Location())
		}
	})
}

func TestNewDatetimeAllowFuture(t *testing.T) {
	t.Run("allows future time", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)

		got, err := shared.NewDatetimeAllowFuture(futureTime)

		assertNoError(t, err)
		if !got.Time().Equal(futureTime.UTC()) {
			t.Errorf("got %v, want %v", got.Time(), futureTime.UTC())
		}
	})

	t.Run("allows past time", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)

		got, err := shared.NewDatetimeAllowFuture(pastTime)

		assertNoError(t, err)
		if !got.Time().Equal(pastTime.UTC()) {
			t.Errorf("got %v, want %v", got.Time(), pastTime.UTC())
		}
	})

	t.Run("rejects zero time", func(t *testing.T) {
		var zeroTime time.Time

		_, err := shared.NewDatetimeAllowFuture(zeroTime)

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestDatetime_Time(t *testing.T) {
	want := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	dt, _ := shared.NewDatetime(want)

	got := dt.Time()

	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDatetime_Before(t *testing.T) {
	earlier := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	later := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	dt1, _ := shared.NewDatetime(earlier)
	dt2, _ := shared.NewDatetime(later)

	t.Run("earlier is before later", func(t *testing.T) {
		if !dt1.Before(dt2) {
			t.Error("expected dt1 to be before dt2")
		}
	})

	t.Run("later is not before earlier", func(t *testing.T) {
		if dt2.Before(dt1) {
			t.Error("expected dt2 not to be before dt1")
		}
	})

	t.Run("same time is not before itself", func(t *testing.T) {
		if dt1.Before(dt1) {
			t.Error("expected dt1 not to be before itself")
		}
	})
}

func TestDatetime_After(t *testing.T) {
	earlier := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	later := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	dt1, _ := shared.NewDatetime(earlier)
	dt2, _ := shared.NewDatetime(later)

	t.Run("later is after earlier", func(t *testing.T) {
		if !dt2.After(dt1) {
			t.Error("expected dt2 to be after dt1")
		}
	})

	t.Run("earlier is not after later", func(t *testing.T) {
		if dt1.After(dt2) {
			t.Error("expected dt1 not to be after dt2")
		}
	})

	t.Run("same time is not after itself", func(t *testing.T) {
		if dt1.After(dt1) {
			t.Error("expected dt1 not to be after itself")
		}
	})
}

func TestDatetime_Equal(t *testing.T) {
	time1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	time3 := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	dt1, _ := shared.NewDatetime(time1)
	dt2, _ := shared.NewDatetime(time2)
	dt3, _ := shared.NewDatetime(time3)

	t.Run("same times are equal", func(t *testing.T) {
		if !dt1.Equal(dt2) {
			t.Error("expected dt1 to equal dt2")
		}
	})

	t.Run("different times are not equal", func(t *testing.T) {
		if dt1.Equal(dt3) {
			t.Error("expected dt1 not to equal dt3")
		}
	})
}

func TestDatetime_String(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	dt, _ := shared.NewDatetime(testTime)

	got := dt.String()
	want := "2024-01-15T14:30:45Z"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDatetime_Validate(t *testing.T) {
	t.Run("valid past datetime passes", func(t *testing.T) {
		dt, _ := shared.NewDatetime(time.Now().Add(-time.Hour))

		err := dt.Validate()

		assertNoError(t, err)
	})

	t.Run("zero datetime fails", func(t *testing.T) {
		dt := shared.Datetime{}

		err := dt.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})

	t.Run("future datetime fails", func(t *testing.T) {
		// Create a datetime with AllowFuture, then validate with normal Validate
		futureTime := time.Now().Add(24 * time.Hour)
		dt, _ := shared.NewDatetimeAllowFuture(futureTime)

		err := dt.Validate()

		assertError(t, err)
		assertErrorCode(t, err, kernel.EInvalid)
	})
}

func TestDatetime_EdgeCases(t *testing.T) {
	t.Run("handles different timezones", func(t *testing.T) {
		locations := []string{
			"America/New_York",
			"Europe/London",
			"Asia/Tokyo",
			"Australia/Sydney",
		}

		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		for _, locName := range locations {
			t.Run(locName, func(t *testing.T) {
				loc, err := time.LoadLocation(locName)
				if err != nil {
					t.Skip("timezone not available")
				}

				localTime := baseTime.In(loc)
				dt, err := shared.NewDatetime(localTime)

				assertNoError(t, err)

				// Should be converted to UTC
				if dt.Time().Location() != time.UTC {
					t.Errorf("expected UTC, got %v", dt.Time().Location())
				}

				// Should represent the same moment
				if !dt.Time().Equal(localTime.UTC()) {
					t.Errorf("time not properly converted")
				}
			})
		}
	})

	t.Run("handles nanosecond precision", func(t *testing.T) {
		testTime := time.Date(2024, 1, 1, 12, 0, 0, 123456789, time.UTC)

		dt, err := shared.NewDatetime(testTime)

		assertNoError(t, err)
		if dt.Time().Nanosecond() != testTime.Nanosecond() {
			t.Errorf("nanosecond precision lost: got %d, want %d",
				dt.Time().Nanosecond(), testTime.Nanosecond())
		}
	})
}
