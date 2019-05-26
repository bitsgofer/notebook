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
I will use Prometheus as an example.

> P.S: Some familiarity with SQL and analytics, though not strictly necessary, will be helpful.

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
  have sections like `should-buy`, `should-sale`, which are fuzzy and not dynamic.
