---
title: Summary and commentary on talks at Gophercon Singapore 2019
slug: summary-and-commentary-gophercon-singapore-2019
author: mark
published: 2019-05-03T00:00:00Z
tags:
  - programming, conference
---

> Disclaimer: This is my subjective summary and commentary on talks at Gophercon Singapore 2019.
>
> It is mainly for me to keep track and distil ideas from the conference.
> Obviously there will be talks where I take more/less notes based on interests/existing knowledge.
> It doesn't necessarily reflects anything about the speaker (I'm much grateful to learn from them :D).
>
> YMMV!

On overall, this year have more talks that kept me thinking than last year :)

# Clear is better than clever

[Dave Cheney](https://2019.gophercon.sg/speakers/#dave-cheney) talked about clarity and how it is different from readability. If you have seen his notes/book for Practical Go workshop at [Gophercon China](https://dave.cheney.net/practical-go/presentations/qcon-china.html) or [Gophercon Singapore 2019](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html), it will feel very similar.

I like this talk as it brings up this topic of engineering software, my pet interest, again.

> Software engineering is what happens to programming when you add time and other programmers.
>
> --- [Russ Cox, referencing: "Go at Google: Language Design in the Service of Software Engineering"](https://research.swtch.com/vgo-eng)

Dave argues that there is a difference between clarity and readability, and that Go programmers should strive for clarity.

Most of the talk go over the concept of "code is read more often than it is written" (Guido van Rossum) and that "simplicity is prerequisite for reliability" (Edsger W. Dijkstra). Basically in response to the typical programmers's complaint: "I don't understand that piece of code".

Dave also gave [some guidelines on naming and declaring variables, indenting flow and designing public APIs](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html) to make it clear what the code is doing.

The talk ends with a reality of software project: they should be written with maintainability in mind since we eventually move to new jobs/projects. (it should not be the only concern, obviously).

### Other tidbits from the talk/workshop

#### 1. [variable shadowing](https://en.wikipedia.org/wiki/Variable_shadowing)

When we see the same variable name appear in different scope and mess with your head.

#### 2. books

- [The Limits of Software: People, project and Perspectives](https://www.goodreads.com/book/show/3369746-the-limits-of-software)
- [Principles of Operating Systems](https://mitpress.mit.edu/books/principles-operating-systems)

#### 3. Sometimes what we really want out of an argument is (a subset of) its [method set](https://github.com/golang/go/wiki/MethodSets#the-spec).

For example, we write:

```go
func (thing *MyStruct) SaveTo(f *os.File) {
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

On the other hand, `*os.File` have a lot of other methods we don't use: `.Fd()`, `.Stat()`, etc. We don't really want to use those.

If what we only want is `.Write()` and `.Close()`, we can substitute the type `*os.File` with `io.WriteCloser`.

This have some nice effects:

- We use only what is needed, thus it's not possible for others to do anything funny like calling `.Truncate()` in the future. (a good thing).
- It's easier to test now: a `*bytes.Buffer` also implements `io.WriteCloser`, so we can write unit tests in memory w/o opening a real file.


gh(er) Reliability Software Patterns for Go#### 4. Make "variadic" function params but requires at least one value

[This part was from the workshop](https://dave.cheney.net/practical-go/presentations/gophercon-singapore-2019.html#_prefer_var_args_to_t_parameters). I should have seen this coming after seeing the `x:xs` thing in Haskell :D

```go
func check(first int, remaining ...int) {
	// ...
}
```

This will only compile when there is at least one element, sometimes at the cost of more work to merge `first` and `remaining` to be a slice again.

#### 5. Gofmt and social contract

One thing that Go get right is `gofmt`, as it has become an enforced social contract for most Go programmers before any "style camp" appear.

******

# High(er) Reliability Software Patterns for Go

[Junade Ali](https://2019.gophercon.sg/speakers/#junade-ali) talked about design by contracts and showed some example of how to do this in Go.

This is what we need when writing critical software. However the implementation in Go leaves a lot to be desired.

Another problem that might be a concern is a performance (more work/function call, especially on
systems with high load).

### Other tidbits

#### 1. critical failures from software

- [Toyota Camry's unintended acceleration](https://users.ece.cmu.edu/~koopman/pubs/koopman14_toyota_ua_slides.pdf)
- [Quantas Flight 72's uncommanded pitch-down](https://en.wikipedia.org/wiki/Qantas_Flight_72)

#### 2. Language with contracts

- [Ada SPARK](https://en.wikipedia.org/wiki/SPARK_(programming_language))
- [Eiffel](https://en.wikipedia.org/wiki/Eiffel_(programming_language))


#### 3. other readings

- [Design by Contract - Bertrand Meyer](http://se.inf.ethz.ch/~meyer/publications/computer/contract.pdf)

******

# Understanding Allocations: The stack and the heap

[Jacob Walker](https://2019.gophercon.sg/speakers/#jacob-walker) talked about heap and stack in Go.

Basically, you shouldn't need to know about stack and heap (most of the time). Go compiler should make the correct choice (but not necessary the most performance-friendly one).

Some rule of thumbs:

- Sharing down (pointer used further down in the function/variables passed into next function call) typically stays in the stack
- Sharing up (pointer returned to call functions) typically gets allocated from heap
- https://golang.org/doc/faq#stack_or_heap

### Other tidbits

#### 1. gcflags

`go build -gcflags '-m'`. I have been looking for this for a long time :)

#### 2. Explains why io.Reader API

```
Read(p []byte)
```

It make sense now, since we usually call `Read()` in tight loops.
If we use `Read() ([]byte, int)`, slices will escape to heap and there will be a lot of `malloc` + GC => horrible performance.

******

# Going secure with Go

[Natalie Pistunovich](https://2019.gophercon.sg/speakers/#natalie-pistunovich) talked about some guidelines on writing secure application in Go.

For me this is mostly new tools/projects to look at:

- Environment variable is not always safe (compromised host can show /proc/pid/environ).
- Uses [gosec](https://github.com/securego/gosec) to analyze code.
- Uses [depguard](https://github.com/OpenPeeDeeP/depguard) to check dependecies against a verified list.
- pprof HTTP server should not be expose to public
- Uses splunk, sumologic, entreprise Elasticsearch to hide user-identifiable info from dev/ops.
- Know your transitive dependecies (a.k.a, for me, this means be aware of the whole vendor tree)
- Uses [dependabot](https://dependabot.com/) for automated PR when popular project gets updated.
- Having a central dependecy repo is risky (hello, npm)
- [Kritis](https://github.com/grafeas/kritis)

******

# Using and Writing Go Analyses

[Michael Matloob](https://2019.gophercon.sg/speakers/#michael-matloob) talked about [analysis](https://godoc.org/golang.org/x/tools/go/analysis).

This seems most helpful for writing tools that deal with the source's AST.

It might be worth some investment if we want to build linters that enforce company-wide practices.

******

# Deep learning in Go

[Karthic Rao](https://2019.gophercon.sg/speakers/#karthic-rao) talks about Go and deeplearning.

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
- sweep: go through things in heap that was marked for GC

General advice is basically try to allocate (to heap) less -> less work for the GC.

### other tidbits

#### 1. nice GC visualization

- At <https://spin.atomicobject.com/2014/09/03/visualizing-garbage-collection-algorithms/>

#### 2. tracing the GC

run with `GODEBUG=gctrace=1`

#### 3. personal thoughts

Writing GC-aware code will be a strugle between trying to reduce allocation on heap vs preserving
clarity. Do this responsibly (only when it's actually needed, and benchmark first).

******

# Go, pls stop breaking my editor

[Rebecca Stambler](https://2019.gophercon.sg/speakers/#rebecca-stambler) laid out the problems
that lead to broken open-source Go tools after each releases (tight coupling with the build tool).

She then introduced LSP ([Language Server Protocol](https://microsoft.github.io/language-server-protoco]))
for Go, which will help hooking up with editor easiers (more stable API)>

******

# Optimizing Go code without a blindfold

[Daniel Martí](https://2019.gophercon.sg/speakers/#daniel-marti) talked about statistic and benchmarking.

He explained statisic approaches that will help with benchmarking (more samples, p-values, etc),
which is universal for CPU/memory benchmarking.

Some useful tools:

- [perflock](https://github.com/aclements/perflock): throttle workload by default -> easier to benchmark without getting CPU hot.
- [benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat) for auto statistics
- [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) to compare micro benchmarks.

### other tidbits

#### 1. More gcflags:


```
-m
-d=ssa/check_bce/debug=1 (turn off out-of-bound check for slices)
ssa/prove/debug=2
```

Sometimes you have to give up on understanding performance, as there are too many layers.

The compiler already optimize some correct syntax, so we don't have to think too much about
performance then:


```go
// m map[string]string

for k, _ := range m {
	delete(m, k)
}
```

-> no more iteration over keys, just delete the whole thing -> faster

```go
len([]rune(str))
```

no alloc/conversion -> just give the length

more ways to print debug log:

```
GOSSAFUNC=...

using:
cmd/compile/README
cmd/compile/internal/ssa/README

```

******

# Controlling Distributed Energy Resources with Edge Computing and Go

[Sau Sheong](https://2019.gophercon.sg/speakers/#sausheong-chang) and [Rully](https://2019.gophercon.sg/speakers/#rullyadrian-santosa)
talks about projects at my company.

******

# Data Journey with Golang, GraphQL and Microservices

[Imre Nagi](https://2019.gophercon.sg/speakers/#imre-nagi) shows examples of GraphQL and their
architecutre on GCP.

******

# Writing Microservice Integration Tests in Go (Finally)

[Michael Farinacci](https://2019.gophercon.sg/speakers/#michael-farinacci) talked about writing
mocks that uses channels in the place of making calls over the network.

This might not map 100% to integration tests, however.

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
and [Out of the Tar Pit](http://curtclifton.net/papers/MoseleyMarks06a.pdf), though they are
presented in the context of GOJEK.
