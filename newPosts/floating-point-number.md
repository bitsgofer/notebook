title:  Floating point numbers
author: mark
published: 2018-04-10T00:00:00Z
tags:
  - programming
summary: |
  Explaining floating point numbers
----



Recently, one of my colleagues tried to write the value `0.1` into `OpenTSDB` but
got `0.100000001490116` when querying it back.

<pre class="language-bash"><code class="language-bash">
$ echo "run OpenTSDB, listening on :4242"

$ echo "write 0.1"
$ curl -sX POST '127.0.0.1:/4242/api/put' \
        -H 'Content-Type: application/json' \
        -d '[{"metric":"test","tags":{"k":"v"},"timestamp":1523353120,"value":0.1}]' \
        -i
HTTP/1.1 204 No Content

$ echo "query back"
$ curl -sX POST '127.0.0.1:/4242/api/query' \
        -H 'Content-Type: application/json' \
        -d '{"start":1523353100,"end":1523353140,"queries":[{"metric":"test","tags":{"k":"v"},"aggregator":"none"}]}' \
        | json_pp
[
	{
		"aggregateTags": [],
		"metric": "test",
		"tags": {
			"k": "v"
		},
		"dps": {
			"1523353120": 0.100000001490116
		}
	}
]
</code></pre>

The thus a series of WTF commenced.

Eventually, we figured out what happened, which was because `OpenTSDB 2.3.0`, the version we used,
only use 32 bits to store the floating point numbers.

