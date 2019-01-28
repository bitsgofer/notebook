---
title: Bootstrap a Go project: creating build and testing flow
slug: bootstrap-go-project-build-and-test-flow
author: mark
published: 2019-01-16T00:00:00+08:00
tags:
  - go
  - testing
  - ops
---

> This articles describe the process of setting up a workflow for building (using `make`)
and testing Go code (using `go test).
>
> Take it with a grain (or a lot) of salt, because most of the ideas here are opinionated.

# Intro

## Why should you care / why should you do this?

IMO, bootstraping build and testing flow is one of the necessary grunt work
that is not talked about a lot. However, it does have a big impact on the project, because:

- A smooth build process enables new developers to start working on the project faster.
- A good build process also reduces time spent on "why it doesn't work on my laptop?".
- You need something to run tests and provide devs with feedback. The sooner this is done, the better.
- If writing tests is easy, devs write more tests -> more coverage. Hopefully this means better code quality.
- Trade your time (possibly doing tedious/painful) for a better mental model of your project.

## Scope and assumptions

- The project builds Go code (possibly involves cgo) and servers serving gRPC (no frontend).
- We might release libraries (Go packages) or docker images with prebuilt binaries.
- We won't really work from scratch. There might be some code already (a few thousand lines).
- You have a CI machine already setup. We just need to hook this project in.

For starter, let's assume that most devs have these on their laptop:

- Something to run docker: docker for OSX / docker-ce for Linux.
- Terminal emulators.

## Formats of the talks

Let's approach the talks this way:

- First, we explore different choices that we could do at certain stages.
  However, we can only cover a few representative possibilities (there are too many ways).
- Next, We first talk about certain principles that we want to follow. This helps narrowing the choices.
- Finally, we talk about what was actually chosen and what we learned doing so.

Hopefully that give you more insights, instead of just providing information.

# Bootstraping

## 1.Development environment

The first thing we will need is a way to setup computers to build our projects.
The computer can be your developers' machine.

### Why do this?

There is always some setup work to do. We rely much more than just source code to run our projects:

- You need a Go tool chain to compile and run tests
- You might need to compile protobuf/gRPC definition into Go, Ruby client definitions.
- You might need some C libraries (if using cgo).
- You might need to run scaled down versions of your infra (Postgres, Cassandra, Elasticsearch, Kafka, etc)
  to cover your actual interactions (more on testing later).
  - Those infra might have their own pre-requisite as well (e.g: Java)

### What are the choices?

- Provide a flow in README.md. Everyone do your own setup:
  - Things gets ugly fast. Back when I used to build C++ during internship, we have 10 people
    and 10 different combinations of (CPU architecture x OS version x Boost C++ libraries version).
    When it breaks on your machine, you are on your own -> not very productive.
  - Building Go is a lot easier. However you still (at least) have differences in your OS.
- Have some scripts that setup machines for devs. Everyone runs it.
  - There will be updates. Someone will forget to update.
  - You will be doing something stupid on your machine
    (e.g: edit `/etc/hosts`, install some conflicting packages, etc)
  - You might be different from the CI machines, still (at least in how loaded the CPUs are).
- Vagrant, anyone? (or clould VMs, for that matter)
  - Better. Implicit things (how to setup) is now explicit.
  - The problem of this is how do you get people to adopt it.
    Either we get to do this very early in the company culture, or not at all.
- Something like Vagrant, but doesn't involve persuading peoople? -> containers.
  - Because if you are deploying docker containers, chances are high that you also run
    docker daemon locally. And that will be the only dependencies we need.

### How to do it?

- Provide a special "dev environment" container.
- Install everything you need here.
- When the times comes to do things (compile, generate code, unit tests, etc),
  mount the source code into the container and do it.
- NOTE: I try to separate the tools from the things you can generate (protobuf file, SSL keypairs, etc).
  However the line is not always clear (e.g: if you need to check in some protobuf-generated files).

### How did it go?

- The only dependencies is now:
  - The docker daemon you are running
  - Making sure that image is always up-to-date.
- I build the "dev environment" img as pre-requisite before most tasks (from source).
  Because docker layers can be cached, the cost is amortized over time.
  Like most armortized-cost thing, it only works out if you do this a lot ->
  It's still anoying for people who don't work on the code base frequently.
  We could push the layers to the registry to help with this (but I haven't started).
- Mostly add to the image -> more layers, larger image.
  So it wil benefits from periodic cleanup work.
- Added benefit: if you can run containers in CI, you can reuse the work.

## 2.Automate build and test pipeline

After having your environment setup, you can start typing randomly :)
The next hurdle comes in the form of building and testing your code.

### Why do this?

I assume that at the end of the day, you want:

- Your code to compile (even better if it works on the first try :P)
- All the tests to pass (btw, do you know `go test -v ./... | sed -e 's|FAIL|PASS|g'` is also a thing)?

To achieve this, we will need some glue code to do all the boring work, which basically is:

- Generate any necessary files.
- Compile and run tests.
- Compile source code to object files && link them into binaries (`go build` do this for you).
- If possible, provide good build caching so you only need to build what's necessary.
  (Go 1.10+ do this for you).

> This is a very interesting topics.
> You might want to read <http://www.lihaoyi.com/post/SowhatswrongwithSBT.html>

### What are the choices?

- Once again, a `README.md` on how to do it. Not practical for large project.
  However, it might be possible your future self will be able to follow through -> good for small side projects.
- Bash scripts works. However its programing experience is terrible.
- You can go pure CI too (with `Jenkinsfile`/`Groovy` as your glue code.).
  However, you will lose the ability to run things locally (and debug).
  Probably only make sense if your CI systems >>> your local builds.
  Finally, programming in `Jenkinsfile` is still annoying.
- There's an interesting option picking up momentum: Bazel.
  It deserve a dedicated article later. Unfortunately, I won't be able to cover it here.
- Make is a good, practical candidate.
  Even though its programing experience is still lacking, it can do the job.
  There are good things you can take advantage off: parallel commands (`-j`),
  automatic dependency management, etc.

### How to do it?

- Write your `Makefile` :)
  Usually it involves typing the same commands that you would run on the terminals.
- Profit.
- P.S: There's the whole suite of GNU build system (`Autotools`). However, I only know it by name.

### How did it go?

- It works!
- There are still times when I want proper variable, scope and better way to namespace things.
- The "baggage" piles up fast (I probably need to spend more effort learing `Autotools`).
  The most annoying things is tracing the substitution flow and understand what they do.
- I still need to use bash scripts in conjunction with make sometime
  (it is just so damn hard to do certain things).

## 3.Setup your test harness

While we are at this, it might be interesting to read <https://www.hillelwayne.com/post/a-bunch-of-tests/>
for a different categorization of tests. My short take away is that there are 2 main groups:
- Auto-manual tests (manually written, automatically executed)
- Generative tests (automatically written and excuted).

Go have pretty good test harness when it comes to auto-manual tests, espcially those that only
require Go code.
However, you will probably need more at some points, for example:

- Run tests that involve interacting with real infra (not your stubs), so you can check those paths
  and understand what happen IRL.
- Load tests.
- Chaos-monkey kind of tests

Here, I only attempt to cover the auto-manual tests that requires real infra to be present.
The other topis requires much more indepth discussions.

### Why do this?

- Because you need the coverage for the real thing, and you can't do it with stubs.
- Because you don't want to test in production.
- Because you also want to run everything in your stack (infra and code),
  so you can poke at things and see how it reacts.

### What are the choices?

#### Topic 1: How do you run your infra things

- Run some QA infra on the clould and points your local stuff to it.
  I tried this and it ended badly, especially when you depend on data/some other code
  that someone else might be testing.
- Run multiple pieces of infra on the clould.
  I tried this too. It also ended badly (some poor soul gets trapped maintaining the thing).
  (it's under-utilized and cost $$$).
- Let everyone run infra things locally. Connect your code to it.
  Yeah, everyone will have the same things on the same port and everyone is a devops.
  In your dreams!

Containers to the rescue again. This time you need run multiple containers and orchestrates them. So:

- minikube (let's try to copy the whole production setup).
  Now you are really asking everyone to be a devops.
- docker-compose, is a (somewhat) lighter option. It's not that different from minikube functionally.
  However operating is a lot easier (you just need an additional binary, the additionaly syntax part
  is the same).

#### Topic 2: How do you run your tests

Now, you will need to setup your tests with the infra pieces.

There's a few choices, again (not so many, though):

- Compile everything and run. Then (somehow) check that things are running correctly.
  -> Depends on what you do, this part could be hard, since you might have to design your own
     test harness again.
- Use go test again, utilizing build tags.
  This way, you don't need to reinvent another test harness, and can use a fully functionaly
  programming language to write your tests.

### How to do it?

- `//+bulid integration`
- I usually try to prioritize using the package test (`package X_test`),
  so I also check the public API first.

### How did it go?

- It sort of work (more coverage), with some annoyances.
- I keep having to wipe the infra pieces between "integration" tests.
  This doesn't work very well with non ACID stuff. Ended up using `time.Sleep`, which is unpredictable.
- At somepoint, maybe it is better to just use the `t.Name()` to setup separate "spaces" for these
  tests to run.

## Conclusion

- Things sort of work, but could be better.
- The ability to run things locally is useful (debugging on Jenkins sucks).
