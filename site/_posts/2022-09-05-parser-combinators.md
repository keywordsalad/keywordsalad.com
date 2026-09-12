---
title: "Parser Combinators"
description: Combining functions to build a parser
author: Logan McGrath
date: 2022-06-20T10:37:30-0700
published: 2022-09-05T12:55:18-0700
tags: functional programming, programming, scala, design patterns, combinators, parsing
layout: post
comments: true
code_repo: https://github.com/keywordsalad/parser-combinators/tree/post/parser-combinators
---

In my post {{linked_title "_posts/2022-06-16-set-function.md"}} I introduced function combinators as a pattern for building complex logic. I added the caveat that combinators should not be used to define sets, which I feel undersells the value of combinators as a general programming pattern. In this post we will explore combinators as an applied solution to a common problem: _parsing text_.

<!--more-->

> The code that accompanies this post may be found [here]({{code_repo}} "Accompanying code repository").

## The `parse()` function

Parsers begin as a simple signature, a function called `parse()` from an input `Int` offset and `String` to either a failed offset or matched offset.

:::{.numberLines}
```scala
trait Parse extends (((Int, String)) => Either[Int, (Int, String)])
```
:::

**This signature enables two capabilities:**

1. On failure the offset where the `parse()` function failed to recognize the input is returned.
2. On success the offset where the end of the recognized input is returned along with the input, which can be used to extract the matched input and also be passed to the next `parse()` function as its starting point.

_Below we will introduce specialized `parse()` functions within the `Parse` companion object._

### Implementing the first `parse()` function

The first `parse()` function we can implement is `satisfy()`. This function works by using a _predicate_ to match the `Char` at the current offset within the input.

:::{.numberLines}
```scala
def satisfy(predicate: Char => Boolean): Parse = {
  case (offset, input) if predicate(input(offset)) => Right(offset + 1, input)
  case (offset, _)                                 => Left(offset)
}
```
:::

This function is very simple; it can match single `Char`s easily.

:::{.numberLines}
```scala
"satisfy" can {
  "recognize a single character if it matches" in {
    val parse = satisfy(_ == 'a')
    parse((0, "abc")) shouldBe Right((1, "abc"))
  }
  "reject a character that doesn't match" in {
    val parse = satisfy(_ == 'a')
    parse((0, "xyz")) shouldBe Left(0)
  }
}
```
:::

With a little more work, it can be made to match Unicode `Char`s!

:::{.numberLines}
```scala
"recognize Unicode characters" in {
  val parse = {
    val pattern = Pattern.compile("\\\\p{Ll}")
    satisfy(c => pattern.matcher(CharBuffer.wrap(ArrayCharSequence(Array(c)))).matches())
  }
  parse((0, "π is pi")) shouldBe Right((1, "π is pi"))
  parse((0, "Π is uppercase pi")) shouldBe Left(0)
}
```
:::

Thus `satisfy()` is sufficiently versatile to meet a wide range of needs, though it requires that you be creative in how you create your predicates.

### Specializing the `satisfy()` parse function

Based on the previous example, recognizing a character with a regular expression is really cumbersome. Let's define a `parse()` function that makes this kind of recognizer easier to define.

:::{.numberLines}
```scala
def satisfyRegex(pattern: String): Parse = {
  val pattern = Pattern.compile("\\\\p{Ll}")
  satisfy(c => pattern.matcher(CharBuffer.wrap(ArrayCharSequence(Array(c)))).matches())
}
```
:::

Now a regular expression can be provided without creating the boilerplate to support the predicate directly. The test to show that the `satisfyRegex()` function is identical to the original, as well.

:::{.numberLines}
```scala
"satisfyRegex" can {
  "recognize a character using a regex" in {
    val parse = satisfyRegex("\\\\p{Ll}")
    parse((0, "π is lowercase pi")) shouldBe Right((1, "π is lowercase pi"))
    parse((0, "Π is uppercase pi")) shouldBe Left(0)
  }
}
```
:::

