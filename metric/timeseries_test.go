package metric_test

import (
	"testing"
	"time"

	"github.com/ali.ghanem/http-log-monitoring/metric"
)

func TestTimeSeries_CountSince(t *testing.T) {
	type testCase struct {
		Date          time.Time
		Increment     int64
		CountSince    time.Time
		ExpectedCount int64
	}

	startDate := time.Date(2020, 02, 20, 10, 25, 32, 0, time.UTC)

	cases := map[string]testCase{
		"new date": {
			Date:          startDate.Add(55 * time.Second),
			Increment:     3,
			CountSince:    startDate.Add(55 * time.Second),
			ExpectedCount: 3,
		},
		"existing date": {
			Date:          startDate.Add(30 * time.Second),
			Increment:     5,
			CountSince:    startDate.Add(28 * time.Second),
			ExpectedCount: 8,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ts := setupTimeSeries(t, startDate)
			ts.Inc(c.Date, c.Increment)

			actual := ts.CountSince(c.CountSince)
			if c.ExpectedCount != actual {
				t.Fatal("unexpected count", "expected", c.ExpectedCount, "actual", actual)
			}
		})
	}
}

func TestTimeSeries_Clean(t *testing.T) {
	startDate := time.Date(2020, 02, 20, 10, 25, 32, 0, time.UTC)

	type testCase struct {
		CleanBefore       time.Time
		ExpectedRemaining int64
	}

	cases := map[string]testCase{
		"older values cleaned": {
			CleanBefore:       startDate.Add(25 * time.Second),
			ExpectedRemaining: 3,
		},
		"only new values no clean done": {
			CleanBefore:       startDate.Add(-2 * time.Minute),
			ExpectedRemaining: 5,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ts := setupTimeSeries(t, startDate)
			ts.Clean(c.CleanBefore)

			actual := ts.CountSince(c.CleanBefore)

			if c.ExpectedRemaining != actual {
				t.Fatal("unexpected remaining", "expected", c.ExpectedRemaining, "actual", actual)
			}
		})
	}
}

func setupTimeSeries(t *testing.T, starDate time.Time) *metric.TimeSeries {
	ts := metric.NewTimeSeries()
	incDate := starDate

	for i := 1; i <= 5; i++ {
		incDate = incDate.Add(10 * time.Second)
		ts.Inc(incDate, 1)
	}

	return ts
}
