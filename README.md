# HTTP log monitoring

## Description

Console program to monitor any log file written in w3c-formatted HTTP access log 
(https://www.w3.org/Daemon/User/Config/Logging.html).

It will: 
 * Generate traffic load alerts if it exceeds.
 * Inform that the traffic is or back to normal.
 * Display statistics about the traffic. 

## Launch & Configuration

To run the monitoring, you have to configure some environment variables:


| Variable                        | Type      | Description                                            | Example                            |
| ------------------------------- | --------- |  ----------------------------------------------------- | ---------------------------------- |
| `LOG_TO_MONITOR`                | string    |  Path to the log file to monitor                       | "/tmp/access.log"                  |
| `STATISTICS_DISPLAY_INTERVAL`   | duration  |  Regular interval to display traffic statistics        | "10s" for 10 seconds               |
| `STATISTICS_TOP_SECTIONS_COUNT` | int       |  Number of sections with more hits to display          | "3"                                |
| `TRAFFIC_LOAD_CHECK_INTERVAL`   | duration  |  Regular interval to check the traffic load            | "1m" for 1 minute                  |
| `TRAFFIC_LOAD_PERIOD`           | duration  |  Period to verify for traffic load                     | "2m" for 2 minutes                 |
| `TRAFFIC_THRESHOLD`             | int       |  Traffic threshold (number of requests per second)     | "100" 100 requests / sec           |
| `CLEANING_INTERVAL`             | duration  |  Interval to clean older time series                   | "5m" cache cleaned every 5 minutes |
| `LOG_OUTPUT`                    | string    |  Path to program logs                                  | "out.log"                          |
 
This is an example of the command to execute the program:

    LOG_TO_MONITOR="logs.log" STATISTICS_DISPLAY_INTERVAL="10s" STATISTICS_TOP_SECTIONS_COUNT="3" 
    TRAFFIC_LOAD_CHECK_INTERVAL="20s" TRAFFIC_LOAD_PERIOD="2m" TRAFFIC_THRESHOLD="10" CLEANING_INTERVAL="45s" 
    LOG_OUTPUT="out.log" go run .
 
 
## External libs

 * https://github.com/hpcloud/tail: lib to monitor any modification on a log file.
 * https://github.com/bouk/monkey: lib to mock behaviors during tests 

## Future improvements

To improve the program, we can:

 * Add multiple workers to parse the log files.
   Currently, there is only one worker to parse the log lines. 
   If the load becomes very important, maybe we can reach a point 
   that our code cannot handle the amount of lines written.
   
 * Use a real system to manage metrics like prometheus.

 * Use a metric storage like VictoriaMetrics. For the need of the exercise, 
   the metrics store have been simplified.
   It is only in memory. We had to use mutexes to avoid concurrency access on metrics.
   
 * Separate the features: all the features are in the same program, we can have 3 programs: 
    * one dedicated to consume and parse the logs and to feed the metrics store.
    * one api to get statistics about the metrics
    * one alerting program based on prometheus alerts.
    
  * Improve the computation of the traffic load: actually, the program will compute 
    an average of the number of hits. So, it will not differentiate a spike from a increasing traffic load.
    Based on prometheus, we can compute the average by using rate function.

## Miscellaneous

 * To test the program in local, use this tool (https://github.com/mingrammer/flog)
  to generate random logs.

