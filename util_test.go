package main

import "testing"

func TestMergeMap(t *testing.T) {
	a := Resource{
		"foo":    "bar",
		"nested": Resource{"foo": "bar"},
	}
	b := Resource{
		"baz":    "bax",
		"nested": Resource{"baz": "bax"},
	}
	r1 := MergeMaps(a, b)
	if r1["foo"] != "bar" || r1["baz"] != "bax" || r1["nested"].(Resource)["foo"] != "bar" || r1["nested"].(Resource)["baz"] != "bax" {
		t.Errorf("failed to merge a+b")
	}
}
