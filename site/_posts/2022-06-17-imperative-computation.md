---
title: Imperative Computation in Functional Programming
description: Leveraging the case of the previous effect to indicate whether computation of the next effect should proceed.
author: Logan McGrath
date: 2022-05-30T10:04:42-0700
published: 2022-06-17T08:47:11-0700
tags: functional programming, programming, scala, design patterns, contexts
layout: post
comments: true
og:
  image:
    url: /images/tags/functional-programming/functional-grass-512x512.png
    alt:  Leveraging the case of the previous effect to indicate whether computation of the next effect should proceed.
code_repo: https://github.com/keywordsalad/embracing-nondeterminism-code/tree/part3
---

Remember **functors** and **applicatives**? In my last post {{linked_title "_posts/2022-06-05-enabling-control-flow.md"}} we explored how functors and applicatives abstract over **desired** and **undesired cases** of **contexts** in order to express control flow and permit independent computation. In this post we will explore **monads** and how to leverage their specific abstraction to express **imperative** control flow.

<!--more-->

> **This post is part of a series:**
>
> 1. {{linked_title "_posts/2022-03-15-contexts-and-effects.md"}}
> 2. {{linked_title "_posts/2022-06-05-enabling-control-flow.md"}}
> 3. **{{title}}**

_The code that accompanies this post may be found [here]({{code_repo}} "Accompanying code repository")._

## Motivating monads as a design pattern

Recall the `Functor` and `Applicative` typeclasses:

:::{.numberLines}
```scala
trait Functor[F[_]] {
  def map[A, B](fa: F[A])(f: A => B): F[B]
}
object Functor {
  def apply[F[_]: Functor]: Functor[F] = implicitly[Functor[F]]
}

trait Applicative[F[_]] extends Functor[F] {
  def pure[A](a: A): F[A]

  def ap[A, B](ff: F[A => B])(fa: F[A]): F[B]

  def map[A, B](fa: F[A]): F[B] =
    ap(pure(f))(fa)

  def map2[A, B, C](fa: F[A], fb: F[B])(f: (A, B) => C): F[C] =
    ap(ap(pure(f.curried))(fa))(fb)

  def sequence[A](listFa: List[F[A]]): F[List[A]] =
    listFa.foldLeft(pure(List[A]()))((fListA, fa) => map2(fa, fListA)(_ :: _))
}
object Applicative {
  def apply[F: Applicative]: Applicative[F] = implicitly[Applicative[F]]
}
```
:::

