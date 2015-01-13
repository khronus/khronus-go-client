package khronusgoapi

import "testing"

func TestRecord(t *testing.T) {
	m := Counter("test")
	m.Record(1, 2, 3, 4, 5)
	m.Record(6, 7, 8, 9, 10)

	if len(m.Measurements) != 2 {
		t.Fail()
	}

	for k, v := range m.Measurements[0].Values {
		if v != uint64(k+1) {
			t.Fail()
		}
	}
	for k, v := range m.Measurements[1].Values {
		if v != uint64(k+6) {
			t.Fail()
		}
	}
}

func TestRecordWithTs(t *testing.T) {
	m := Counter("test")
	m.RecordWithTs(11111, 1, 2, 3, 4, 5)

	if len(m.Measurements) != 1 {
		t.Fail()
	}

	for k, v := range m.Measurements[0].Values {
		if v != uint64(k+1) {
			t.Fail()
		}
	}

	if m.Measurements[0].Timestamp != 11111 {
		t.Fail()
	}
}

func TestAppend(t *testing.T) {

	m1 := Counter("test")
	m2 := Counter("test")

	m1.Record(1, 2)
	m2.Record(3, 4)

	m1.Append(m2)

	if len(m1.Measurements) != 2 {
		t.Fail()
	}

	for k, v := range m1.Measurements[0].Values {
		if v != uint64(k+1) {
			t.Fail()
		}
	}
	for k, v := range m1.Measurements[1].Values {
		if v != uint64(k+3) {
			t.Fail()
		}
	}

}
