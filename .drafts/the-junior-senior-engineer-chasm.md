---
title: The junior-senior engineer chasm
slug: the-junior-senior-engineer-chasm
author: mark
published: 2019-01-28T10:54:00+07:00
tags:
  - thoughts
  - engineering
---

There are many type of chasm: the physical space between rocks, the unknown time (for startups)
between having early adopters and realing becoming sustainable.
For engineers (at least me - a sample size of one), there seems to be another type of chasm:
the gap between being a junior and engineer and a truly "senior" one.

After talking with a lot of people, visiting meta sites like [CSCQ](https://www.reddit.com/r/cscareerquestions)
and a lot of self-reflection, I ended up believing that this chasm is real.
For better or worse, I am inside it. Still, I have hope that the chasm can be mapped and overcame.

For me, the junior-senior chasm comes in these different dimensions:

## Technology changes, but we are still iterating on the same few problems:

This was the an lesson learnt early for me. After years of chasing after new language and frameworks,
it becomes apparent that sometimes we are just building incrementail changes to problems/tools that
we face daily as programmers. Some (overly simplified) examples are:

- It's easier and more efficient to let computers to construct proof of your programs than
doing it yourself
  => This correlates to efforts in strong typing, testing, formal specification and verification, etc
- Many things is recognizing (problematic) patterns and applying pre-defined solutions
  => Design patterns recognize groups of tasks; analtyics and monitoring recognize irregular changes
  in products/server; devops identify similar tasks to be automated; agile helps dealing with the
  risks for building for the unknonw/never-complete specifications, etc.
- A lot of our work is built on the basis of trust in blackboxes
  => Programmers actually trust far more than those we give credits too: APIs; our compilers/interpreters;
  CAs; computational/physical limits. You can argue that things are open-sourced, but how many people
  actually read and understand JSON spec regarding number fully, let alone Linux kernels?

That is not to say you don't need to learn because the problems are the same, however.
In fact, while most chefs can cook omelettes, not many can make a (relatively) perfect one
(do google this, it's quite interesting). To do so would require a lot of mastery and attention
to the tools at their disposal: heat, prior experience (with egg, heat, salt and butter),
techniques, etc.

Similarly, while the problems programmers deals with are the same at large, we don't always have
good mastery over our tools. It doesn't help that our toolkit is also quite enormous:
IDEs/editor; debugging tools; monitoring tool; programming languages, databases, kernel, etc.

Here, I am taking a different stand from the usually touted "programming is about problem solving".
As a practitioner, I can safely say that knowing how to solve problems (with ideas) only get you
so far. We still have to sit down and write code, or read through some debug log and understand
wtf is going on.

In the end, what matters is recognizing the core problems, understand them and spending effort
in updating yourself with each new iteration of the solutions.

## Work is about delivering value, technology is a means to an end

This took a while to sink in for me. Yet it's quite simple, actually.
The poignant, capitalistic and over-exaggerated version is:

- You need money to live (like many of us).
- The company need money to pay you.
- The customer need to see something they value to give the company money.

What complicates things are:

- What peope perceive as value changes over time (sometimes unpredictably).
- People can be "played" into valuing one thing more than another.
- There are always new player (more hungry people, other people with automations) to enter the rat race.

Sometimes it feels unfair that people get luckly and find new valuable things to offer.
Other times they are smart at automating the sh\*t out of their work to deliver more value than you.
Nontheless, you are still in this race for delivering value, period.

P.S: on the less-sad side, if you are smart about it, you can probably find ways to deliver value
that also suite your lifestyle.

Realizing this focus on delivering is the next important things for a senior, IOM.
This actually creates **constraints** to what you do.
However, these constraint also help you prune many paths on the decision trees to be traversed.
Again, there are always nuances to this, but if you prune correctly most of the time, your time
toiling around, searching for a good-enough solution is shortened.

## Realizing your limits

This one is the hardest, and one that I am still learning.

I think ultimately we all have some ceiling to performance.
However, if doing something feels hard and uncomfortable, it might not necessarily be outside
of this limit. It is an opportunity to test and see if you can push your boundary.

However, there are things you can't do this way, still. A crude example is that if you type at
~110-120 wpm, you can only chunk out roughly 172k words/24 hour. You can't push much beyond of that.
To have more words, you need more people. And while programming is not the same as typing words,
certain ways of doing things have their limits, too.

Finding and overcoming these pseduo limits, I feel, is a very important factor that separates
a junior from a senior (the 10x guy, so to speak).

IMO, there are 2 sides of things to improve here:

The first order of business is to learn using tools better and to use better tools.
Using tools better is about builidng depth, so you are intimately familiar with programming languages,
OS, network, APIs, etc. Good understanding, when paired with experimental experiences develops
heuristics in your brain (not unlike how deep-learning work, perhaps) that helps you prune paths
in an ever-growing tree of possibilities. The increased speed in reaching some (local) optima is
always welcome.

The second type of things to improve is similar to concurrency. To take on bigger projects, you
must be able to break it down to many more smaller tasks to be done independently, yet
can be interacting well with each other later. These are created (usually with team effort)
with the intention of farming them to more engineers to execute in parallel. Much of the work
will be at defining the spec and interactions, as well as transmitting it to other engineers
and help them execute it at an optimal pace. (It's hard not to draw similarities here, even though
there are a lot of human-related effort to be put as well).

For me, the 2nd point is something with lots of room to improve. Hopefully it will get better
with effort and practice.

## Closing thoughts

Let's see how things will go 3-5 years down the line from here. It would be great if some of these
ideas are the right heuristic :)