From the process, it seems there are some fundamental facts about floating point that engineers should know.
There are many soures on this, such as:
- [What every computer scientist should know about floating-point arithmetic](https://dl.acm.org/citation.cfm?id=103163)
  <br/>or [alternative link](https://docs.oracle.com/cd/E19957-01/806-3568/ncg_goldberg.html)
- [IEEE 754-2008](https://ieeexplore.ieee.org/document/4610935/)
  <br/>or [alternative link](http://eng.umb.edu/~cuckov/classes/engin341/Reference/IEEE754.pdf)
- <http://floating-point-gui.de>

However, I still feel those sources doesn't explain it clear enough for a confused person to understand.
So this is another stab at it.

> Disclaimer: there's a lot of corner cases with floating point math, this article might miss
> them, still

# TL,DR

1. Most real numbers can't be stored accurately in a computer.
2. They get approximated, usually using a 32-bit or 64-bit representations.
3. Once approximated, you should round/truncate on display and comparison, but NEVER on calculation.
4. If you use either 32-bit or 64-bit, keep it consistent across your system.

******

# What a confused programmers should remember about floating point numbers

## Terms

- `integers`: numbers with no decimal digits (e.g: `0`, `1`, `99999`, etc)
- `real numbers`: numbers in the set ℝ
- `floating point numers`: representation of a `real number` in computer, stored in **N** bits of memory
- `float`, `float32`: `floating point numbers` stored in **32 bits**
- `double`, `float64`: `floating point numbers` stored in **64 bits**

## 1. There is not enough bits to represent real numbers

Recall that when working with `integers`, you also have a limited range
(e.g: `-128 - +127` for `int8`) due to bits limit.
The same problem presents itself with `real numbers`, since you have to cram an infinite amount
of numbers within ℝ into **N bits**.

Unlike `integers` where there is an exactly amount of numbers between 2 `integer` values,
(e.g: only 1 number between `0` and `1`),
there is an infinite amount of numbers between two `real numbers`.
For example, between `0.0` and `0.1`, there can be `0.01`, `0.001`, `0.00000000001`, etc.

This means it's impossible to fit any range of `real numbers` into 32 or 64 bits.

## 2. Real numbers are represented using 3 integers: sign, significant precision and exponent

Recall that you can express `integers` as sum of powers of 2, e.g:

<pre class="language-bash"><code class="language-bash">
15 = 8 + 4 + 2 + 1 = 1 * 2^3 + 1 * 2^2 + 1 * 2^z + 1 * 2^0
</code></pre>

Real numbers is also expressed in somewhat similar way:

<pre class="language-bash"><code class="language-bash">
usually = (-1)^sign * 2^(exponent-1023) * 1.significant
or      = (-1)^sign * 2^-1022           * 0.significant        (very small numbers)
</code></pre>

Basically, we partition the N bits into 3 parts to represent 3 integers:
- sign
- exponent
- significant precision


The exact number of bits used per part are:

<pre class="language-bash"><code class="language-bash">
| type    | sign | exponent | significant precision |
|---------|------|----------|-----------------------|
| 32-bit  |    1 |        8 |                    23 |
| 64-bit  |    1 |       11 |                    52 |
</code></pre>

For example, in 64 bits format:

<pre class="language-bash"><code class="language-bash">
| value  | sign | exponent    | significant precision                                |
|--------|------|-------------|------------------------------------------------------|
|    0.5 |    0 | 01111111110 | 0000000000000000000000000000000000000000000000000000 |
|        |    0 |        1022 |                                                    0 |
|--------|------|-------------|------------------------------------------------------|
|  -64.5 |    1 | 10000000101 | 0000001000000000000000000000000000000000000000000000 |
|        |    1 |        1029 |                                       35184372088832 |
|--------|------|-------------|------------------------------------------------------|
|  112.1 |    0 | 10000000101 | 1100000001100110011001100110011001100110011001100110 |
| approx |    0 |        1030 |                                     3384736594945638 |
</code></pre>

## 3. Many real numbers can only be approximated in floating point format

From the examples above, if you work the representation of 112.1 backwards, you will only get
around `112.099999999999994315658113919`.

This represents the most critical part about `floating point numbers`: they are mostly approximations.

Unlike `integers`, most `floating point numbers` doesn't have an exact representation
in binary systems. An iconic example is `0.1`, illustrated here:

<pre class="language-go line-numbers"><code class="language-go">
func main() {
	fmt.Printf("%0.64f\n", 0.1) // 0.1000000000000000055511151231257827021181583404541015625000000000
}
</code></pre>

## 4. Rounding/truncation doesn't help with retaining precision

During our confusion about `OpenTSDB`'s behavior, one of my colleagues suggested to round or
truncate the result read, up to 5 decimal digits.

It seems to work, but actually is a bad way to deal with floating point numbers.

First, because there's no way you can represent `0.1`, rounding/truncating doesn't reall change this.
You will get another approximated number.

Secondly, consider this code that simulate the effect of rounding.

<pre class="language-go line-numbers"><code class="language-go">
package main

import (
	"fmt"

	"gonum.org/v1/gonum/floats"
)

func main() {
	var x float64 = 0.1

	const n = 5
	const precision = 5

	fmt.Println("multiply:")
	for i := 0; i < n; i++ {
		x = floats.Round(x, precision)
		x *= float64(0.1)
		fmt.Printf("%0.64f\n", x)
	}

	fmt.Println("divide:")
	for i := 0; i < n; i++ {
		x = floats.Round(x, precision)
		x /= float64(0.1)
		fmt.Printf("%0.64f\n", x)
	}

	fmt.Printf("%0.64f\n", x)
}
</code></pre>

<pre class="language-bash"><code class="language-bash">
multiply:
0.0100000000000000019428902930940239457413554191589355468750000000
0.0010000000000000000208166817117216851329430937767028808593750000
0.0001000000000000000047921736023859295983129413798451423645019531
0.0000100000000000000008180305391403130954586231382563710212707520
0.0000010000000000000001665063486394613434526945638936012983322144

divide:
0.0000000000000000000000000000000000000000000000000000000000000000
0.0000000000000000000000000000000000000000000000000000000000000000
0.0000000000000000000000000000000000000000000000000000000000000000
0.0000000000000000000000000000000000000000000000000000000000000000
0.0000000000000000000000000000000000000000000000000000000000000000
0.0000000000000000000000000000000000000000000000000000000000000000
</code></pre>


Mathematically, you would expect to get `0.1` afterwards, but got `0` instead.

By rounding/truncating, you have lost more info in the `significant precision` bits
(i.e: used less than the amount of available bits).

The bottom line is: if it's already a float, don't try to round/truncate it when doing computaion.

## 5. However, rounding/truncate help when you are displaying and comparing

Now, if the user enter `0.1` into your program, what should you display back?
If you show them `0.1000000000000000055511151231257827021181583404541015625`, they will surely
be suprised (everyone is trained in math, but not many knows how computer store numbers, sadly).

Hence, you will probably need to round this off so it looks like `0.1`.

Similarly, when you are comparing 2 expressions, they might be giving a
["same same but different"](https://www.urbandictionary.com/define.php?term=same%20same%20but%20different)
result.

> TODO(mark): lookup for an example to illustrate here

You would also want to round/truncate things, upto a certain decimal digits there.

## 6. Keep the bit size consistent across your system

This is our problem when using `OpenTSDB 2.3.0`. Consider this flow:

1. User enter `0.1` in front-end code
2. Front-end encode `0.1` to JSON, sends to backend API
3. Backend API parse `0.1` in JSON, store as 64-bit approximation
4. Backend API encode the "approximated" `0.1` to JSON, sends to OpenTSDB via HTTP API
5. OpenTSDB parse `0.1` in JSON, **store as 32-bit approximation**
6. OpenTSDB encode 32-bit approximated value of `0.1` into JSON and send to backend

At step 5, our `OpenTSDB` breaks the convention. It truncates the significant precision bits.

When the result comes back and our backend try to store is value into 64-bit format, it thinks
that is a different number. There would be no problem if everyone uses th IEEE standard for 64-bit
`floating point numbers`.

This is an important point to remember when you write your own binary/JSON encoder/decoder
(or anything that deals with binary representation, for that matter). You **MUST** conform to
the standard to prevent nasty suprises.

We decided to upgrade our `OpenTSDB` to `2.4.0RC2`, because we do need 64-bit representation.

## 7. Floating point math is hard, read plenty articles

Read the IEEE standard, blog posts from other people, run experiment with your code AND pen
and paper.

Alaways be alert when it comes to floating point math :)
