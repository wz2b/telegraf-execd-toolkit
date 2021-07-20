# Telegraf Execd Toolkit

This library is a set of useful tools for quickly building a telegraf
external plugin.  It includes some wrappers around:

* Metric output, in a format that's compatible with telegraf but without a lot
  of boilerplate
* Logging, in a way that's configurable by the user via the command line and
  can feed log messages into telegraf as metrics, or be directed elsewhere
  

## Metric Generation and Encoding

The toolkit outputs metrics in
[influx line protocol format](https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/),
a well-documented, simple format that telegraf understands (in fact, if you don't
specify a different format, line protocol is the default).  The format is simple  From
Influx Data's description:

```
// Syntax
<measurement>[,<tag_key>=<tag_value>[,<tag_key>=<tag_value>]] <field_key>=<field_value>[,<field_key>=<field_value>] [<timestamp>]

// Example
myMeasurement,tag1=value1,tag2=value2 fieldKey="fieldValue" 1556813561098000000
```

This format can be built a variety of ways including using influxdata's
[own library](https://github.com/influxdata/line-protocol) or, in some _very_ simple cases, `fmt.Sprintf`
(not really recommended).

This library provides an alternate approach that removes some of the boilerplate of using
the Influx Data library.  The metrics that it creates are still MutableMetrics so you have access to the
normal MutableMetric methods but with a little bit added:

  * There is are fluent methods to add fields and tags, `.WithTags()` and `.WithFields()`, as well as
to set the timestamp of the record using `.WithTime()`
  * The timestamp defaults to `time.Now()` (which is usually what you want)
  * When writing a field, if the field is a go _error_ type it will get converted to a string using `error.Error()`
  * The encoder shares buffers to avoid buffer object churn
  * You can create an encoder pool that is thread-safe

Since the intent here is to create a metric then write it, a metric encoder is included.  The encoder
can write to either an `io.Writer` or `[]byte`.

The simplest possible input plugin using this library looks like this:

```go
package main
import (
	encoder "github.com/wz2b/telegraf-execd-toolkit/line-metric-encoder"
    "os"
	"time"
)

func main() {
  metricPool := encoder.NewMetricEncoderPool()

  i := 0
  for {
  	metric := metricPool.NewMetric("my_counter").WithField("count", i)
  	metric.Write(os.Stdout)
  	i++
  	time.Sleep(30 * time.Second)  // or whatever you want to do
  }
}
```

Chaining calls against the metric can be helpful or cumbersome depending on the
complexity of your metrics.  If you don't want to chain, that's perfectly fine.
_With_ shouldn't be interpreted to imply object creation - it doesn't.  This is
perfectly acceptable:

```go
metric := metricPool.NewMetric("my_counter")
metric.WithTime(time.Now()) // if you want to set it to something else
metric.WithField("count", i)
metric.WithTag("importance", "none")
metric.Write(os.Stdout)
```

and of course since our metric is still a MutableMetric you can use `.AddField()` and `.AddTag()`
directly - however you won't get the error-to-string conversion (and any other conversions that
get added in the future) so you are encouraged to use `.WithField()` and `.WithTag()` instead.

`.WithTime()` can be useful in situations where you get the data with its own timestamp, or
if you want to write a bunch of metrics sharing the same timestamp - or any other reason
you want to set the time to something other than the default (`time.Now()`)

## Logging

The toolkit provides the ability to direct logs to a format and location of your choosing.
Note that this is not about collecting logs from monitored services/servers/devices.  This
is about logging by and of the plugin itself, for debugging or if the external plugin
itself has something happen that it wants you to note.  An external plug is, of course,
always able to emit metrics about itself if it chooses.

Most plugins probably don't need this.  They can log to stderr in whatever format they want
and these entries will end up in telegraf's log (perhaps in /var/log/telegraf/ depending on
your system).  This logging system gives you a few more options, including redirecting log
messages to the stream of metrics or to their own file.

Some features include:

  * You can choose to write logs to standard output, standard error, or a file 
  * You can choose the format: logfmt, json, or line protocol format.  The default log format is logfmt.
    Writing to line-protocol is useful if you have occasional logs and you want to redirect them
    to the metric stream.  This is especially useful if the destination is something like mqtt.
  * Log rotation and space management (when writing to log files) using 
    [Lumberjack]("gopkg.in/natefinch/lumberjack.v2").

Logging can be configured via command line options or manually.  The logger that is created is
a level-aware go-kit/log.  Example usage:

```go
import (
    "github.com/go-kit/kit/log/level"
    tlogger "github.com/wz2b/telegraf-execd-toolkit/telegraf-logger"
)

var klog kitlog.Logger

func main() {
	logFactory, err := tlogger.NewTelegrafLoggerConfiguration(true)

	//
	// If the command line options aren't being used, set parameters manually here
	// logFactory.LogFile = "/var/log/my_agent.log" // or "stderr" or "stdout"
	// logFactory.LogLevel = "info" // could be "info", "error", "warn" (the default), "debug", "all", or "none"
	// logFactory.LogFormat = "line" // could be "line", "json", or "logfmt" (the default)
	// logFactory.LogMetric = "log" // only used if the log format is "line"
	//
	// Otherwise, set these flags on the command line:
	//  --log  stdout   # destination
	//  --log-format line
	//  --log-level info
	//  --log-metric xyzzy
	//
	if err != nil {
		panic(err)
	}
    klog = logFactory.Create()	
	
    // ...

	level.Error(klog).Log("msg", "This is an error", "error", err)
    level.Debug(klog).Log("msg", "This is a debug message", "field", "value1", "code", "8008135")
```

Generally speaking, Telegraf doesn't appreciate you sending logs to standard output.  If you try
to do that using this library, one of two things will happen:

  * If your log format is line, then the log message will come out formatted as if it were a metric.
    It will be processed by telegraf as any other metric, meaning it will get written to your
    configured telegraf outputs pending it is not first filtered or dropped by some rule you have
    written into telegraf.conf
  * If you attempt to use standard output with another format, lines will be prepended with a
    comment mark _#_ consistent with line protocol standards.  If data were being piped directly
    into influxdb this line would be ignored.  Telegraf still sees this as an error, though, and
    will write it to the telegraf log.
 

Generally speaking, telegraf logging for execd plugins would benefit from some improvement.  If
it's an issue, you can always direct your plugin's logs to its own file.  For most purposes,
writing to standard error (the default) is sufficient.

   

For more information on how level-based logging works in go-kit's logger see
https://pkg.go.dev/github.com/go-kit/kit/log/level

