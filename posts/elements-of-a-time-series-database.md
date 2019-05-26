---
title: Elements of a time series database
slug: elements-of-a-time-series-database
author: mark
published: 2019-05-26T00:00:00Z
tags:
  - programming
  - database
---

# Elements of a time series database

This post is an attempt to deconstruct time series database to its components.
While preparing the docs for something similar at work, I felt that having an intro article to
time series database would be helpful. Hence this.

> P.S: Some familiarity with any SQL database (e.g: PostgreSQL) and analytics,
> though not strictly necessary, will be helpful.

## Data model

The basic elements of time series database is quite simple, consisting of:

- Data points: tuples of `(timestamp, value)`
- Time series: a collection/set of data points with unique timestamps within a `[start, end]` range.
  Usually sorted by ascending order of timestamps.
- Primary identifiers: properties that uniquely identifier time series (primary keys), including:
  - Metric name: generic category, usually refers to the values recorded (prices, temperature, etc)
  - Dimensions: A set of `(key, value)` tuples that further partition the values.
    Usually values with the same key are disjoint while different keys refer to orthorgonal concepts.

As we expands our use cases, more type of secondary data might emerge:

- Aggregates: summary the time series, usually over time or over dimensions.
  They causes loss of resolution, however.
- Secondary identifiers: properties that identifies time series but is not primary identifiers.
  E.g: When recording the temperature of a place over a year, you might refer to disjoint, static
  sections such as `spring`, `summer`, `fall`, `winter`. On the other hand, stock prices might
  have sections like `should-buy`, `should-sale`, which are fuzzy and dynamic.

## What's in a database?

This is oversimplification, but usually you can break down a database like this:

- A storage engine that keeps the data somewhere (memory/on disk).
  - Actual data on disks
  - Indexes
- A layer that parses (read/write) queries, then interact with the storage engine.

Because the most basic form of data storage is continuous chunks (i.e: array, files, etc),
looking for them becomes expensive tasks. Hence there will be indexes used to lookup where
a particular piece of data is. Then we can do pointer arithmetic/seek to the location

> P.S: We purposely left out things that control consistency (session, transaction, etc), as well
> things that computer/disk drivers do such as memory paging.

## Mapping those ideas to a time series database (Prometheus)

So this part really depends on how we implement the database. If we look into Prometheus,
you will roughly see:

- The storage engine: <https://github.com/prometheus/tsdb>
- The query execution engine: <https://github.com/prometheus/prometheus/tree/master/promql>

Let's dig into the respective parts

### The storage engine

Again, this section is oversimplifying things!

This was how the old (2.0) storage engine in Prometheus is designed.

Basically, data is broken into chunks, each for a `[start, end]` range.
This takes advantage of the facts that writes for Prometheus usually happen at time close to
wall clock times (as its primary purpose is monitoring things and therefore periodically scrape
exporters for data).

Time series were identified by a hash of its sorted identifiers.

Data came into memory and is kept there for a while until it's (presumably) "completed" (no more
writes). Then they goes on disk in a predictable pattern.

When we need to query for a `(timeseries, [start, end])`, we run the same hash function to find
where the file is, then use some math to find where the data we want is.

In their new 3.0 storage engine, things have changed a bit, but they are still conceptually
familiar. Data is laid out like this on disk

```
./data
├── 01BKGV7JBM69T2G1BGBGM6KB12
│   └── meta.json
├── 01BKGTZQ1SYQJTR4PB43C8PD98
│   ├── chunks
│   │   └── 000001
│   ├── tombstones
│   ├── index
│   └── meta.json
├── 01BKGTZQ1HHWHV8FBJXW1Y3W0K
│   └── meta.json
├── 01BKGV7JC0RY8A6MACW02A2PJD
│   ├── chunks
│   │   └── 000001
│   ├── tombstones
│   ├── index
│   └── meta.json
└── wal
    ├── 00000002
    └── checkpoint.000001
```

