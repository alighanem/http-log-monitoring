package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	LogToMonitor string // File path of the log to monitor

	StatsDisplayInterval  time.Duration // Time to wait before displaying the statistics of the consumption
	StatsTopSectionsCount int           // The number of sections with maximum hits

	TrafficLoadCheckInterval time.Duration // Check alert interval
	TrafficLoadPeriod        time.Duration // Period to verify for the traffic load
	TrafficThreshold         int64         // Traffic threshold in number of requests / second

	CleaningInterval time.Duration // Interval to clean time series

	LogOutput string // File path to output logs of the monitor execution
}

func ReadConfiguration() (config Configuration, err error) {
	config.LogToMonitor, err = readString("LOG_TO_MONITOR")
	if err != nil {
		return config, err
	}

	config.LogOutput, err = readString("LOG_OUTPUT")
	if err != nil {
		return config, err
	}

	config.StatsDisplayInterval, err = readDuration("STATISTICS_DISPLAY_INTERVAL")
	if err != nil {
		return config, err
	}

	config.StatsTopSectionsCount, err = readInt("STATISTICS_TOP_SECTIONS_COUNT")
	if err != nil {
		return config, err
	}

	config.TrafficLoadCheckInterval, err = readDuration("TRAFFIC_LOAD_CHECK_INTERVAL")
	if err != nil {
		return config, err
	}

	config.TrafficLoadPeriod, err = readDuration("TRAFFIC_LOAD_PERIOD")
	if err != nil {
		return config, err
	}

	config.TrafficThreshold, err = readInt64("TRAFFIC_THRESHOLD")
	if err != nil {
		return config, err
	}

	config.CleaningInterval, err = readDuration("CLEANING_INTERVAL")
	if err != nil {
		return config, err
	}

	return config, nil
}

func readString(key string) (string, error) {
	raw := os.Getenv(key)
	if len(raw) == 0 {
		return "", fmt.Errorf("key %s not found", key)
	}

	return raw, nil
}

func readDuration(key string) (time.Duration, error) {
	raw, err := readString(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		return duration, fmt.Errorf("cannot parse key: %s - err %w", key, err)
	}

	return duration, nil
}

func readInt(key string) (int, error) {
	raw, err := readString(key)
	if err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cannot parse key: %s - err %w", key, err)
	}

	return value, nil
}

func readInt64(key string) (int64, error) {
	raw, err := readString(key)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse key: %s - err %w", key, err)
	}

	return value, nil
}
