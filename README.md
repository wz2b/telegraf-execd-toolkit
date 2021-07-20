# Telegraf Execd Toolkit

This library is a set of (hopefully) useful tools for building
[telegraf](https://www.influxdata.com/time-series-platform/telegraf/)
external plugins. [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/)
is a plugin-driven server agent for collecting and sending metrics and events from databases, systems,
and IoT sensors. [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/)
is written in Go and compiles into a single binary with no external dependencies,
and requires a very minimal memory footprint.

There are two main reasons to consider using this library:

    * Reduce boilerplate code when it comes to generating output (metrics)
    * Make plugin log handling (logs about the plugin itself) flexible but consistent across all external plugins

Metric serialization (encoding the metrics to bytes, to a string, or to an io.Writer like standard output) is
accomplished with an easy-to-use encoding pool that tries to minimize buffer churn.  Logging can be configured
by parsing command line flags, which is optional but encouraged.  If used, a plugin can be configured as:

```toml
[[inputs.execd]]
  command = [
        "myExternalPlugin",
        "-log", "stdout",
        "-log-level", "debug",
        "-log-format", "line",
        "-log-metric", "log" ]
```
## Something to consider

Before writing your own external plug-in, please look at the
[pretty extensive list of internal plug-ins](https://docs.influxdata.com/telegraf/v1.19/plugins/).  If
you're writing your own plugin just for fun or for the experience, by all means go for it.  Just keep
in mind that a number of times I have thought I needed a custom plugin when one of the standard ones
might have worked - for example, the HTTP input can get data from servers using one of the standard
input formats (such as json) which probably covers 90% of REST use cases.

Most importantly, have fun with the idea of using telegraf as a general-purpose agent.  I wrote this
because Telegraf is one of my favorite open-source projects and it's the glue of my IoT and IIoT world.

## Metric Generation and Encoding

The toolkit outputs metrics in
[influx line protocol format](https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/),
a well-documented, simple format that telegraf understands.  The format is simple:

```
// Syntax
<measurement>[,<tag_key>=<tag_value>[,<tag_key>=<tag_value>]] <field_key>=<field_value>[,<field_key>=<field_value>] [<timestamp>]

// Example
myMeasurement,tag1=value1,tag2=value2 fieldKey="fieldValue" 1556813561098000000
```

This format can be built a variety of ways including using influxdata's
[line protocol library](https://github.com/influxdata/line-protocol) or, in some _very_ simple cases, `fmt.Sprintf`
(that's not really recommended, but possible).

This library provides an alternate approach that removes some of the boilerplate required by
the Influx Data library.  The metrics that it creates are still MutableMetrics so you have access to the
normal MutableMetric methods but with a little bit of functionality added on top:

    * A fluent interface for adding fields and tags
    * The timestamp defaults to `time.Now()` (you can override this behavior)
    * Fields of type _error_ are written as strings using `error.Error()`
    * Thread-safe, shared serialization buffers avoid buffer object churn

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
  	time.Sleep(60 * time.Second)  // or whatever you want to do
  }
}
```

Chaining calls against the metric can be helpful in some cases, but if you don't want to chain calls
that's fine too.  _With_ shouldn't be interpreted to imply object copying or creation - it doesn't.  This is
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

Both `.AddField()` and `.WithField()` silently ignore (and don't add) any field whose value is nil
or for some reason not convertible to one of the standard line metric types (numbers, strings, and
in the case of `.WithField()` also _error_)

### Metric Builder

it is possible to build metrics with an even more fluent interface.

```go
metric := mp.NewMetric("weather").
	WithTime(*t).
	BuildTag("station").Value("My_House").
	BuildField("temperature").Value(112.5").
	BuildField("how_I_feel").Value("hot").
		...
	Write(os.Stdout)
```

There is also a more specialized `ValueIfNoErr( value interface{}, err error )` whose purpose is to 
emit a field only there is no error (i.e. err = nil).  This is completely optional, and mainly useful
if the accessor for the value returns `( interface{}, error )`.  In this example, `observation.GetTemperature()`
returns a float with a non-nil error if the value could not be fetched:

```go
metric := mp.NewMetric("observation").
	BuildField("temperature").ValueIfNoErr(observation.GetTemperature())
        ...
    Write(os.Stdout)
```


## Plugin Logging

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

# Questions
If you have problems (even if it's just a question) please
[create an issue](https://github.com/wz2b/telegraf-execd-toolkit/issues)
