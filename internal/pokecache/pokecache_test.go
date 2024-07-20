package pokecache

import (
	"testing"
	"time"
)

func TestReapLoop(t *testing.T) {
	const interval = time.Millisecond * 5
	const waitTime = time.Millisecond*5 + interval
	cache := NewCache(interval)
	cache.Add("test.com", []byte("data"))

	_, ok := cache.Get("test.com")
	if !ok {
		t.Errorf("no key found")
		return
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("test.com")
	if ok {
		t.Errorf("key not reaped")
		return
	}
}

func TestMethods(t *testing.T) {
	const interval = time.Millisecond * 5

	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "test1.com",
			val: []byte("datafortest1"),
		},
		{
			key: "",
			val: []byte("datafortest2"),
		},
		{
			key: "test2.com",
			val: []byte(""),
		},
	}

	for _, testCase := range cases {
		c := NewCache(interval)
		k, v := testCase.key, testCase.val
		c.Add(k, v)
		val, ok := c.Get(k)
		if !ok {
			t.Error("expected to receive")
			continue
		}
		if string(val) != string(v) {
			t.Error("values not equal")
			continue
		}
	}
}