- There is a write-ahead-log (WAL) that helps recovering from crashes.
- Data is split into 2-hour blocks (the top-level folder).
  - Each block have multiple chunk files, each contain data for all time series in a range
  - There is a files for the index and metadata.
  - There are also tombstones, used to denote deleted data (to remove during compaction).

### The query engine

Usually, the purpose of query engine is to:

- Parse a string into a data structure representing a query.
- Convert this query into an execution plan.
- Follow this plan, fetching data from storage engine when necessary.
- Perform any other computations as required in the query.

Our PromQL query engines also follows a similar pattern, with some additional notes:

- When parsing a query, it takes a data structure from the Prometheus API server, this includes:
  - A string query
  - Other parameters related to the type of query: instant/range

It then parse the string query into an abstract syntax tree (AST).
If you are familiar with compiler/interpreter, you will see two similar steps: lexing and parsing.

It's probably easier to look at their tests:

<pre class="language-go"><code class="language-go">
// Lexer
// https://github.com/prometheus/prometheus/blob/master/promql/lex_test.go#L497

{
	input: `test_name{on!~"bar"}[4m:4s]`,
	expected: []item{
		{ItemIdentifier, 0, `test_name`},
		{ItemLeftBrace, 9, `{`},
		{ItemIdentifier, 10, `on`},
		{ItemNEQRegex, 12, `!~`},
		{ItemString, 14, `"bar"`},
		{ItemRightBrace, 19, `}`},
		{ItemLeftBracket, 20, `[`},
		{ItemDuration, 21, `4m`},
		{ItemColon, 23, `:`},
		{ItemDuration, 24, `4s`},
		{ItemRightBracket, 26, `]`},
	},
},
</code></pre>

<pre class="language-go"><code class="language-go">
// Parser
// https://github.com/prometheus/prometheus/blob/master/promql/parse_test.go#L999

{
	input: `test{a="b"}[5y] OFFSET 3d`,
	expected: &MatrixSelector{
		Name:   "test",
		Offset: 3 * 24 * time.Hour,
		Range:  5 * 365 * 24 * time.Hour,
		LabelMatchers: []*labels.Matcher{
			mustLabelMatcher(labels.MatchEqual, "a", "b"),
			mustLabelMatcher(labels.MatchEqual, string(model.MetricNameLabel), "test"),
		},
	}
}
</code></pre>

Then there's an [engine](https://github.com/prometheus/prometheus/tree/master/promql) that goes
through this AST and executes. This is probably the most convoluted parts. A debugger will be
more helpful to understand what's going on.

## Challenges

There are quite a few interesting challenges when building a time series database.
These are written in details quite well [in this article](https://fabxc.org/tsdb). To summarize:

- Write patterns are vertical: Due to the nature of data to be captured, Prometheus writes to
several time series at one point in time.
- Query patterns can spawn large time ranges and over lots of time series
- Need for effective ways to purge data that no longer need to be retained.

## Conclusion && References

I hope this article gave you a good starting point when thinking of time series database.

As usual, the devil is in the detail, so there are papers/code to read and experiments to run.

- [Gorilla, an in-memory time series database developed at Facebook](http://www.vldb.org/pvldb/vol8/p1816-teller.pdf).  - [Timescale DB](https://docs.timescale.com/v1.3/main), a different architecture built on PostgreSQL.
- [kdb+](https://kx.com/discover/time-series-database/), another time series database. Time-tested (16 years old).
- [InfluxDB](https://docs.influxdata.com/influxdb/v1.7).
- [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics), Prometheus with more [features](https://github.com/VictoriaMetrics/VictoriaMetrics/wiki/ExtendedPromQL).
- [Uber's M3](https://eng.uber.com/m3/).
- [Thanos](https://github.com/improbable-eng/thanos).
- [Veneur](https://github.com/stripe/veneur), for aggregation.

Plus maybe search VLDB journal for more research related to columnar, LDAP and key-value database.
