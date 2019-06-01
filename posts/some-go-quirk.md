---
title: Some quirk about Golang
slug: some-golang-quirk
author: mark
published: 2018-08-23T21:56:00+08:00
tags:
  - golang
---

Some quirky features of Go that I found suprising.

## 1. Declare arrays/slices the associative way

I have seen this before, but the idea that array/slices can be declared associatively (to the index)
goes against what was ingrained in me. Anw, this is how it look like

<pre class="language-go"><code class="language-go">
var arr = [5]int {
	2: 3,
	1: 4,
    3: 2
}
fmt.Printf("%#v\n", arr) // [5]int{0, 4, 3, 2, 0}



var slice = []int {
	2: 3,
	1: 4,
    3: 2
}
fmt.Printf("%#v\n", slice) // []int{0, 4, 3, 2}
</code></pre>

That surely looked like a `map`, but it isn't.

The properties of this are:

- All elements of the (backing) array is initialized (to value associated with an index,
  or to the zero-value of the type).
- If it's a slice, the backing array contains up to the largest index (`len == max(index) + 1`).
- No ordering of index is required in this form.

I'm not really sure where would this come in handy, though.

## 2. Terminating testing.T

This one is best [illustrated in this example](https://github.com/bitsgofer/gowat/tree/master/channel-in-test).

Basically, the docs for [testing.T](https://golang.org/pkg/testing/#T) says:

	A test ends when its Test function returns or calls any of the methods FailNow, Fatal,
	Fatalf, SkipNow, Skip, or Skipf. Those methods, as well as the Parallel method,
	must be called only from the goroutine running the Test function.

So, if you happen to call `Fatalf` in a goroutine, which I did, the test will not stop.
You will then have to deal with problems due to it not stopping (a deadlock) in my case.