> See [`Functor`]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/data/Functor.scala)
> and [`Applicative`'s]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/control/Applicative.scala)
> definitions in the sample repository.

`Functor` abstracts over contexts' **unknown cases** by _lifting_ a function via `map()` and applying it to instances of the term produced by the context if they exist. Control flow can't be expressed using `map()` because it does not permit the case of the context to be modified.

`Applicative` can create contexts in the **desired case** via `pure()`. It can also apply _lifted_ functions to _lifted_ arguments via `ap()` if _both_ contexts are in their **desired case**. `Applicative` is capable of expressing control flow through `ap()` because supplying either the lifted function or lifted argument in an **undesired case** will not permit further computation. The context in the **undesired case** is _propagated_ instead.

The problem however is that `Applicative` requires that _both_ the lifted function and lifted argument supplied to `ap()` be computed _before_ deciding whether to permit or halt further computation. This means that regardless of whether one of the function or argument contexts successfully computes, if the other one fails then the succeeding computation will be discarded. Both of these contexts' computation are independent, and `Applicative` will only use them if they are both are in their **desired case**.

All functions derived from `ap()` represent _all-or-nothing_ operations accordingly, and there is no practical way of ordering the computation of the arguments or preventing computation of all arguments if any preceding arguments fail to compute. Therefore, totally imperative programming is not possible using `Applicative`.

## Introducing imperative control flow

In order to restrict computation such that subsequent computation occurs only if previous computation succeeds, we must introduce a new abstraction over the cases of contexts.

What function might limit subsequent computations from running if a previous one has failed? There is one, and you've probably seen it out in the wild:

:::{.numberLines}
```scala
def flatMap[F[_], A, B](fa: F[A])(f: A => F[B]): F[B]
```
:::

What `flatMap()` does is allow for the injection of a context into a pipeline of computations to either permit computation to proceed or force it to halt. The function argument `f`, which you supply, has full control of what case the returned context `F[B]` should be in. If `F[B]` is in the **undesired case**, then all further computations are skipped, _propagating_ this **undesired case** instead.

Take for example this definition of a `filter()` function from the sample repository's `sbt console`:

:::{.numberLines}
```scala
scala> :load console/imports.scala

scala> def filter[A](fa: Option[A])(f: A => Boolean) = fa.flatMap {
     |   case a if f(a) => Some(a)
     |   case _ => None
     | }
def filter[A](fa: Option[A])(f: A => Boolean): Option[A]

scala> filter(Some(4))(_ % 2 == 0)
val res1: Option[Int] = Some(4)

scala> filter(Some(3))(_ % 2 == 0)
val res2: Option[Int] = None
```
:::

This example is only very simple. But it demonstrates the key enabling feature of `flatMap()`: we are able to _choose_ whether the context remains in its **desired case**. If we continue with this example, we can try chaining `map()` to the context returned by it.

:::{.numberLines}
```scala
scala> filter(Some(4))(_ % 2 == 0).map(n => s"The square of even number $n is ${n * n}")
val res3: Option[String] = Some(The square of even number 4 is 16)

scala> filter(Some(3))(_ % 2 == 0).map(n => s"The square of even number $n is ${n * n}")
val res4: Option[String] = None
```
:::

Thus we are able to gate `map()` behind the case of the context we return from the `flatMap()` we used in `filter()`, enabling imperative control flow.

## Monads

The `flatMap()` function is implemented using a new structure, a specialization of an applicative functor called a **monad**.

Monads are a specialization that arises in the type of `A` contained within a context `F[_]`:

* If `A` is an opaque type, then `F[A]` is a **functor**.
* If `A` is known to have type `A => B`, that is _`A` is a function_, then `F[A => B]` is an **applicative** functor.
* If `A` is known to have type `F[A]`, then this means that `F[F[A]]` is _nested within itself_ and thus a **monad**.

The `Monad` typeclass defines two functions, which by default are defined in terms of each other:

:::{.numberLines}
```scala
trait Monad[F[_]] extends Applicative[F] {

  def flatMap[A, B](fa: F[A])(f: A => F[B]): F[B] =
    flatten(map(fa)(f))

  def flatten[A](ffa: F[F[A]]): F[A] =
    flatMap(fa => fa)
}

object Monad {

  def apply[F[_]: Monad]: Monad[F] = implicitly[Monad[F]]
}
```
:::

This means that in order to implement an instance of the `Monad` typeclass, you must implement one of either `flatMap()` or `flatten()` or calling one will recurse until the stack overflows. Some contexts are best defined using one over the other, and you have the choice to pick between the two.

> [See here]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/control/Monad.scala) for the definition in the sample repository.

### Composing monads

`Monad` also defines a special composition operator. Recall that functions may be composed together, as in `h = g ∘ f` or _"h is g after f"_. This may be expressed in two ways in vanilla Scala. Try below using the sample repository's `sbt console`:

:::{.numberLines}
```scala
scala> val f: Int => String = n => (n * n).toString
val f: Int => String = $Lambda$5252/0x000000010195e040@7a76dc63

scala> val g: String => Int = n => n.length
val g: String => Int = $Lambda$5254/0x000000010195f840@3843085

scala> val h = g.compose(f)
val h: Int => Int = scala.Function1$$Lambda$5244/0x0000000101958840@388b1b42

scala> h(3)
val res1: Int = 1

scala> h(25)
val res2: Int = 3

scala> val hh = f.andThen(g)

scala> hh(3)
val res3: Int = 1

scala> hh(25)
val res4: Int = 3
```
:::

