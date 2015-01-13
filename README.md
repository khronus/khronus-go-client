Go client for Khronus
=======================

A simple Go client for Khronus.

It works by buffering metrics to send them later in a batch way.

## Install

```bash
    go get -u github.com/despegar/khronus-go-client
```

## How to use it

#### 1) Create the client
```go
    import . "github.com/despegar/khronus-go-client"

    func main() {
        client := Client{}

        urls := []string{
                    "http://KhronusHost1:KhronusPort/khronus/metrics",
                    "http://KhronusHost2:KhronusPort/khronus/metrics",
                  }

        client.Config().Urls(urls).Interval(30)
    }
```
The urls array is a list of all api endpoints of the khronus servers
The send interval is how often the metrics are sent to Khronus.

#### 2) Using a client to measure
```go
client.Timer("metric-name").Record(300)
```

This will push to Khronus a metric named "metric-name" with a measured value of 300 milliseconds with the current timestamp.

```go
client.Timer("metric-name").RecordWithTs(time.Now().Unix()*1000, 300)
```
This will push to Khronus a metric with a particular timestamp (timestamps are in milliseconds).

```go
client.Timer("metric-name").Record(300, 100, 4400, 35)
```

You can also pass multiple value for the same timestamp.

Available type of metrics are Counter, Timer, Gauge. All use the Record() and RecordWithTs() methods to store and send metics<sup>1</sup>

```go
client.Timer("metric-timer").Record(300)
client.Gauge("metric-gauge").Record(300)
client.Counter("metric-counter").Record(300)
```

#### 3) Using a channel to send metrics

You can also use a channel to send metrics 

```go

    import . "github.com/despegar/khronus-go-client"

    func main() {

        channel := make(chan *Metric, 10)
        client := Client{}
        
        // Leave interval and urls as default
        // interval : 30 seconds 
        // urls :127.0.0.1:80
            
        client.Config().Channel(channel) 
        channel <- Gauge("test").Record(100)       
    }

```

<sup>1</sup> Of course this is idiomatically incorrect but simplify code. Timer should use RecordTime, counter should use Increment, etc.