## Parser combinators

So far we have two `parse()` functions: `satisfy()` and `satisfyRegex()`. Neither of these do more than recognize single `Char`s and advance the offset of the input. We want to be able to recognize sections of input.

Let's start first by recognizing two letters, one after the other, with our first combinator: the `&` combinator.

**What is a combinator?** A combinator is a function that builds a function from other functions. In this case, we will be building a new `parse()` function that _combines_ the capabilities of two other `parse()` functions.

Here is how we define the combinator on the `Parse` trait:

:::{.numberLines}
```scala
def &(second: Parse): Parse =
  input => this(input).flatMap(nextInput => second(nextInput))
```
:::

By using the `&` combinator we can define new `parse()` functions from two existing ones, allowing us to recognize a length of input. Here is a short example using a test:

:::{.numberLines}
```scala
"&" which {
  val parseFirst = satisfy(_ == 'a')
  val parseSecond = satisfy(_ == 'b')
  val parse = parseFirst & parseSecond
  "recognizes the first letter and then the second" in {
    parse((0, "abc")) shouldBe Right((2, "abc"))
  }
  "recognizes the first letter but rejects the second" in {
    parse((0, "acc")) shouldBe Left(1)
  }
  "rejects the first letter" in {
    parse((0, "ccc")) shouldBe Left(0)
  }
}
```
:::

### Recognizing alternatives with a second `parse()` function

Given the possibility that a `parse()` function may fail to recognize the input, is there a way that we could fall back to an alternative `parse()` function that would be able to recognize it? There is!

The next combinator is `|`. This combinator tries to recognize the input with the first `parse()` function and will use the second `parse()` function if it fails.

Here is how we define the combinator on the `Parse` trait:

:::{.numberLines}
```scala
def |(second: Parse): Parse =
  input => this(input).left.flatMap(_ => second(input))
```
:::

This combinator allows us to choose between possibilities of input. A simple example is demonstrated in its test:

:::{.numberLines}
```scala
"|" which {
  val parseFirst = satisfy(_ == 'a')
  val parseSecond = satisfy(_ == 'b')
  val parse = parseFirst | parseSecond
  "recognizes using the first parse function" in {
    parse((0, "abc")) shouldBe Right((1, "abc"))
  }
  "recognizes using the second parse function" in {
    parse((0, "bcd")) shouldBe Right((1, "bcd"))
  }
  "rejects if neither parser recognizes" in {
    parse((0, "cde")) shouldBe Left(0)
  }
}
```
:::

## Putting together what we have so far

Now that we have sequential and alternative recognizers, we can start defining more complex `parse()` functions. Let's start with a term recognizer.

:::{.numberLines}
```scala
def term(value: String): Parse =
  value.tail.foldLeft(satisfy(_ == value.head))((p, c) => p & satisfy(_ == c))
```
:::

By using the classic functional programming idiom `foldLeft()`, we can build a term recognizer using only the functions and combinators we have defined so far. Slick!

Here is the test to show that our `term()` function works, including one using the `|` combinator with a second term:

:::{.numberLines}
```scala
"term" which {
  val parse = term("banana")
  "recognizes a word" in {
    parse((0, "banana apple")) shouldBe Right((6, "banana apple"))
  }
  "rejects input that doesn't match" in {
    parse((0, "banapple")) shouldBe Left(4)
  }
  "alternate with a second term" in {
    val combined = parse | term("apple")
    combined((0, "banana orange")) shouldBe Right((6, "banana orange"))
    combined((0, "apple orange")) shouldBe Right((5, "apple orange"))
  }
}
```
:::

## How do we get the matched input?

There's a small problem with our `parse()` function. It only gives the end offset of the recognized input and there isn't any way to collect or modify the recognized input. This means we have to change the signature of the `parse()` function.

### Modeling input

First, we need to consider what our input needs to model. Our input is characterized by two things, the text being parsed and the offset that the current `parse()` function is operating from. This defines a _cursor_:

:::{.numberLines}
```scala
case class InputCursor(offset: Int, input: String) {

  def peek: Char = input(offset)

  def +(amount: Int): InputCursor = copy(offset = offset + amount)

  def -(amount: Int): InputCursor = copy(offset = offset - amount)
}
```
:::

A cursor allows us to hold a reference to the input itself, unmodified, but increment or decrement the position we are at within the input. As our `parse()` functions recognize input, they increment the offset so that the following functions have a place to start.

### Modeling output

The output of the `parse()` function changes as well. We want to be able to return the recognized input. This changes the signature of the `parse()` function:

:::{.numberLines}
```scala
trait Parse extends (InputCursor => Either[InputCursor, (???, InputCursor)])
```
:::

But what should be returned isn't entirely clear. If the recognizer matches a `Char`, then the `parse()` function should return a `Char`. But what about the `term()` recognizer? That should return a `String`. What if we're using a function that recognizes integers?

This means that the `parse()` function is generic in its return type:

:::{.numberLines}
```scala
trait Parse[A] extends (InputCursor => Either[InputCursor, (A, InputCursor)])
```
:::

The overall return type is also becoming verbose by using the general-purpose `Either` class. So let's define something specialized:

:::{.numberLines}
```scala
sealed trait ParseResult[A]

case class ParseFailure[A](cursor: InputCursor) extends ParseResult[A]

case class ParseSuccess[A](result: A, cursor: InputCursor) extends ParseResult
```
:::

_**But before we use it this means that all of our function signatures have to change!**_

## The new shape of the `parse()` function

Our `parse()` function will soon become generic its return type, which makes it _covariant_ in that `parse()` can be adapted to produce anything we want. What the `parse()` function produces is contextualized by whether it is recognized in the input, and as such there is a chance that nothing is produced because the input wasn't recognized. This means the the `parse()` function forms a specific structure: a _context_ with the shape of a **functor**.

> I have written about contexts, functors, and related structures extensively. Start from the following article if these terms are unfamiliar to you: **{{linked_title "_posts/2022-03-15-contexts-and-effects.md"}}**

Take a look at what the `parse()` function signature will look like after we start using `ParseResult`:

:::{.numberLines}
```scala
trait Parse[A] extends (InputCursor => ParseResult[A])
```
:::

Let's compare that with the shape of a _functor_: `Parse[A]` and `F[A]`.

In fact, `ParseResult[A]` has the same shape!

To save some time, I've added the [`cats`][] library to leverage the functor typeclass abstraction. First, let's implement the `Functor` typeclass for `ParseResult[A]`:

:::{.numberLines}
```scala
object ParseResult {

  implicit val parseResultFunctor: Functor[ParseResult] = new Functor[ParseResult] {

    override def map[A, B](fa: ParseResult[A])(f: A => B): ParseResult[B] =
      fa match {
        case ParseFailure(failedCursor)        => ParseFailure(failedCursor)
        case ParseSuccess(matched, nextCursor) => ParseSuccess(f(matched), nextCursor)
      }
  }
}
```
:::

This allows us to transform the contents of the `ParseResult` at-will.

### Changing the return type of the `parse()` function itself

The `parse()` function itself is a functor. But that doesn't quite make sense: it doesn't contain anything because it's a function!

However, it still has the _shape_ of a functor. Recall the new signature that it will have `InputCursor => ParseResult[A]`. It's covariant in `A`, giving it the shape of `F[A]`. So how do we treat it as a functor? Let's change the signature first and see how the code breaks.

_**And all of the `parse()` functions break!**_ So we will step through each and get them working again, and we will leverage the `ParseResult` functor.

1. **`satisfy()`** only requires that we change the return:

    ```scala
    def satisfy(predicate: Char => Boolean): Parse[Char] = cursor =>
      if (predicate(cursor.peek)) {
        ParseSuccess(cursor.peek, cursor + 1)
      } else {
        ParseFailure(cursor)
      }
    ```