Scala out of the box affords the `compose()` and `andThen()` functions for composing functions of the form of `A => B` together. But these composition functions don't work for the signature of `flatMap()`, which operates on functions of the form of `A => F[B]`.

There is a form of composition for monads known as _[Kleisli composition][]_. This composition will lift the following function into the context returned by the previous function and apply it to the term contained within the context if the context is in the **desired case**. Kleisli composition is expressed using the _fish_ or _arrow operator_ `>=>` which is defined in the sample repository as an extension operator on functions of the form `A => F[B]` where `F[_]` has an instance of `Monad`:

:::{.numberLines}
```scala
implicit class KleisliCompositionOps[F[_], A, B](val f: A => F[B]) extends AnyVal {

  def >=>[C](g: B => F[C])(implicit F: Monad[F]): A => F[C] =
    (x: A) => F.flatMap(f(x))(g)
}
```
:::

> [See here]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/syntax/monad.scala#L48) for the definition in the sample repository.

Kleisli composition is useful for monads in that from two or more `flatMap()`-compatible functions a single function may be created that takes an unlifted argument and produces an output context as the result of applying each function in sequence first to the unlifted argument and then to each individual functions' output contexts.

:::{.numberLines}
```scala
scala> :load console/imports.scala

scala> val squareEvens: Int => Either[String, Int] = {
     |   case n if n % 2 == 0 => Either.right[String, Int](n * n)
     |   case n               => Either.left[String, Int](s"$n is an odd number")
     | }
val squareEvens: Int => Either[String,Int] = $Lambda$5278

scala> val piEvenLastDigit: Int => Either[String, Int] = { n =>
     |   val nPi = n * scala.math.Pi
     |   val lastDigit = nPi.toString.last.asDigit
     |   if (lastDigit % 2 == 0) {
     |     Either.right[String, Int](lastDigit)
     |   } else {
     |     Either.left[String, Int](s"$n * Pi's last digit is an odd number")
     |   }
     | }

scala> val evensSquaredWithEvenLastDigitsOfPi = squareEvens >=> piEvenLastDigit
val evensSquaredWithEvenLastDigitsOfPi: Int => Either[String,Int] = Monad$Syntax$KleisliCompositionOps$$$Lambda$5241

scala> evensSquaredWithEvenLastDigitsOfPi(2)
val res1: Either[String,Int] = Right(2)

scala> evensSquaredWithEvenLastDigitsOfPi(23)
val res2: Either[String,Int] = Left(23 is an odd number)

scala> evensSquaredWithEvenLastDigitsOfPi(48)
val res3: Either[String,Int] = Left(2304 * Pi's last digit is an odd number)
```
:::

Above we are using the composed function directly, but because it has the form of `A => F[B]` we are able to apply it using `flatMap()` as well:

:::{.numberLines}
```scala
scala> Either.right[String, Int](2).flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res4: Either[String,Int] = Right(2)

scala> Either.right[String, Int](19).flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res5: Either[String,Int] = Left(19 is an odd number)

scala> Either.right[String, Int](4).flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res6: Either[String,Int] = Left(16 * Pi's last digit is an odd number)

scala> Either.left[String, Int]("Nothing's gonna happen").flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res7: Either[String,Int] = Left(Nothing's gonna happen)
```
:::

Thus Kleisli composition is to `flatMap()` as function composition is to `map()`, with the key difference being that the case of the context may be changed by any one of the composed functions. Kleisli composition is powerful in that entire processes may be composed into step-by-step operations and applied as a single unit to a context via `flatMap()` or directly to an unlifted argument.

### The _for comprehension_ and imperative programming

Scala provides a syntax sugar over `flatMap()` in the form of the _[for comprehension][]_. Any type that provides both the `map()` and `flatMap()` functions are able to participate in this syntax sugar, and it allows for monadic operations to be expressed as if they were written as procedural code. This means that you don't have to rely on Kleisli composition alone to express complex flows of logic, and that you also aren't limited to single-argument input/output pipelines of functions.

As a simple example, here are the above functions leveraged using a for comprehension:

:::{.numberLines}
```scala
scala> :load console/imports.scala

scala> val squareEvens: Int => Either[String, Int] = {
     |   case n if n % 2 == 0 => Either.right[String, Int](n * n)
     |   case n               => Either.left[String, Int](s"$n is an odd number")
     | }
val squareEvens: Int => Either[String,Int] = $Lambda$5278

scala> val piEvenLastDigit: Int => Either[String, Int] = { n =>
     |   val nPi = n * scala.math.Pi
     |   val lastDigit = nPi.toString.last.asDigit
     |   if (lastDigit % 2 == 0) {
     |     Either.right[String, Int](lastDigit)
     |   } else {
     |     Either.left[String, Int](s"$n * Pi's last digit is an odd number")
     |   }
     | }

scala> val evensSquaredWithEvenLastDigitsOfPi: Int => Either[String, Int] =
     | n => for {
     |   squaredEvenN <- squareEvens(n)
     |   piEvenLastDigitN <- piEvenLastDigit(squaredEvenN)
     | } yield piEvenLastDigitN

val evensSquaredWithEvenLastDigitsOfPi: Int => Either[String,Int] = $Lambda$5263

scala> evensSquaredWithEvenLastDigitsOfPi(2)
val res1: Either[String,Int] = Right(2)

scala> Either.right[String, Int](2).flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res2: Either[String,Int] = Right(2)

scala> Either.left[String, Int]("Nothing's gonna happen").flatMap(evensSquaredWithEvenLastDigitsOfPi)
val res3: Either[String,Int] = Left(Nothing's gonna happen)
```
:::

The key thing that a for comprehension allows for is multiple arguments to be lowered from contexts and applied to a multi-argument function that returns another context. An example of such an operation might look like this:

:::{.numberLines}
```scala
def sendMarketingUpdate(updateId: Int, userId: Int): Future[MarketingUpdateStatus] =
  for {
    user <- loadUser(userId)
    marketingTemplate <- loadTemplate(updateId)
    composedEmail <- composeEmailForUser(user, marketingTemplate)
    result <- sendEmail(composedEmail)
  } yield result.toMarketingUpdateStatus
```
:::

Because `composeEmailForUser()` requires both a `user` and `marketingTemplate`, this operation function is more easily expressed using a for comprehension than Kleisli composition.

However you don't have to abandon applicative functions when using for comprehensions. The above function may be written to independently load the user and template:

:::{.numberLines}
```scala
def sendMarketingUpdate(updateId: Int, userId: Int): Future[MarketingUpdateStatus] =
  for {
    (user, marketingTemplate) <- (loadUser(userId), loadTemplate(updateId)).tupled
    composedEmail <- composeEmailForUser(user, marketingTemplate)
    result <- sendEmail(composedEmail)
  } yield result.toMarketingUpdateStatus
```
:::

The code can be trimmed just a little bit more, as well:

:::{.numberLines}
```scala
def sendMarketingUpdate(updateId: Int, userId: Int): Future[MarketingUpdateStatus] =
  for {
    composedEmail <- (loadUser(userId), loadTemplate(updateId)).mapN(composeEmailForUser)
    result <- sendEmail(composedEmail)
  } yield result.toMarketingUpdateStatus
```
:::

This means that with applicative functions, you can express imperative code to some degree by gating functions dependent on a set of inputs, such as `composeEmailForUser()`, and lean on monad's `flatMap()` for when you need absolute ordering of computations. These two styles of computation thus complement each other.

## Becoming a Monad

In order to become a `Monad`, an effect type must implement the typeclass. Let's implement instances for the usual suspects, `Option`, `Either`, and `List`. As `Monad` is a specialization of `Applicative` and thus also `Functor`, we can simply upgrade our `Applicative` instances to become `Monad`s:

