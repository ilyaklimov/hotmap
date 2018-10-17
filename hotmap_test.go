package hotmap

import (
	"testing"
	"sync"
	"time"
)

type itemTest struct {
	key string
	value string
}

var itemTests = []itemTest{
	itemTest{"the", "editor"},
	itemTest{"and", "poet"},
	itemTest{"were", "not"},
	itemTest{"so", "much"},
	itemTest{"surprised", "by"},
	itemTest{"fact", "that"},
	itemTest{"cigarette", "case"},
	itemTest{"actually", "contained"},
	itemTest{"Our", "Brand"},
	itemTest{"as", "itself"},
}

func TestHotmap_Set(t *testing.T) {
	hm := New()
	defer hm.Close()

	hm.Set("", "")
	v, ok := hm.Get("")
	if ok != true {
		t.Errorf("empty value not found")
	}

	hm.Set("hello", "world")
	v, ok = hm.Get("hello")
	if v != "world" {
		t.Errorf("value not found")
	}

	hm.Set("hello", "world")
	hm.Set("hello", "multiverse")
	v, ok = hm.Get("hello")
	if !ok || v != "multiverse" {
		t.Errorf("value is not overwritten")
	}


	_, ok = hm.Get("hello")
	if ok {
		t.Errorf("value is not deleted")
	}

	wg := new(sync.WaitGroup)
	ch := make(chan itemTest)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				hm.Set(item.key, item.value)
			}
		}()
	}

	for _, item := range itemTests {
		ch <- item
	}

	close(ch)
	wg.Wait()

	time.Sleep(31 * time.Second)

	if hm.Len() != 0 {
		t.Errorf("auto cleaning don't work")
	}
}

func TestHotmap_Get(t *testing.T) {
	hm := New()
	defer hm.Close()

	for _, item := range itemTests {
		hm.Set(item.key, item.value)
	}

	ch := make(chan itemTest)
	wg := new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				v, ok := hm.Get(item.key)
				if !ok {
					t.Errorf("value not found:\nkey:\t%#v\nvalue:\t%#v\ngot:\t%#v\nwant:\ttrue\n\n", item.key, item.value, ok)
				} else if v != item.value {
					t.Errorf("incorrect value:\nkey:\t%#v\nvalue:\t%#v\nwant:\t%#v\n\n", item.key, item.value, v)
				}
			}
		}()
	}

	for _, item := range itemTests {
		ch <- item
	}

	close(ch)
	wg.Wait()

	if hm.Len() != 0 {
		t.Errorf("values used aren't deleted")
	}

}

func TestHotmap_Delete(t *testing.T) {
	hm := New()
	defer hm.Close()

	for _, item := range itemTests {
		hm.Set(item.key, item.value)
	}
	
	ch := make(chan itemTest)
	wg := new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				hm.Delete(item.key)
				_, ok := hm.Get(item.key)
				if ok {
					t.Errorf("item (key: %s, value: %s) not deleted", item.key, item.value)
				}
			}
		}()
	}

	for _, item := range itemTests {
		ch <- item
	}

	close(ch)
	wg.Wait()
}

func TestHotmap_SetDuration(t *testing.T) {
	hm := New()
	defer hm.Close()
	ds := []int64{0, 1, 5, 10}
	for _, d := range ds {
		hm.SetDuration(time.Duration(d) * time.Second)
		for _, item := range itemTests {
			hm.Set(item.key, item.value)
		}
		time.Sleep(time.Duration(d+1) * time.Second)
		if hm.Len() != 0 {
			t.Errorf("auto cleaning don't work: duration %ds", d)
		}
	}
}

func TestHotmap_Close(t *testing.T) {
	hm := New()

	wg := new(sync.WaitGroup)
	ch := make(chan itemTest)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				hm.Set(item.key, item.value)
			}
		}()
	}

	for _, item := range itemTests {
		ch <- item
	}

	close(ch)
	wg.Wait()

	hm.Close()

	for _, item := range itemTests {
		if _, ok := hm.Get(item.key); ok {
			t.Errorf("cannot close Hotmap: item (key: %s, value: %s) not deleted", item.key, item.value)
		}
	}


}