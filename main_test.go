package main

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/ali.ghanem/http-log-monitoring/commonlog"
	"github.com/ali.ghanem/http-log-monitoring/metric"
	"github.com/ali.ghanem/http-log-monitoring/timetest"
)

func TestLogMonitor_HandleEvent(t *testing.T) {

	type testCase struct {
		Event            commonlog.Event
		ExpectedHits     int64
		ExpectedSections map[string]int64
		ExpectedCalls    map[string]int64
		ExpectedSize     int64
	}

	cases := map[string]testCase{
		"new event": {
			Event: commonlog.Event{
				Host:    "10.20.55.10",
				RFC931:  "-",
				User:    "my_user",
				Date:    time.Now().Truncate(time.Second).Add(-10 * time.Second),
				Request: "DELETE /markets/cutting-edge/vertical HTTP/1.1",
				Status:  http.StatusOK,
				Bytes:   65040,
				Section: "markets",
			},
			ExpectedHits: 11,
			ExpectedSize: 75040,
			ExpectedSections: map[string]int64{
				"pages":   10,
				"markets": 1,
			},
			ExpectedCalls: map[string]int64{
				"total":   11,
				"succeed": 1,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := setupLogMonitor(t)

			m.HandleEvent(c.Event)

			actual := m.HitsSeries.CountSince(time.Now().Add(-1 * time.Minute))
			if c.ExpectedHits != actual {
				t.Error("unexpected hits number", "expected", c.ExpectedHits, "actual", actual)
			}

			actual = m.Bytes.Value()
			if c.ExpectedSize != actual {
				t.Error("unexpected size", "expected", c.ExpectedSize, "actual", actual)
			}

			actualValues := m.Calls.AllValues()
			if !reflect.DeepEqual(c.ExpectedCalls, actualValues) {
				t.Error("unexpected calls", "expected", c.ExpectedCalls, "actual", actualValues)
			}

			actualValues = m.Sections.AllValues()
			if !reflect.DeepEqual(c.ExpectedSections, actualValues) {
				t.Error("unexpected sections visited", "expected", c.ExpectedCalls, "actual", actualValues)
			}
		})
	}
}

func TestLogMonitor_Statistics(t *testing.T) {
	type testCase struct {
		Events             []commonlog.Event
		ExpectedStatistics Statistics
	}

	cases := map[string]testCase{
		"get statistics": {
			Events: []commonlog.Event{
				{
					Host:    "10.20.55.10",
					RFC931:  "-",
					User:    "my_user",
					Date:    time.Now().Truncate(time.Second).Add(-10 * time.Second),
					Request: "DELETE /markets/cutting-edge/vertical HTTP/1.1",
					Status:  http.StatusOK,
					Bytes:   65040,
					Section: "markets",
				},
				{
					Host:    "122.20.55.10",
					RFC931:  "-",
					User:    "my_user",
					Date:    time.Now().Truncate(time.Second).Add(-15 * time.Second),
					Request: "DELETE /markets/cutting-edge/vertical HTTP/1.1",
					Status:  http.StatusOK,
					Bytes:   65040,
					Section: "markets",
				},
				{
					Host:    "192.168.0.1",
					RFC931:  "-",
					User:    "user 2",
					Date:    time.Now().Truncate(time.Second).Add(-20 * time.Second),
					Request: "PATCH /killer/models/deliver HTTP/2.0",
					Status:  http.StatusOK,
					Bytes:   90200,
					Section: "killer",
				},
				{
					Host:    "192.168.0.1",
					RFC931:  "-",
					User:    "user 2",
					Date:    time.Now().Truncate(time.Second).Add(-10 * time.Second),
					Request: "POST /killer/models/deliver HTTP/2.0",
					Status:  http.StatusInternalServerError,
					Bytes:   55700,
					Section: "killer",
				},
				{
					Host:    "192.168.0.1",
					RFC931:  "-",
					User:    "user 2",
					Date:    time.Now().Truncate(time.Second).Add(-10 * time.Second),
					Request: "PUT /killer/models/deliver HTTP/2.0",
					Status:  http.StatusBadRequest,
					Bytes:   55700,
					Section: "killer",
				},
			},
			ExpectedStatistics: Statistics{
				TopSections: []Section{
					{
						Name: "pages",
						Hits: 10,
					},
					{
						Name: "killer",
						Hits: 3,
					},
				},
				HitsByStatus: map[string]int64{
					"total":        15,
					"succeed":      3,
					"client_error": 1,
					"server_error": 1,
				},
				TotalBytes: 341680,
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := setupLogMonitor(t)

			for _, event := range c.Events {
				m.HandleEvent(event)
			}

			actual := m.Statistics(2)
			if !reflect.DeepEqual(c.ExpectedStatistics, actual) {
				t.Fatal("unexpected statistics", "expected", c.ExpectedStatistics, "actual", actual)
			}
		})
	}
}

