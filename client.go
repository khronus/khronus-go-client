package khronusgoclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	buffer map[string]*Metric
	mu     sync.RWMutex
	client *http.Client
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
	for k := range c.buffer {
		//c.buffer[k].Measurements = []Measure{}
		delete(c.buffer, k)
	}
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
	return fmt.Sprintf("{\"metrics\":%s}\n", string(js)), nil
}

func (c *Client) sender() {
	rrindex := 0 // simple round-robin index

	c.config.mu.Lock()
	urlsize := len(c.config.urls)
	tick := time.Tick(time.Duration(c.config.interval) * time.Second)
	c.config.mu.Unlock()

	c.client = &http.Client{}
	c.client.Timeout = time.Duration(5) * time.Second

	for {
		select {
		case m := <-c.config.channel:
			{
				c.addMetric(m)
			}
		case <-tick:
			{
				// [FIXME]
				c.config.mu.Lock()
				urlsize = len(c.config.urls)
				// tick = time.Tick(time.Duration(c.config.interval) * time.Second)
				c.config.mu.Unlock()

				if len(c.buffer) == 0 {
					continue
				}

				c.mu.Lock()
				js, err := c.toJson()
				c.emptyBuffer()
				c.mu.Unlock()

				if err != nil {
					fmt.Println("Error marshaling json DataPoints", err)
					// Something really wrong is happening here
					break
				}

				for _ = range c.config.urls {
					if rrindex > urlsize-1 {
						rrindex = 0
					}

					c.mu.Lock()
					url := c.config.urls[rrindex]
					c.mu.Unlock()

					err = c.postData(url, &js)

					rrindex++

					if err != nil {
						fmt.Println("Error connecting khronus server : ", url)
						fmt.Println(err)
					} else {
						break
					}
				}
			}
		}
	}
}

func (c *Client) postData(url string, js *string) error {

	buff := bytes.NewBufferString(*js)

	resp, err := c.client.Post(url, "application/json", buff)

	if err != nil {
		return err
		// Only discards data if no servers are responding the request
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println("Error reading error body")
			return err
		}

		if resp.StatusCode == 200 {
			// Be are ok, so let's continue
			return nil
		} else {
			// Could not connect to any server
			// For anything else ...
			fmt.Printf("Error connecting khronus url: %s\n", url)
			fmt.Printf("Http Status code : %d\n", resp.StatusCode)
			fmt.Printf("Http Response : %s\n", bytes.NewBuffer(body))
			return errors.New(string(body))
		}
	}

}
