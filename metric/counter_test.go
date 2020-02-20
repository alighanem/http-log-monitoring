package metric_test

import (
	"reflect"
	"testing"

	"github.com/ali.ghanem/http-log-monitoring/metric"
)

func TestCounter_Value(t *testing.T) {
	t.Run("increment", func(t *testing.T) {
		counter := metric.NewCounter()

		// first increment
		var expected int64 = 10
		counter.Inc(int64(10))
		actual := counter.Value()
		if actual != expected {
			t.Fatal("unexpected value", "expected", expected, "actual", actual)
		}

		// second
		expected = 35
		counter.Inc(int64(25))
		actual = counter.Value()

		if counter.Value() != expected {
			t.Fatal("unexpected value", "expected", expected, "actual", actual)
		}
	})
}

func TestCounterVec_Value(t *testing.T) {
	type testCase struct {
		Label         string
		Increment     int64
		ExpectedValue int64
	}

	cases := map[string]testCase{
		"existing label": {
			Label:         "label1",
			Increment:     10,
			ExpectedValue: 20,
		},
		"new label": {
			Label:         "new_label",
			Increment:     33,
			ExpectedValue: 33,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cv := setupCounterVec(t)
			cv.Inc(c.Label, c.Increment)

			actual := cv.Value(c.Label)
			if c.ExpectedValue != actual {
				t.Fatal("unexpected value", "expected", c.ExpectedValue, "actual", actual)
			}
		})
	}
}

func TestCounterVec_AllValues(t *testing.T) {
	type testCase struct {
		Label          string
		Increment      int64
		ExpectedValues map[string]int64
	}

	cases := map[string]testCase{
		"add new label": {
			Label:     "new_label",
			Increment: 22,
			ExpectedValues: map[string]int64{
				"label1":    10,
				"label2":    35,
				"new_label": 22,
			},
		},
		"increment existing label": {
			Label:     "label2",
			Increment: 18,
			ExpectedValues: map[string]int64{
				"label1": 10,
				"label2": 53,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cv := setupCounterVec(t)
			cv.Inc(c.Label, c.Increment)

			actual := cv.AllValues()
			if !reflect.DeepEqual(c.ExpectedValues, actual) {
				t.Fatal("unexpected values", "expected", c.ExpectedValues, "actual", actual)
			}
		})
	}
}

func setupCounterVec(t *testing.T) *metric.CounterVec {
	cv := metric.NewCounterVec()

	cv.Inc("label1", 10)
	cv.Inc("label2", 35)

	return cv
}
