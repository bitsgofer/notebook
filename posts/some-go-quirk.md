---
title: Some quirk about Golang
slug: some-golang-quirk
author: mark
published: 2018-08-23T21:56:00+08:00
tags:
  - golang
---

This is a post documenting some quirky features of Go that I found suprising
(mostly in the "hey, that's funny" way).

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

I'm not really sure where would this come in handy, though. Any examples?