:::{.numberLines}
```scala
implicit val optionMonad: Monad[Option] = new Monad[Option] {

  // keep existing definitions

  override def flatten[A](ffa: Option[Option[A]]): Option[A] =
    ffa match {
      case Some(fa) => fa
      case None     => None
    }
}

implicit def eitherMonad[X]: Monad[Either[X, *]] = new Monad[Either[X, *]] {

  // keep existing definitions

  override def flatten[A](ffa: Either[X, Either[X, A]]): Either[X, A] =
    ffa match {
      case Right(fa) => fa
      case Left(x) => Left(x)
    }
}

implicit val listMonad: Monad[List] = new Monad[List] {

  // keep existing definitions

  override def flatten[A](ffa: List[List[A]]): List[A] =
    ffa
      .foldLeft(List[A]()) { (outerResult, as) =>
        as.foldLeft(outerResult) { (innerResult, a) =>
          a :: innerResult
        }
      }
      .reverse
}
```
:::

> See the instances of `Monad` for
> [`Option`]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/data/Option.scala#L159-L218),
> [`Either`]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/data/Either.scala#L114-L174),
> and [`List`]({{code_repo}}/src/main/scala/green/thisfieldwas/embracingnondeterminism/data/List.scala#L239-L302) in the sample repository.

In the above example, I implemented the `Monad` instances using `flatten()`. These could alternatively be implemented using `flatMap()`. You might try to do this if you wish as an exercise.

## Monad laws

How do we know that `Option`, `Either`, and `List`'s `Monad` instances are well-behaved as monads? Like functors, [monads are formally defined][] in the higher math of [category theory][] and expected to conform to a set of laws.

There are **three monad laws**, which must hold for all monads in addition to the applicative and functor laws.

1. **Preservation of left identity**: Kleisli composition of `pure()` with a function of form `f: A => F[B]` applied to an unlifted argument `a: A` is the same as applying the unlifted argument directly to `f: A => F[B]`.
2. **Preservation of right identity**: Kleisli composition of a function of form `f: A => F[B]` with `pure()` applied to an unlifted argument `a: A` is the same as applying the unlifted argument directly to `f: A => F[B]`.
3. **Associativity**: Kleisli composition of three functions `f: A => F[B]`, `g: B => F[C]`, and `h: C => F[D]` applied to an argument `a: A` produces the same result regardless of grouping: `(f >=> g) >=> h` is the same as `f >=> (g >=> h)`.

Note that `pure()` is interpreted as the identity element of the monad, as it merely lifts a value into its context, unmodified.

### Defining monad laws as properties

The monad laws may be defined as `scalacheck` properties, as [previously for applicatives][]. These properties assert "for all" instances of a particular monad `F[_]`, the properties should hold. We know we have a well-behaved monad if the property checks pass.

:::{.numberLines}
```scala
trait MonadLaws { this: Laws with ApplicativeLaws =>

  implicit def arbitraryFA[F[_]: LiftedGen, A: Arbitrary]: Arbitrary[F[A]] = Arbitrary(arbitrary[A].lift)

  /** Defined per Monad laws taken from the Haskell wiki:
    * [[https://wiki.haskell.org/Monad_laws#The_three_laws]]
    *
    * These laws extend the Applicative laws, so `checkApplicativeLaws[F]()` should be
    * executed alongside this function.
    *
    * The laws as defined here leverage Kleisli composition, which is defined
    * using the operator `>=>` in terms of `flatMap()`, to better highlight the
    * left and right identities and associativity that should be exhibited by
    * the composition of monadic operations.
    *
    * @param TT The type tag of the context
    * @tparam F The context type being tested
    */
  def checkMonadLaws[F[_]: Monad: LiftedGen]()(implicit TT: TypeTag[F[Any]]): Unit = {
    property(s"${TT.name} Monad composition preserves left identity") {
      // The identity of Kleisli composition simply lifts a value into the
      // context. Kleisli composition of a function after `pure()` when applied
      // to a value will produce the same result as applying the function
      // directly to the value.
      forAll(for {
        a <- arbitrary[Int]
        h <- arbitrary[Int => F[String]]
      } yield (a, h)) { case (a, h) =>
        val leftIdentity = ((_: Int).pure[F]) >=> h
        leftIdentity(a) mustBe h(a)
      }
    }
    property(s"${TT.name} Monad composition preserves right identity") {
      // The identity of Kleisli composition simply lifts a value into the
      // context. Kleisli composition of `pure()` after a function when applied
      // to a value will produce the same result as applying the function
      // directly to the value.
      forAll(for {
        a <- arbitrary[Int]
        h <- arbitrary[Int => F[String]]
      } yield (a, h)) { case (a, h) =>
        val rightIdentity = h >=> ((_: String).pure[F])
        rightIdentity(a) mustBe h(a)
      }
    }
    property(s"${TT.name} Monad composition is associative") {
      // `flatMap()` and thus Kleisli composition are both associative.
      // This means that your program may be factored with these operations
      // in any arbitrary grouping and the output will be the same.
      forAll(for {
        a <- arbitrary[Double]
        f <- arbitrary[Double => F[String]]
        g <- arbitrary[String => F[Int]]
        h <- arbitrary[Int => F[Boolean]]
      } yield (a, f, g, h)) { case (a, f, g, h) =>
        val assocLeft = (f >=> g) >=> h
        val assocRight = f >=> (g >=> h)
        assocLeft(a) mustBe assocRight(a)
      }
    }
  }
}
```
:::

> [See here]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/control/MonadLaws.scala) for the definition of the trait.