2. **The `&` combinator** has a `flatMap()` screaming to be abstracted:

    ```scala
    def &[B](second: Parse[B]): Parse[(A, B)] =
      cursor =>
        this(cursor) match {
          case ParseFailure(failedCursor) => ParseFailure(failedCursor)
          case ParseSuccess(firstMatched, nextCursor) =>
            second(nextCursor).map(secondMatched => (firstMatched, secondMatched))
        }
    ```

3. **The `|` combinator** requires an abstraction to act on failure:

    ```scala
    def |(second: Parse[A]): Parse[A] =
      cursor =>
        this(cursor) match {
          case _: ParseFailure    => second(cursor)
          case s: ParseSuccess[_] => s
        }
    ```

4. **term()** is affected the most, and the rewrite requires threading a cursor through a lot of `parse()` functions and handling the results. It even requires adding a specialized `parse()` function to always return a `Unit` and succeed!

    ```scala
    def unit: Parse[Unit] = cursor => ParseSuccess((), cursor)

    def term(value: String): Parse[String] =
      outerCursor =>
        value
          .map(c => satisfy(_ == c))
          .foldLeft(unit)((up, cp) => innerCursor => (up & cp)(innerCursor).map(_ => ()))(outerCursor)
          .map(_ => value)
    ```

### The `parse()` function as a functor

First, let's define a `Functor` instance for `Parse`:

:::{.numberLines}
```scala

implicit val parseInstances: ParseInstances = new ParseInstances()

class ParseInstances extends Functor[Parse] {

  override def map[A, B](fa: Parse[A])(f: A => B): Parse[B] =
    cursor => fa(cursor).map(f)
}
```
:::

This leverages the ability to use `ParseResult` as a functor. The new `map()` function for `parse()` allows us to change what the function produces by wrapping the supplied function `f: A => B` in a new `parse()` function! We are effectively able to change the "contents" of the function this way.

Using `parse()` as a functor, we can refactor the `term()` function:

:::{.numberLines}
```scala
def term(value: String): Parse[String] =
  value
    .map(c => satisfy(_ == c))
    .foldLeft(unit)((up, cp) => (up & cp).map(_ => ()))
    .map(_ => value)
```
:::

Notice that we no longer need to manually pass around the input cursor. We've simply folded a sequence of `parse()` functions to `Char` down to a single `parse()` function to `String`. The functor abstraction allowed us to declaratively create this function without having to worry about the plumbing; we don't even have to declare the cursor as an explicit input.

### Sequential parsing with the `parse()` function

We can't apply the functor abstraction to the `&` combinator as it sequences one `parse()` function after a successful application of a preceding `parse()` function. This means that it executes _imperatively_. The specific abstraction we have to use for this kind of execution is a **[monad][]**.

A monad is a specialization of a functor. We can replace our functor instance of the `parse()` function with one for monad:

:::{.numberLines}
```scala
class ParseInstances extends Monad[Parse] {

  override def flatMap[A, B](fa: Parse[A])(f: A => Parse[B]): Parse[B] =
    cursor =>
      fa(cursor) match {
        case ParseFailure(failedCursor)        => ParseFailure(failedCursor)
        case ParseSuccess(matched, nextCursor) => f(matched)(nextCursor)
      }

  override def tailRecM[A, B](a: A)(f: A => Parse[Either[A, B]]): Parse[B] = {
    @tailrec
    def recur(cursor: InputCursor, parse: Parse[Either[A, B]]): ParseResult[B] = {
      parse(cursor) match {
        case ParseFailure(failedCursor) => ParseFailure(failedCursor)
        case ParseSuccess(matched, nextCursor) =>
          matched match {
            case Left(a)  => recur(nextCursor, f(a))
            case Right(b) => ParseSuccess(b, nextCursor)
          }
      }
    }
    cursor => recur(cursor, f(a))
  }

  override def pure[A](x: A): Parse[A] =
    cursor => ParseSuccess(x, cursor)
}

```
:::

**`Monad` gives us the following capabilities:**

