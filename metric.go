package khronusgoapi

import (
	"errors"
	"time"
)

type Type string

const (
	counter Type = "counter"
	timer   Type = "timer"
	gauge   Type = "gauge"
)

type Measure struct {
	Timestamp int64    `json:"ts"`
	Values    []uint64 `json:"values"`
}

type Metric struct {
	Name         string    `json:"name"`
	Measurements []Measure `json:"measurements"`
	Type         Type      `json:"mtype"`
}

func (m *Metric) RecordWithTs(ts int64, args ...uint64) *Metric {
	if m == nil {
		return nil
	}

	m.Measurements = append(m.Measurements, Measure{Timestamp: ts, Values: args})

	return m
}

func (dst *Metric) Append(src *Metric) error {
	if dst.Name == src.Name && dst.Type == src.Type {
		dst.Measurements = append(dst.Measurements, src.Measurements...)
		return nil
	} else {
		return errors.New("Error appending differents metrics")
	}
}

func (m *Metric) Record(args ...uint64) *Metric {
	return m.RecordWithTs(time.Now().Unix()*1000, args...)
}

func Gauge(name string) *Metric {
	return newMetric(name, gauge)
}

func Counter(name string) *Metric {
	return newMetric(name, counter)
}

func Timer(name string) *Metric {
	return newMetric(name, timer)
}

func newMetric(name string, t Type) *Metric {
	m := &Metric{Name: name, Type: t}
	return m
}
