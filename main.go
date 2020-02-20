package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ali.ghanem/http-log-monitoring/commonlog"
	"github.com/ali.ghanem/http-log-monitoring/metric"
	"github.com/hpcloud/tail"
)

func main() {
	config, err := ReadConfiguration()
	if err != nil {
		log.Fatal(fmt.Sprintf("cannot read configuration - err: %s", err))
	}

	logFile, err := os.OpenFile(config.LogOutput, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Println("failed to close file", "err", err)
		}
	}()

	// add multiple log outputs
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	log.Println("configuration read", config)

	var statsTicker = time.NewTicker(config.StatsDisplayInterval)
	var alertingTicker = time.NewTicker(config.TrafficLoadCheckInterval)
	var cleaningTimeSeriesTicker = time.NewTicker(config.CleaningInterval)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var monitor = LogMonitor{
		Sections:   metric.NewCounterVec(),
		HitsSeries: metric.NewTimeSeries(),
		Calls:      metric.NewCounterVec(),
		Bytes:      metric.NewCounter(),
	}

	log.Println("start monitoring")
	t, err := tail.TailFile(config.LogToMonitor, tail.Config{Follow: true, ReOpen: true, Poll: true})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-ctx.Done():
			log.Println("stopped")
			return
		case line := <-t.Lines:
			if ctx.Err() != nil {
				log.Println("context canceled")
				return
			}
			// consumes the logs
			if line.Err != nil {
				log.Println("cannot read line", "err", line.Err, "line", line.Text)
				continue
			}

			event, err := commonlog.Parse(line.Text)
			if err != nil {
				log.Println("cannot parse line", "err", err, "line", line.Text)
				continue
			}

			monitor.HandleEvent(event)
		case <-c:
			log.Println("stopping")
			statsTicker.Stop()
			alertingTicker.Stop()
			cancel()

		case <-statsTicker.C:
			// display statistics
			statistics := monitor.Statistics(config.StatsTopSectionsCount)

			log.Println("number of events received", statistics.HitsByStatus[Total])
			for status, hits := range statistics.HitsByStatus {
				if status == Total {
					continue
				}

				log.Println(fmt.Sprintf("hits by status %s", status), hits)
			}

			log.Println("top sections visited", len(statistics.TopSections))
			for _, s := range statistics.TopSections {
				log.Println("section", s.Name, "hits", s.Hits)
			}

			log.Println("total bytes", formatSize(statistics.TotalBytes))

		case <-alertingTicker.C:
			// check if traffic generated an alert to display
			hits := monitor.HitsSeries.CountSince(time.Now().Add(-1 * config.TrafficLoadPeriod))

			alert := monitor.CheckTrafficLoad(hits, config.TrafficLoadPeriod, config.TrafficThreshold)
			if alert == nil {
				// no alerting
				log.Println(fmt.Sprintf("traffic is normal - %v", hits))
				continue
			}

			if alert.Exceed {
				log.Println(fmt.Sprintf("high traffic generated an alert - hits: %v - rate: %v hits/s - triggered at: %s",
					hits, alert.AverageRate, alert.TriggeredAt))
				continue
			}

			log.Println(fmt.Sprintf("traffic came back to normal - hits: %v - rate: %v hits/s", hits, alert.AverageRate))

		case <-cleaningTimeSeriesTicker.C:
			// clean the older time series to avoid important memory usage
			log.Println("cleaning time series")
			cleaned := monitor.HitsSeries.Clean(time.Now().Add(-1 * config.TrafficLoadPeriod))
			log.Println("time series cleaned", cleaned)
		}
	}
}

// format a size in a human readable string
func formatSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
