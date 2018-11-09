Hot key-value storage: thread-safe concurrent map + auto cleaning.

# Example

```
// Create a new Hotmap
hm := hotmap.New()

// Close Hotmap
defer hm.Close()

// Set key, value
hm.Set("hello", "world")

// Get value by key (after the value is deleted)
v, ok = hm.Get("hello")
if ok {
	fmt.Println(v)
}
```

For more examples have a look at hotmap_test.go



