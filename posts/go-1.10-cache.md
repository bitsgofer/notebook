---
title: Go 1.10 build & test caching
slug: go-1-10-build-and-test-caching
author: mark
published: 2018-04-25T00:00:00Z
tags:
  - go
---

Since I didn't read the release note for 1.10, I "accidentally" discovered a new feature in Go 1.10:
build & test caching.

Basically, the 1.10 release do these things:

- Cache the result of `go build`, `go run` and `go test` in `$HOME/.cache`.
- When they way your program run changes (usually, when source code change or when some flags is used
differently), the cache is rebuilt.
- New builds will take advantage of the cache to only do incremental build.
- Similarly, if the way your test code run haven't changed, `go test` will print the previous test's result.

I learn of this when running a test that is dependent on `postgres`. After I have stopped the PG server,
the test still pass. This resulted in a good hour debugging and checking with a coworker that I am not
imagining things.

The feature doesn't feel very welcoming at first, plus, I was worried about tests that:
- Depends on communication over network
- Use content from a file/env variable.

Then I realized I have been running Go 1.10 for more than a month without any problems. This prompt
for more reading on `go-nuts` and I discovered a great thing: the Go team has taken
a lot of work to ensure that opening files and reading environment variable also invalidates the
test's cache. However, they have also explicity not try to handle all network-related one.

Actually, it is not a new approach (Makefile has taken the incremental build approach, but it can't
cache test result). Furthermore, the underlying problem is **cache invalidation**, and it's hard.

The more I read & think about it, the more it start to make sense. In the end, it started to feel
okay after a while.

An interesting thought is the test cahing seems to be good for unit-like tests, where we can control
all inputs (exhausive). We can surely disable the cache, but if it's design that way, we can take
advantage of most Go's feature.
