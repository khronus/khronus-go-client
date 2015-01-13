package khronusgoapi

import (
	"bytes"
	"testing"
)

func TestClientConfig(t *testing.T) {

	urls := []string{"http://localhost:8080", "http://localhost:9290"}
	c := Client{}
	c.Config().Urls(urls).Interval(9999)

	for k, v := range c.config.urls {
		if v != urls[k] {
			t.Fail()
		}
	}

	if c.config.interval != 9999 {
		t.Fail()
	}
}

func TestCreation(t *testing.T) {
	urls := []string{"http://localhost:8080", "http://localhost:9290"}
	c := Client{}
	c.Config().Urls(urls).Interval(9999)

	if c.Timer("test").Type != timer {
		t.Fail()
	}

	if c.Gauge("test2").Type != gauge {
		t.Fail()
	}

	if c.Counter("test3").Type != counter {
		t.Fail()
	}
}

func TestEmptyBuffer(t *testing.T) {

	c := Client{}
	c.Config()

	c.Timer("test").Record(1, 2, 3, 4, 5, 6)

	if len(c.buffer) != 1 {
		t.Fail()
	}

	c.emptyBuffer()

	if len(c.buffer) != 0 {
		t.Fail()
	}
}

func TestJsonOneMetric(t *testing.T) {
	c := Client{}
	c.Config()

	resp := `{"metrics":[{"name":"test","measurements":[{"ts":11111,"values":[1234]},{"ts":11111,"values":[456]}],"mtype":"timer"}]}`
	resp2 := `{"metrics":[{"name":"test","measurements":[{"ts":11111,"values":[456]},{"ts":11111,"values":[1234]}],"mtype":"timer"}]}`

	c.Timer("test").RecordWithTs(11111, 1234)
	c.Timer("test").RecordWithTs(11111, 456)
	js, err := c.toJson()

	if err != nil {
		t.Fail()
	}

	if js != resp+"\n" && js != resp2+"\n" {
		t.Logf("\n%q\n%q", bytes.NewBufferString(js), bytes.NewBufferString(resp))
		t.Fail()
	}
}

func TestJsonTwoMetrics(t *testing.T) {
	c := Client{}
	c.Config()

	// Client store metrics in a map which are not ordered

	resp := `{"metrics":[{"name":"test","measurements":[{"ts":11111,"values":[1234]},{"ts":11111,"values":[456]}],"mtype":"timer"},{"name":"test2","measurements":[{"ts":11111,"values":[1234]},{"ts":11111,"values":[456]}],"mtype":"timer"}]}`
	resp2 := `{"metrics":[{"name":"test2","measurements":[{"ts":11111,"values":[1234]},{"ts":11111,"values":[456]}],"mtype":"timer"},{"name":"test","measurements":[{"ts":11111,"values":[1234]},{"ts":11111,"values":[456]}],"mtype":"timer"}]}`

	c.Timer("test").RecordWithTs(11111, 1234)
	c.Timer("test").RecordWithTs(11111, 456)
	c.Timer("test2").RecordWithTs(11111, 1234)
	c.Timer("test2").RecordWithTs(11111, 456)

	js, err := c.toJson()

	if err != nil {
		t.Fail()
	}

	if js != resp+"\n" && js != resp2+"\n" {
		t.Logf("\n%q\n%q", bytes.NewBufferString(js), bytes.NewBufferString(resp))
		t.Fail()
	}
}

func TestWrongUseOfMetrics(t *testing.T) {
	c := Client{}
	c.Config()

	m := c.Timer("test").RecordWithTs(11111, 1234)

	if m == nil {
		t.Fail()
	}

	m = c.Counter("test").RecordWithTs(11111, 1234)

	if m != nil {
		t.Fail()
	}
}