Our laws specs can now extend this trait and assert that their tested contexts conform to the monad laws. For example, the laws spec for `Option`:

:::{.numberLines}
```scala
class OptionLaws extends Laws with FunctorLaws with ApplicativeLaws with MonadLaws {

  import Option.Instances._

  implicit val optionLiftedGen: LiftedGen[Option] = new LiftedGen[Option] {

    override def lift[A](gen: Gen[A]): Gen[Option[A]] =
      Gen.lzy(
        Gen.oneOf(
          Gen.const(None),
          gen.map(Some(_)),
        )
      )
  }

  checkFunctorLaws[Option]()
  checkApplicativeLaws[Option]()
  checkMonadLaws[Option]()
}
```
:::

> See the laws specs for [Option]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/data/OptionSpec.scala#L127-L146),
> [Either]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/data/EitherSpec.scala#L112-L136),
> and [List]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/data/ListSpec.scala#L159-L181)

### Implications of the monad laws

You may have noticed that the characteristics of the monad laws are somewhat different from the functor and applicative laws. They build using composition directly, specifically Kleisli composition, but don't assert that function composition generally is retained within their contexts the same way that functor or applicative's laws do.

**Let's reiterate the laws:**

1. _Left identity_
2. _Right identity_
3. _Associativity_

These laws specifically also define another typeclass called a **monoid**, which is a specialization of a **semigroup** that adds an _identity_ or _empty_ element. The identity element is special in that combining it with any other element produces that other element.

:::{.numberLines}
```scala
trait Monoid[M] extends Semigroup[M] {

  def empty: M
}

object Monoid {

  def apply[M: Monoid]: Monoid[M] = implicitly[Monoid[M]]
}
```
:::

Monoids are nearly as common as semigroups, as not all semigroups are monoids, and you've probably used quite a few of them:

