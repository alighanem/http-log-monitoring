package main

import (
	"sort"
	"time"

	"github.com/ali.ghanem/http-log-monitoring/commonlog"
	"github.com/ali.ghanem/http-log-monitoring/metric"
)

// Service to monitor a log file in the W3C common format
type LogMonitor struct {
	// Sections contains the number of hits by section visited
	Sections *metric.CounterVec

	// HitsSeries stores the number of hits by bucket of time
	HitsSeries *metric.TimeSeries

	// Last alert occurred during monitoring
	LastAlert *Alert

	// Metrics
	Calls *metric.CounterVec
	Bytes *metric.Counter
}

// Represents a traffic alert
type Alert struct {
	Exceed      bool
	Hits        int64
	AverageRate int64
	TriggeredAt time.Time
}

// Statistics about traffic
type Statistics struct {
	TopSections  []Section
	HitsByStatus map[string]int64
	TotalBytes   int64
}

// Section visited
type Section struct {
	Name string
	Hits int64
}

// HandleEvent manages the log event by the monitor
func (l *LogMonitor) HandleEvent(event commonlog.Event) {
	l.Calls.Inc(Total, 1)
	switch {
	case IsInformational(event.Status):
		l.Calls.Inc(Info, 1)
	case IsSuccess(event.Status):
		l.Calls.Inc(Succeed, 1)
	case IsRedirection(event.Status):
		l.Calls.Inc(Redirected, 1)
	case IsClientError(event.Status):
		l.Calls.Inc(ClientError, 1)
	case IsServerError(event.Status):
		l.Calls.Inc(ServerError, 1)
	default:
		l.Calls.Inc(Unknown, 1)
	}

	l.Bytes.Inc(int64(event.Bytes))

	l.HitsSeries.Inc(event.Date, 1)

	l.Sections.Inc(event.Section, 1)
}

// CheckTrafficLoad check the traffic and may return an alert about important change on the load
func (l *LogMonitor) CheckTrafficLoad(hits int64, interval time.Duration, threshold int64) *Alert {
	avgRate := hits / int64(interval.Seconds())

	if avgRate >= threshold {
		// exceeds the threshold
		if l.LastAlert != nil && l.LastAlert.Hits == hits {
			// same number of hits => return the original alert
			return l.LastAlert
		}

		l.LastAlert = &Alert{
			Exceed:      true,
			Hits:        hits,
			AverageRate: avgRate,
			TriggeredAt: time.Now(),
		}
		return l.LastAlert
	}

	if l.LastAlert == nil {
		// no previous alert no need to return an alert to say that the traffic came back to normal
		return nil
	}

	// lower than the threshold => back to normal no more alert to follow
	l.LastAlert = nil
	return &Alert{
		Exceed:      false,
		Hits:        hits,
		AverageRate: avgRate,
		TriggeredAt: time.Now(),
	}
}

// Statistics returns statistics about the traffic generated
func (l *LogMonitor) Statistics(maxSections int) Statistics {
	sections := l.Sections.AllValues()
	// sort the sections
	cs := make([]Section, len(sections))
	i := 0
	for name, hits := range sections {
		cs[i] = Section{name, hits}
		i++
	}
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Hits > cs[j].Hits
	})

	if len(cs) >= maxSections {
		cs = cs[0:maxSections]
	}

	return Statistics{
		TopSections:  cs,
		HitsByStatus: l.Calls.AllValues(),
		TotalBytes:   l.Bytes.Value(),
	}
}

// HTTP status category
const (
	Total       = "total"
	Info        = "info"
	Succeed     = "succeed"
	Redirected  = "redirected"
	ClientError = "client_error"
	ServerError = "server_error"
	Unknown     = "unknown"
)