func TestLogMonitor_CheckTrafficLoad(t *testing.T) {
	type testCase struct {
		LastAlert         *Alert
		Hits              int64
		Interval          time.Duration
		Threshold         int64
		ExpectedAlert     *Alert
		ExpectedLastAlert *Alert
	}

	triggeredAt := time.Date(2006, 01, 02, 15, 04, 05, 000, time.UTC)

	cases := map[string]testCase{
		"no alert occurred traffic is normal": {
			LastAlert:         nil,
			Hits:              1000,
			Interval:          time.Minute,
			Threshold:         100,
			ExpectedAlert:     nil,
			ExpectedLastAlert: nil,
		},
		"alert because it exceeds the threshold": {
			LastAlert: nil,
			Hits:      10000,
			Interval:  time.Minute,
			Threshold: 100,
			ExpectedAlert: &Alert{
				Exceed:      true,
				Hits:        10000,
				AverageRate: 166,
				TriggeredAt: triggeredAt,
			},
			ExpectedLastAlert: &Alert{
				Exceed:      true,
				Hits:        10000,
				AverageRate: 166,
				TriggeredAt: triggeredAt,
			},
		},
		"new alert because traffic load increased": {
			LastAlert: &Alert{
				Exceed:      true,
				Hits:        10000,
				AverageRate: 166,
				TriggeredAt: triggeredAt,
			},
			Hits:      20000,
			Interval:  time.Minute,
			Threshold: 100,
			ExpectedAlert: &Alert{
				Exceed:      true,
				Hits:        20000,
				AverageRate: 333,
				TriggeredAt: triggeredAt,
			},
			ExpectedLastAlert: &Alert{
				Exceed:      true,
				Hits:        20000,
				AverageRate: 333,
				TriggeredAt: triggeredAt,
			},
		},
		"traffic load is back to normal": {
			LastAlert: &Alert{
				Exceed:      true,
				Hits:        20000,
				AverageRate: 333,
				TriggeredAt: triggeredAt,
			},
			Hits:      1000,
			Interval:  time.Minute,
			Threshold: 100,
			ExpectedAlert: &Alert{
				Exceed:      false,
				Hits:        1000,
				AverageRate: 16,
				TriggeredAt: triggeredAt,
			},
			ExpectedLastAlert: nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			timetest.FreezeTime()
			defer timetest.UnfreezeTime()

			m := setupLogMonitor(t)
			m.LastAlert = c.LastAlert

			alert := m.CheckTrafficLoad(c.Hits, c.Interval, c.Threshold)
			if !reflect.DeepEqual(c.ExpectedAlert, alert) {
				t.Fatal("unexpected alert", "expected", c.ExpectedAlert, "actual", alert)
			}

			if !reflect.DeepEqual(c.ExpectedLastAlert, m.LastAlert) {
				t.Fatal("unexpected alert saved", "expected", c.ExpectedLastAlert, "actual", m.LastAlert)
			}
		})
	}
}

func setupLogMonitor(t *testing.T) *LogMonitor {
	m := LogMonitor{
		Sections:   metric.NewCounterVec(),
		HitsSeries: metric.NewTimeSeries(),
		Calls:      metric.NewCounterVec(),
		Bytes:      metric.NewCounter(),
	}

	m.Bytes.Inc(10000)
	m.Sections.Inc("pages", 10)
	m.Calls.Inc(Total, 10)
	m.HitsSeries.Inc(time.Now().Truncate(time.Second).Add(-10*time.Second), 10)

	return &m
}
