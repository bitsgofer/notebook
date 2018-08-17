---
title: Hope is not a strategy
slug: hope-is-not-a-strategy
author: mark
published: 2018-08-17T23:52:00+08:00
tags:
  - thoughts
---

> This is a rant post. It is not directed at any particular person && is more as an observation, if anything.

Just talking about reliability is easy. Most proably, all you need to do is is to take 3 things:

- a word in this set: `{uptime, latency, MMTF, MTTR, etc}`
- a number, preferrably in `[99%, 100%)`
- name of a service

And there you have a bullet point to add to your SLO, easy right...

Well, the reality is way more complicated, and sometimes it seems people forgets that the right SLO is
**a balance** between goals and capability.

- Only looking at goals without consulting people on the ground about capability
gives SLO that is not realistic.
- Only looking at capability (usually as a "belief" from devs and ops)
gives a random value from overly pestimistic to overly confident.

By itself, understanding the true capability of the system you are running is hard enough.
You probably need to:

### 1. Run your app and infrastructure automatically and continously

Note that infrastructure also comes into place, because app doesn't run in isolation.
You might have test and stuff for 1 instance of the app, but running N instances of the app
in M places is a different ball game. Plus, apps need to use the network, to access DBs, etc.

### 2. Measure lots of metrics for your system

This changes for different kind of system: web apps, API servers, data pipelines, storage system, etc...
To top this, you have to balance between getting too many data (useless, verbose noise) vs too little (not enough info).

### 3. Have a f\*\*king good understanding of the what you run

Most complex programs have abstrations. Remember what they tell you about abstraction,
that it's exposes only the relevant details and help decompose problem?
That is stil true, but it benefits the person who builds the stuff
(i.e: good if you use your developer hat).

On the other hand, most abstractions hides details, and details are what you needs as an operator.
(unless you operate formally verified code, that is)

And don't think that going micro-services/single-responsibility component solves all your problem.
It means only means your are problems now distributed (pun intended).

### 4. And all this are recursive

So, all this starts with the app that you write and operate, and then everything it depends on.
You name it: DB, message queue, kubernetes, etcd, file system, DNS, etc. The list go on so long
it probably paints a very bleak future.


### but... there's hope, right?

But hey, I'm using managed services run on GCloud/AWS/Azure, etc, so I'm safer right?

Nope, there is a reason they never tell you things will be working 100% of the time.
There is always a chance for things to fail, and they will (sometimes spetacularly).

Last week logs Azure agent fill my DB server's disk with log. Yesterday Google migrated my VM
and didn't reboot it. And I have no idea what will happen tomorrow.

I think it's almost comical that sometimes we are very vocal about all the reliability strategies,
yet depends on the hope that something, somewhere runs (the routing system for trains you take,
the code that process your bank transfer, the disk that persists bytes that you write to it, etc).
If anything, we should be more cautious on promises, the more you understand about how things really run.

The real battle is to overcome this pestimism, bit by bit with proper engineering...
