---
title: Summary and commentary on talks at Gophercon Singapore 2019
slug: summary-and-commentary-gophercon-singapore-2019
author: mark
published: 2019-05-03T00:00:00Z
tags:
  - programming, conference
---

> Disclaimer: This is a highly subjective summary and commentary on talks at Gophercon Singapore 2019.
>
> It is mainly for me to keep distil ideas from the conference. I took liberty to paraphrase many
> things, so if you want the original talk, wait for official videos. Please expect that YMMV, too.
>
> UPDATE: Rephrased a few parts. I write most of this from memory now, so there might be things that
> don't match with what happened >.<

# Clear is better than clever

[Dave Cheney](https://2019.gophercon.sg/speakers/#dave-cheney) talked about clarity, how it
is different from readability and why it is important. If you have seen his notes/book for
the Practical Go workshop at [Gophercon China](https://dave.cheney.net/practical-go/presentations/qcon-china.html) or
[Gophercon Singapore 2019](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html),
it will feel very similar.

Programmers with more "war stories" will appreciate it more, I think.

The talk start with a quote from last year's keynote:

> Software engineering is what happens to programming when you add time and other programmers.
>
> --- [Russ Cox, referencing: "Go at Google: Language Design in the Service of Software Engineering"](https://research.swtch.com/vgo-eng)

To me, the concept of clarity proposed is essentially "easy for new maintainers to pickup".
This is subjective though. Personally, this probably means you can get a fairly experienced programmer
who is familiar with the project's languages, show them the code and not getting a lot of WTFs.

The first half of the talk re-emphasize that "code is read more often than it is written"
(Guido van Rossum) and that "simplicity is prerequisite for reliability" (Edsger W. Dijkstra).
In the later half, Dave gave [guidelines on naming, declaring variables, indenting flow and
designing public APIs](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html),
to make it clear what the code is doing.

The talk ends with a reality of software project: people change jobs and projects all the time, so
when you write code, try to help the next person.

## Other tidbits

I picked up a few additional things from Dave's workshop, too.

#### 1. [variable shadowing](https://en.wikipedia.org/wiki/Variable_shadowing)

When we see the same variable name appear in different scope and it mess with our head.

#### 2. books

- [The Limits of Software: People, project and Perspectives](https://www.goodreads.com/book/show/3369746-the-limits-of-software)
- [Principles of Operating Systems](https://mitpress.mit.edu/books/principles-operating-systems)

#### 3. Sometimes, what we really want out of a type is (a subset of) its [method set](https://github.com/golang/go/wiki/MethodSets#the-spec).

For example, we write:

```go
func (thing *MyStruct) SaveTo(f *os.File) error {
	// serialize thing -> bytes, example:
	b, err := json.Marshal(thing)
	if err != nil {
		// ...
	}

	// write bytes -> open file, example:
	n, err := f.Write(b)
	// check n and err

	// close file (and persist the content)
	if err := f.Close(); err != nil {
		// ...
	}
}
```

`*os.File` have a lot of methods we don't need like `.Fd()`, `.Stat()`. Hence, we can use an interface/type
that defines only `.Write()` and `.Close()`, which happens to be `io.WriteCloser`, thus we can write.

```go
func (thing *MyStruct) SaveTo(w io.Writer) error {
	// ...

}
```

This have some other nice effects:

- The new type is stricter. It's not possible for funny things like calling `.Truncate()`.
- It's easier to write **some** tests as well, since you can use a `*bytes.Buffer` instead of opening
  a file. (NOTE: You still need an `*os.File` if you want to check the "writing to a file" side-effect.
  However, other tests that don't need it can use something else).

#### 4. Variadic function params that requires at least one value

[This part was from the workshop](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html#_prefer_var_args_to_t_parameters).


```go
func check(first int, remaining ...int) {
	// ...
}
```

This will only compile when there is at least one elements. It also have a small cost where you need
to merge `first` and `remaining` to be a slice again (sometimes only).

It's a shame I only notice this now, despite all the `x:xs` pattern in Haskell :(

#### 5. Gofmt and social contract

One thing that Go get right is `gofmt`, as it became an enforced social contract for most
Go programmers, before any stubborn "style camp" appear.

******

# High(er) Reliability Software Patterns for Go

[Junade Ali](https://2019.gophercon.sg/speakers/#junade-ali) talked about contracts and showed
examples of how to write them in Go.

When writing critical software, we will need them and more (e.g. contracts w.r.t timing, halting,
correctness through concurrent execution, resource usage, etc).

P.S: You shouldn't use Go for critical software, anyway.

Another problematic area is performance (checking contracts cost CPU cycles) and graceful failure
(what to do when contracts fails), which are things we also need to think of.

## Other tidbits

#### 1. Design by contracts vs Defensive programming

This will help with the confusion: <https://softwareengineering.stackexchange.com/questions/125399/differences-between-design-by-contract-and-defensive-programming>

#### 2. Critical failures from software

- [Toyota Camry's unintended acceleration](https://users.ece.cmu.edu/~koopman/pubs/koopman14_toyota_ua_slides.pdf)
- [Quantas Flight 72's uncommanded pitch-down](https://en.wikipedia.org/wiki/Qantas_Flight_72)

#### 3. Language with contracts

- [Ada SPARK](https://en.wikipedia.org/wiki/SPARK_%28programming_language%29)
- [Eiffel](https://en.wikipedia.org/wiki/Eiffel_%28programming_language%29)


#### 4. Other readings

- [Design by Contract - Bertrand Meyer](http://se.inf.ethz.ch/~meyer/publications/computer/contract.pdf)

******

# Understanding Allocations: The stack and the heap

[Jacob Walker](https://2019.gophercon.sg/speakers/#jacob-walker) talked about heap and stack in Go.

Basically, most of the time you shouldn't need to know about whether variables are allocated in
heap or stack. The Go compiler should make the correct choice
(not necessary the most performance-friendly one, however).

Some rule of thumbs:

- Sharing down (pointer used further down in the function/variables passed into next function call)
  typically stays in the stack.
- Sharing up (pointer returned to call functions) typically gets allocated from heap.
- Wrapper structs (slice, map, channel) have special semantics, since they don't contain the actual values.
- More info: <https://golang.org/doc/faq#stack_or_heap>

## Other tidbits

#### 1. gcflags

`go tool compile -h` gives the options to pass in `-gcflags` (go compile flags?).

I have been looking for this since forever, yay!

#### 2. Explains the design of `io.Reader` API

```
Read(p []byte) (n int, err error)
```

It make sense now, since we usually call `Read()` in tight loops.

If the API is `Read() ([]byte, error)`, slices will escape to heap and there will be a lot of things to
`malloc` + GC => horrible performance.

******

# Going secure with Go

[Natalie Pistunovich](https://2019.gophercon.sg/speakers/#natalie-pistunovich) talked about
some guidelines on writing secure application in Go.

Many tools/projects for Go was introduced, though some basic concepts from
[OWASP](https://www.owasp.org/images/7/72/OWASP_Top_10-2017_%28en%29.pdf.pdf) will help establishing
better context.

Some bits that I picked up:

- Environment variable is not always safe (compromised host can show /proc/pid/environ).
  This is a good counter example for "env var is more secure than config file && CLI arguments"!
- Uses [gosec](https://github.com/securego/gosec) to analyze code.
- Uses [depguard](https://github.com/OpenPeeDeeP/depguard) to check dependencies against a verified list.
- pprof HTTP server should not be expose to public
- Uses splunk, sumologic, enterprise Elasticsearch to hide user-identifiable info from dev/ops.
- Know your transitive dependecies (a.k.a, for me, this means be aware of the whole vendor tree)
- Uses [dependabot](https://dependabot.com/) for automated PR when popular project gets updated.
- Having a central dependency repo is risky (hello, npm?).
- [Kritis](https://github.com/grafeas/kritis)

## Other tidbits

I think <https://ma.ttias.be/i-forgot-how-to-manage-a-server/> and <https://ma.ttias.be/automating-unknown/>
will be a good complementary perspective, since the talk suggested automating all these workflow.

******

# Using and Writing Go Analyses

[Michael Matloob](https://2019.gophercon.sg/speakers/#michael-matloob) talked about [analysis](https://godoc.org/golang.org/x/tools/go/analysis).

This seems most helpful for writing tools that deal with the source's AST.

It might be worth some investment if we want to build linters that enforce company-wide practices.

******

# Deep learning in Go

[Karthic Rao](https://2019.gophercon.sg/speakers/#karthic-rao) talks about Go and deep learning.

Long talk, but the basic take-away seems to be: don't do it with Go, uses Python or Java instead.

******

# Garbage Collection Semantics

[Bill Kennedy](https://2019.gophercon.sg/speakers/#bill-kennedy) talks about the (simplified)
behavior of the Go GC and how we can write code to help it.

Mark phrase:

- setup + turn on write barrier
- some assisted marking from goroutines that allocates the most
- stop the world
- turn off write barrier + clean up

Sweep phrase:

- sweep: go through things in heap that was marked for GC (concurrent to other goroutines)

General advice is try to allocate (to heap) less -> less work for the GC.

## other tidbits

#### 1. Nice GC visualization

- At <https://spin.atomicobject.com/2014/09/03/visualizing-garbage-collection-algorithms/>

#### 2. Tracing the GC

Run with `GODEBUG=gctrace=1`

#### 3. Personal thoughts

Writing GC-aware code will be a struggle between trying to reduce allocation on heap vs preserving
clarity.

We ought to do this responsibly (only when it's actually needed, and after benchmarking first).

******

# Go, pls stop breaking my editor

[Rebecca Stambler](https://2019.gophercon.sg/speakers/#rebecca-stambler) laid out the problems
that lead to broken open-source Go tools after each releases (tight coupling with the build tool).

She then introduced [the packages tool](https://godoc.org/golang.org/x/tools/go/packages)
and LSP ([Language Server Protocol](https://microsoft.github.io/language-server-protoco])) for Go,
which will make working with editors easier.

******

# Optimizing Go code without a blindfold

[Daniel Martí](https://2019.gophercon.sg/speakers/#daniel-marti) talked about statistic and benchmarking.

He explained statistical approaches that will help with benchmarking (more samples, take note of p-values, etc).
It's universal and can be used for many types of benchmark.

Useful tools:

- [perflock](https://github.com/aclements/perflock):
  throttle workload by default -> easier to benchmark without getting CPU hot (which affects performance).
- [benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat) for auto statistics.
- [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) to compare micro benchmarks.

## other tidbits

#### 1. Entry points to understanding the compiler

- <https://github.com/golang/go/blob/master/src/cmd/compile/README.md>
- <https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/README.md>

#### 2. Existing optimizations

The compiler already optimize some patterns, so we don't have to think too much about
performance and can focus on correctness:


```go
// m map[string]string

for k, _ := range m {
	delete(m, k)
}
```

This no longer iterate over all keys and will just clear the map.

```go
len([]rune(str))
```

This no longer allocate/convert type and will just give you the length.

#### 3. A way to get more insight

```
GOSSAFUNC=<...> go build
```

******

# Controlling Distributed Energy Resources with Edge Computing and Go

[Sau Sheong](https://2019.gophercon.sg/speakers/#sausheong-chang) and [Rully](https://2019.gophercon.sg/speakers/#rullyadrian-santosa)
talked about projects at my company.

******

# Data Journey with Golang, GraphQL and Microservices

[Imre Nagi](https://2019.gophercon.sg/speakers/#imre-nagi) showed examples of GraphQL at his company,
plus their their application architecture in GCP.

******

# Writing Microservice Integration Tests in Go (Finally)

[Michael Farinacci](https://2019.gophercon.sg/speakers/#michael-farinacci) talked about writing
mocks that uses channels in the place of making calls over the network.

Personally, I have a bit of doubt about the practices shown, as it feels more complex compared
to using the right interface/API. Plus it didn't address the "integration" part that I care about
(the behavior/correctness of whole system).

I would suggest <https://www.hillelwayne.com/post/a-bunch-of-tests/> as complementary reading material.

******

# GOing to Sydney

[Katie Fry](https://2019.gophercon.sg/speakers/#katie-fry) talked about how to present yourself
better during interviews.

******

# Engineering Luck: Essential and Accidental Complexity at GOJEK

[Sidu Ponnappa](https://2019.gophercon.sg/speakers/#sidu-ponnappa) talked about lessons learnt
when scaling GOJEK as a company (most engineering-management related topics).

Most of the idea comes from [No Silver Bullet](https://en.wikipedia.org/wiki/No_Silver_Bullet),
[The Mythical Man-Month](https://en.wikipedia.org/wiki/The_Mythical_Man-Month)
and [Out of the Tar Pit](http://curtclifton.net/papers/MoseleyMarks06a.pdf). They are
presented in the context of GOJEK though, so more stories for us.
