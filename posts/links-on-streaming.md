title: Materials about stream processing
author: mark
published: 2019-06-01T00:00:00+08:00
tags:
  - streaming
summary: |
  A collection of links to read about stream processing.

----

While I didn't sign up to work in data analytics (the business intelligence part),
I got into learning about streaming through a different paths: billing and monitoring.

It was somewhere in 2016 where I started realizing there's something deeper about contennt of database
that changes over time. I was helping to maintain a billing system then, at it had a lot of problems
where we had to guess what the state of the db was in the past to debug.

Fast forward to 2019 and I saw [this paper](https://arxiv.org/abs/1905.12133).
I have been following works by Tyler Akidau for a while, especially about this stream-table duality,
so this looks like something very interesting.

A lot of things clicked together after reading it!

This page is dedicated to collect and curate materials that help develop this understanding about
streaming systems. Thus, articles will be more academic (to understand concepts), rather than the
how-to-run-X type of articles.

******

The list

- [Kafka, Samza and the Unix Philosophy of Distributed Data](https://martin.kleppmann.com/papers/kafka-debull15.pdf)
- [Apache BEAM Tecnical docs](https://drive.google.com/drive/folders/0B-IhJZh9Ab52OFBVZHpsNjc4eXc),
- [Streaming 101](https://www.oreilly.com/ideas/the-world-beyond-batch-streaming-101) and [Streaming 102](https://www.oreilly.com/ideas/the-world-beyond-batch-streaming-102)
- [The Dataflow Model: A Practical Approach to Balancing Correctness, Latency, and Cost in Massive-Scale, Unbounded, Out-of-Order Data Processing](https://ai.google/research/pubs/pub43864)
- [One SQL to Rule Them All](https://arxiv.org/pdf/1905.12133)