* We still get `map()` through a default definition.
* `pure()` which lifts a result into a `parse()` function that will always be recognize the input without consuming from the cursor.
* `flatMap()` which allows for sequential execution of one `parse()` function after another.
* `tailRecM()` which provides _[monadic recursion][]_. We'll be using this later.

Given this monad instance for the `parse()` function, we can now rewrite the `&` combinator:

```scala
def &[B](second: Parse[B]): Parse[(A, B)] =
  for {
    a <- this
    b <- second
  } yield (a, b)
```

This `&` combinator becomes very small. Again, like the `term()` function we no longer have to manually thread a cursor through each function: we simply declare how the individual `parse()` functions relate and modify what they collectively produce. Additionally the `parse()` function is now able to participate in the _for comprehension_ syntax sugar as it defines both the `map()` and `flatMap()` operations.

### Alternative parsing with the `parse()` function

The monad abstraction provides _imperative_ execution. This means that we can't implement the `|` combinator using `flatMap()` as it requires both the first and second `parse()` functions to succeed in sequence. We need to lean on the `Alternative` typeclass to allow us to try one function but use another if it fails.

Let's start at the deciding factor of success or failure of the `parse()` function: `ParseResult`. We need to define a function that allows choice between a successful result and a failed one:

:::{.numberLines}
```scala
sealed trait ParseResult[A] {

  def orElse(other: => ParseResult[A]): ParseResult[A] =
    this match {
      case _: ParseFailure[_] => other
      case _: ParseSuccess[_] => this
    }
}
```
:::

If our result is `ParseSuccess` then `orElse()` returns the current result. If our result is `ParseFailure` then `orElse()` evaluates the `other` argument and returns that instead, regardless of success or failure. This allows us to try producing one result, use it if it's successful, and otherwise try producing the other result.

With this ability to abstract the selection of success or failure, we can now abstract the same ability for the `parse()` function by adding the `Alternative` typeclass trait to the `extends` clause of `ParseInstances`:

:::{.numberLines}
```scala
class ParseInstances extends Monad[Parse] with Alternative[Parse] {

  // monad declarations...

  override def empty[A]: Parse[A] =
    cursor => ParseFailure(cursor)

  override def combineK[A](x: Parse[A], y: Parse[A]): Parse[A] =
    cursor => x(cursor).orElse(y(cursor))
}
```
:::

**`Alternative` gives us the following capabilities:**

* `empty` provides a `parse()` function that always fails, because it's empty.
* `combineK()` allows for the combination of the results produced by the `parse()` function. Concretely, this means that the first function is attempted, and if it fails then the second function is attempted.

Armed with abstraction over selection of success or failure, we can redefine the `|` combinator:

:::{.numberLines}
```scala
import cats.syntax.semigroupk._

def |(second: Parse[A]): Parse[A] =
  this <+> second
```
:::

That's all the `|` combinator is now: `this <+> second`. By leveraging the `Alternative` abstraction, we are able to declaratively specify with a single operator that the first `parse()` function should be attempted before the second.

## Putting it all together

Let's try defining a `parse()` function that's more complex than what we've covered so far: a generic programming language identifier recognizer.

**This `parse()` function should be able to recognize the following inputs:**

* Any alphabetic letter followed by zero or more of alphabetic letters, digits, or underscores.
* Any underscore as long as it is followed by one or more of alphabetic letters, digits, or underscores.

There's a lot going on in those two requirements. Let's break it down:

* Recognize one of a range of possible matches from one `parse()` function.
* Require zero or more matches from one `parse()` function.
* Require at least one or more matches from one `parse()` function.

First, we need to define a `parse()` function that can recognize one match from any number of parsers. More specifically, we need two functions in order to do this:

:::{.numberLines}
```scala
def empty[A]: Parse[A] = Alternative[Parse].empty[A]

def oneOf[A](parsers: Seq[Parse[A]]): Parse[A] =
  parsers.foldLeft(empty[A])(_ | _)

def oneOf[A](parsers: Parse[A]*): Parse[A] =
  oneOf(parsers)
```
:::

