package metric

import (
	"sync"
	"time"
)

// TimeSeries is a collection of counters based on time
type TimeSeries struct {
	sync.RWMutex
	series map[time.Time]int64
}

func NewTimeSeries() *TimeSeries {
	return &TimeSeries{
		series: map[time.Time]int64{},
	}
}

// Increments the time counter value
func (t *TimeSeries) Inc(date time.Time, value int64) {
	// todo remove time.now() before release it
	if date.After(time.Now()) {
		return
	}

	t.Lock()
	defer t.Unlock()

	if _, ok := t.series[date]; !ok {
		t.series[date] = value
		return
	}

	t.series[date] += value
}

// Sums the total of counters since a date
func (t *TimeSeries) CountSince(since time.Time) int64 {
	t.RLock()
	defer t.RUnlock()

	var total int64
	for date, hits := range t.series {
		// todo remove time.now() before release it
		if date.Before(since) || date.After(time.Now()) {
			continue
		}
		total += hits
	}

	return total
}

// Cleans older time series
func (t *TimeSeries) Clean(before time.Time) int64 {
	t.Lock()
	defer t.Unlock()

	var cleaned int64
	newSeries := make(map[time.Time]int64)
	for date, hits := range t.series {
		if date.Before(before) {
			cleaned++
			continue
		}
		newSeries[date] = hits
	}

	t.series = newSeries
	return cleaned
}
