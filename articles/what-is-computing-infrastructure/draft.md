What is "computing infrastructure", exactly?
============================================

> This blog post talks about my exploration of what is considered "computing
> infrastructure", from a view point of someone who worked in more traditional
> Software Engineering role.
>
> It was mainly written to help myself with career planning.
> My hope is that it will be useful for you, if you are more interested in
> working with computers and systems, rather than building features
> at SaaS companies.

What led me to explore computing infrastructure
-----------------------------------------------

For my first and second job, I worked as a software engineer building products
(video on-demand and IoT analytics). However, due to the team size,
I ended up touching things like databases, message queues, CI/CD and monitoring.
For better or worse, I ended up liking those work more than making new features.
It is still good to see people use what you built, but sometimes I can't make
a connection between what I do and the "joy" that someone felt watching
movies or playing games, even if I occasionally enjoy those myself.

Working with computers, on the other hand, yield stimulating tasks of
making the complex systems of silicon, networks, OS/system programs, etc
work together, ultimately bringing a program to life.
Computers are also very concrete: they convert electricity, cooling,
leased bandwidth into useful work that can be billed. Plus, software at this
level doesn't always do everything to chase after profit/product-market fit.
They still retain respects for the "craft", demand rigor like an engineering
discipline, yet exude some curiosity of academic research.
This feel pretty refreshing, after spending a long time in the
startup/corporate world.

In my third gig, I was lucky enough to join an "infrastructure" team in a
game publisher's studio. Attempting to run Kubernetes in a data center
(and later on in GCP) presented new perspectives on what really powers
most tech companies. The feelings I had when learning about lower-level systems
(e.g: nostalgia, excitement, poignancy, etc) really hit me as well.
Thus, as I keep digging new rabbit holes to jump down, I knew it would be hard
to go back...

Yet life doesn't always allow us to freely chase after passion. There is still
aspects about financial, job prospect, etc to work out. It was during one
of these planning sessions that I realized I only know a tiny little corner
of what is considered "computing infrastructure". And so after some panicking
and feeling like an impostor, I attempted to draw my own map for this
uncharted water.

The map
=======

Drawing the boundaries
----------------------

- Q: What is **not** computing infrastructure?
  - still cover a wide range, but most likely not include SaaS software,
  - a bit above the silicon, not dealing directly with transistors and circuits,
    but have a healthy respect for the hardware
  - usually not about embedded systems for specific devices (e.g: health tracker,
    point-of-sale devices, factory assembly lines, etc), but closely-related
  - usually less about mobile apps (mobile devices is a different).
  - definitely include things like: the Internet, OS, public cloud, data center
  - is related to personal computing devices (PC, laptop, mobile) and their OS
  - maybe include a bit of IoT
  => in conclusion, quite a large, fuzzy area

Where is a good entry point
---------------------------

- Strong foundation in OS, computer architecture, networking.
- Data Structure/Algorithm is useful knowledge, but IMO, mostly is meant to help with efficiency
- Understanding of how upper-layer things is important, too (e.g: web services,
  machine learning, graphics workloads)

What are the limiting factors boundaries
----------------------------------------

- Law of (quantum?) physics
- Speed of light
- Quantum computing?
- Economics of computing infra


History
=======

Companies that are influential
------------------------------

- Sun Microsystems
- Google, Amazon, Microsoft (cloud)
- Red Hat, Canonical (open-sourced / OS)

Subcultures
-----------

- Home-lab
- Kernel developers / Linux Tovarlds
- UNIX / Linux / OpenBSD
- Internet sysadmins

Research
--------

Career
======

What are the career paths here
------------------------------

What qualities is advantageous
------------------------------

Other resources
===============
- <https://systemswe.love/>