Here we have a convenience `empty[A]` parser that leverages the `Alternative` typeclass we implemented earlier. By defining a `oneOf()` function that accepts a sequence of `parse()` functions, we can fold each function to the left starting with an empty parser and alternate each one. The first `parse()` function that matches will return a success result. If none match, then the final failure result will be returned instead.

How do we use this to accept any one of the alphabetic letters? It would be tedious to write a `satisfy()` function for each letter by hand, so instead let's specify the letters we want as a range and map over them:

:::{.numberLines}
```scala
def alpha: Parse[Char] =
  oneOf((('a' to 'z') ++ ('A' to 'Z')).map(c => satisfy(_ == c)))
```
:::

We can do the same thing with digits:

:::{.numberLines}
```scala
def digits: Parse[Char] =
  oneOf(('0' to '9').map(c => satisfy(_ == c)))
```
:::

We also need a `parse()` function that allows for recognizing zero or many matches and another for one or many:

:::{.numberLines}
```scala
def zeroOrMany[A](parse: Parse[A]): Parse[Seq[A]] =
  Monad[Parse].tailRecM(Seq[A]()) { seq =>
    parse.map(a => (seq :+ a).asLeft[Seq[A]]) |
      seq.asRight[Seq[A]].pure[Parse]
  }

def oneOrMany[A](parse: Parse[A]): Parse[Seq[A]] =
    (parse & zeroOrMany(parse)).map { case (first, rest) => first +: rest }
```
:::

Notice how our `zeroOrMany()` function leverages the _monadic recursion_ capability from earlier.

Armed with these functions, we can now recognize identifiers!

:::{.numberLines}
```scala
def identifier: Parse[String] = {
  val underscore = satisfy(_ == '_')
  val rest = alpha | digit | underscore
  val alphaHead = alpha & zeroOrMany(rest)
  val underscoreHead = underscore & oneOrMany(rest)
    (alphaHead | underscoreHead).map { case (head, rest) => (head +: rest).mkString }
}
```
:::

But that seems almost too easy! We better write a bunch of tests to make sure this recognizes input as expected.

:::{.numberLines}
```scala
"identifier" can {
  "recognize a single alpha letter" in {
    identifier(InputCursor(0, "a banana")) shouldBe ParseSuccess("a", InputCursor(1, "a banana"))
    identifier(InputCursor(0, "y not")) shouldBe ParseSuccess("y", InputCursor(1, "y not"))
  }
  "recognize an alpha letter followed by alphas, digits, and underscores" in {
    identifier(InputCursor(0, "a_123_Word <-")) shouldBe
      ParseSuccess("a_123_Word", InputCursor(10, "a_123_Word <-"))
  }
  "reject a single underscore" in {
    identifier(InputCursor(0, "_ rejected")) shouldBe ParseFailure(InputCursor(1, "_ rejected"))
  }
  "recognize an underscore followed by alpha digits and underscores" in {
    identifier(InputCursor(0, "_a_123_Word <-")) shouldBe
      ParseSuccess("_a_123_Word", InputCursor(11, "_a_123_Word <-"))
  }
  "reject a digit" in {
    identifier(InputCursor(0, "1234")) shouldBe ParseFailure(InputCursor(0, "1234"))
  }
}
```
:::

## What's next?

With parser combinators you are able to process text into any structure you need. In this post we only covered tuples, `Seq`s, and `String`s, but parse functions can be written to produce full AST's or even interpret scripted programs.

**Some subjects for next time:**

3. Parsing mathematical infix expressions and calculating their results.
1. Error handling and recovery, an especially tricky subject!
2. Stateful parsing, which requires lifting the `parse()` function into a higher-order monad.

[`cats`]: https://typelevel.org/cats/
[monad]: https://thisfieldwas.green/blog/2022/06/17/imperative-computation/
[monadic recursion]: https://typelevel.org/cats/typeclasses/monad.html#tailrecm
