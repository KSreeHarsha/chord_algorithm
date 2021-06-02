package triblab_test

import (
	"testing"

	"trib/entries"
	"trib/randaddr"
	"trib/store"
	"trib/tribtest"
	"triblab"
	. "trib"
)

func TestRPC(t *testing.T) {
	addr := randaddr.Local()
	ready := make(chan bool)

	go func() {
		e := entries.ServeBackSingle(addr, store.NewStorage(), ready)
		if e != nil {
			t.Fatal(e)
		}
	}()

	r := <-ready
	if !r {
		t.Fatal("not ready")
	}

	c := triblab.NewClient(addr)

	tribtest.CheckStorage(t, c)

	// kv-client localhost:12086 get hello -> EMPTY_STRING
	var value string;
	e := c.Get("hello", &value);
	if e != nil || len(value) > 0 {
		t.Fatal("get hello should have returned empty string!");
	}

	// kv-client localhost:12086 set h8liu run -> true
	var succ bool;
	v := new(KeyValue);
	v.Key = "h8liu";
	v.Value = "run";
	e = c.Set(v, &succ);
	if e != nil || !succ {
		t.Fatal("set h8liu to run should have returned true!");
	}

	// kv-client localhost:12086 get h8liu -> run
	e = c.Get("h8liu", &value);
	if e != nil || value != "run" {
		t.Fatal("get h8liu must have returned run!");
	}

	// kv-client localhost:12086 keys h8 -> h8liu
	var lst List;
	var p Pattern;
	p.Prefix = "h8";
	e = c.Keys(&p, &lst);
	if e != nil || lst.L[0] != "h8liu" {
		t.Fatal("keys(h8) should have returned h8liu!");
	}

	// kv-client localhost:12086 list-get hello -> nil
	var lst2 List;
	e = c.ListGet("hello2", &lst2);
	if e != nil || cap(lst2.L) > 0 {
		t.Fatal("list(hello2) is expected to return empty list!");
	}

	// kv-client localhost:12086 list-get h8liu -> nil
	var lst3 List;
	e = c.ListGet("h8liu", &lst3);
	if e != nil || cap(lst3.L) > 0 {
		t.Fatal("list(h8liu) should have returned empty list!");
	}

	// kv-client localhost:12086 list-append h8liu something -> true
	v2 := KeyValue{"h8liu", "something"};
	e = c.ListAppend(&v2, &succ);
	if e != nil || !succ {
		t.Fatal("Failed to append something to List(h8liu)");
	}

	// kv-client localhost:12086 list-get h8liu -> something
	var lst4 List;
	e = c.ListGet("h8liu", &lst4);
	if e != nil || cap(lst4.L) != 1 || lst4.L[0] != "something" {
		t.Fatal("List(h8liu) expected to be equal to something!");
	}

	// kv-client localhost:12086 clock 200 -> 200
	var tmp uint64;
	var t200 uint64;
	t200 = 200;
	e = c.Clock(t200, &tmp);
	if e != nil || tmp != t200 {
		t.Fatal("Clock should have been equal to 200!");
	}
}
