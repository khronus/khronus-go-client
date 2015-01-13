package khronusgoapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	buffer map[string]*Metric
	mu     sync.RWMutex
	config *clientConfig
}

type clientConfig struct {
	interval uint64
	urls     []string
	channel  chan *Metric
	mu       sync.Mutex
}

func (c *Client) defaultConfig() *clientConfig {
	if c.config != nil {
		return c.config
	} else {
		c.config = &clientConfig{}
	}
	c.config.mu.Lock()
	if c.config.interval == 0 {
		c.config.interval = 30
	}

	if c.config.urls == nil {
		c.config.urls = []string{"http://localhost:80"}
	}

	if c.buffer == nil {
		c.buffer = make(map[string](*Metric))

	}

	if c.config.channel == nil {
		c.config.channel = make(chan *Metric, 100)
	}

	c.config.mu.Unlock()

	go c.sender()

	return c.config
}

func (c *Client) Config() *clientConfig {
	cc := c.defaultConfig()

	return cc
}

func (cc *clientConfig) Urls(urls []string) *clientConfig {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.urls = urls
	return cc
}

func (c *Client) String() string {
	return fmt.Sprintf("%#v", c.buffer)
}

func (cc *clientConfig) Interval(interval uint64) *clientConfig {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.interval = interval
	return cc
}

func (cc *clientConfig) Channel(channel chan *Metric) *clientConfig {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.channel = channel
	return cc
}

func (c *Client) newMetric(name string, t Type) *Metric {
	c.mu.RLock()
	if m, ok := c.buffer[name]; ok && m != nil {
		if m.Type != t {
			return nil
		}
		t := m
		c.mu.RUnlock()
		return t
	} else {
		m := &Metric{Name: name, Type: t}
		c.mu.RUnlock()
		c.mu.Lock()
		c.buffer[name] = m
		c.mu.Unlock()
		return m
	}
}

func (c *Client) Gauge(name string) *Metric {
	return c.newMetric(name, gauge)
}

func (c *Client) Counter(name string) *Metric {
	return c.newMetric(name, counter)
}

func (c *Client) Timer(name string) *Metric {
	return c.newMetric(name, timer)
}

func (c *Client) addMetric(metric *Metric) {
	if metric == nil {
	} else {
		c.mu.Lock()
		defer c.mu.Unlock()
		if m, ok := c.buffer[metric.Name]; ok && m != nil {
			m.Append(metric)
		} else {
			c.buffer[metric.Name] = metric
		}
	}
}

func (c *Client) emptyBuffer() {
	c.mu.Lock()
	for k, _ := range c.buffer {
		//c.buffer[k].Measurements = []Measure{}
		delete(c.buffer, k)
	}
	c.mu.Unlock()
}

func (c *Client) toJson() (string, error) {
	t := make([]*Metric, len(c.buffer))
	i := 0
	for _, v := range c.buffer {
		t[i] = v
		i++
	}

	js, err := json.Marshal(t)

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("{\"metrics\":%s}\n", string(js)), err
}

func (c *Client) sender() {
	rrindex := 0 // simple round-robin index

	c.config.mu.Lock()
	urlsize := len(c.config.urls)
	tick := time.Tick(time.Duration(c.config.interval) * time.Second)
	c.config.mu.Unlock()

	client := &http.Client{}
	client.Timeout = time.Duration(5) * time.Second

	for {
		select {
		case m := <-c.config.channel:
			{
				c.addMetric(m)
			}
		case <-tick:
			{
				c.config.mu.Lock()
				urlsize = len(c.config.urls)
				tick = time.Tick(time.Duration(c.config.interval) * time.Second)
				c.config.mu.Unlock()

				if len(c.buffer) == 0 {
					continue
				}

				fmt.Println("Tick")
				for key, _ := range c.config.urls {
					url := c.config.urls[(key+rrindex)%urlsize]

					c.mu.Lock()
					js, err := c.toJson()
					c.mu.Unlock()
					c.emptyBuffer()

					if err != nil {
						log.Println("Error marshaling json DataPoints", err)
						// Something really wrong is happening here
					}

					fmt.Printf("%s", js)

					resp, err := client.Post(url, "application/json", bytes.NewBufferString(js))

					if resp != nil {
						// defer Close of Body if response from server
						defer resp.Body.Close()
					}

					// Could not connect to any server
					if (err != nil || resp == nil) && key == urlsize-1 {
						// Only discards data if no servers are responding the request
						log.Println("Error connecting khronus server")
						log.Println(err)
						rrindex++
						continue
					}

					if resp == nil {
						log.Printf("Error connecting khronus url: %s\n", url)
					} else {
						body, err := ioutil.ReadAll(resp.Body)
						if resp.StatusCode == 200 {
							// Be are ok, so let's continue
							break
						} else {
							// For anything else ...
							log.Printf("Error connecting khronus url: %s\n", url)
							log.Printf("Http Status code : %d\n", resp.StatusCode)
							if err != nil {
								log.Println("Error reading error code from server")
							} else {
								log.Printf("Http Response : %s\n", bytes.NewBuffer(body))
							}
						}
					}
				}
			}
		}
	}
}