* [`List`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/ListSpec.scala#L80-L88)'s identity element is the empty `List()`.
* [`String`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/StringSpec.scala#L21-L29)'s identity element is the empty `""`.
* [`Int`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/IntegerSpec.scala#L10-L22)'s identity element is `0`.
* [`Set`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/SetSpec.scala#L80-L88)'s identity element is the empty `Set()`.

The key difference between monads and these monoids above is that monads form an _additive function_ in contrast to _additive data_. Functions of the form `A => F[B]` are combinable using `>=>` to produce new functions of the same form, and this operation is associative, which means that monads form semigroups under Kleisli composition. They form monoids as the `pure()` function satisfies the identity element in that it doesn't alter its argument when composed to either the left or right side of a function of the form `A => F[B]`.

> To be really pedantic, monads are functors. All functors in Scala are functors from Scala types into other Scala types, making them endofunctors because they map back into the same category: the category of Scala types.
>
> This affirms an infamous joke:
>
>> Haskell gets some resistance due to the complexity of using monads to control side effects. Wadler tries to appease critics by explaining that _"a monad is a monoid in the category of endofunctors, what's the problem?"_
>
> -- From James Iry's "[A Brief, Incomplete, and Mostly Wrong History of Programming Languages][]"

In addition to being monoids, Scala's
[`List`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/ListSpec.scala#L32-L63) and
[`Set`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/SetSpec.scala#L32-L63) are also monads and I have asserted that they are well behaved with property checks.
[`Option`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/OptionSpec.scala#L30-L61)
and [`Either`]({{code_repo}}/src/test/scala/green/thisfieldwas/embracingnondeterminism/stdlib/EitherSpec.scala#L30-L61)
are also monads. You've probably been using them as such without realizing!

The real takeaway from the monad laws is that you get _imperative computation as a composable structure_. You retain referential transparency in that associativity guarantees that your operations may be grouped arbitrarily, allowing you to factor the steps of your program with a large degree of freedom.

## What is enabled by monads?

The `flatMap()` function primarily enables imperative programming through abstraction. It strictly requires that its current context be evaluated into its **desired case** before applying its function argument against the term it contains. If the context is in the **undesired case**, then all subsequent computation halts and this case _propagates_ instead. Thus, in contrast to applicative's _all-or-nothing_ operations, monads offer _one-after-another_ operations. Both evaluation styles complement each other and may be freely mixed, of course.

### Beware the imperative trap

Because monads allow you to express functional programming in a style that closely mirrors procedural code, especially within the syntax of the for comprehension, it's very easy to fall into a trap of procedural spaghetti even though you're using a functional programming abstraction. This occurs because for comprehensions allow you to expose multiple terms and pass them around at varying points within the for comprehension, mimicking procedural state-passing:

:::{.numberLines}
```scala
def performTransaction(accountNo: Long, amount: Currency): Future[Unit] =
  for {
    account <- loadAccount(accountNo)
    transaction <- beginTransaction(account)
    balance <- getBalance(account)
    newBalance = balance + amount
    _ <- {
      if (newBalance < Currency.from(0)) {
        endTransaction(transaction)
          .flatMap(_ => failTransaction("Insufficient funds"))
      } else {
        updateBalance(newBalance)
          .flatMap(_ => endTransaction(transaction))
      }
    }
  } yield ()
```
:::

This code would be hard to refactor as there's a lot of arguments being passed around. It's procedural code wrapped in a for comprehension! A smell in particular is a `Unit`-return, as it indicates something is causing side-effects with no way to sense what actions have been performed or what state the total operation finished in.

## Going forward

These three abstractions, **functors**, **applicatives**, and **monads**, are just the beginning to a rich series of tools that you can leverage to express solid, maintainable, and provable programs. But you may have noticed a few capabilities are missing so far in this series, such as how to:

* _Raise errors_, specifically as **undesired cases**, agnostic of context.
* _Recover from errors_, transforming **undesired cases** to **desired**, agnostic of context.
* _Attempt an alternate operation_ if the first one fails.
* _Collect effectful operations_ agnostic of container.
* Allow for _partial successes and failures_ within collective operations.
* _Compose effect types_ to take advantage of their combined characteristics.
* _Encode DSLs_ as effects.

In my next post, we will explore **raising and recovering from errors** agnostic of context, so that your code may abstract against typeclasses but still be able to force and recover from **undesired cases**.

[Kleisli composition]: https://blog.ploeh.dk/2022/04/04/kleisli-composition/
[for comprehension]: https://docs.scala-lang.org/tour/for-comprehensions.html
[monads are formally defined]: https://en.wikipedia.org/wiki/Monad_(category_theory)#Formal_definition
[category theory]: https://en.m.wikipedia.org/wiki/Category_theory
[previously for applicatives]: {{get_url "_posts/2022-06-05-enabling-control-flow.md"}}#defining-the-applicative-laws-as-properties
[A Brief, Incomplete, and Mostly Wrong History of Programming Languages]: https://web.archive.org/web/20220609203110/https://james-iry.blogspot.com/2009/05/brief-incomplete-and-mostly-wrong.html
