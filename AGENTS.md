# AGENTS.md

Guidance for AI agents and contributors working in this repository.

## Overview

`keywordsalad.com` is Logan McGrath's personal website — a blog published at
<https://keywordsalad.com>. It is a **custom static site generator written in
Haskell on top of [Hakyll](https://jaspervdj.be/hakyll/)** (package name:
`green`). Site content and assets live in `site/`, the generator compiles them
into `_site/`, and publishing pushes that output to a dedicated git branch.

The repo bundles a bespoke Handlebars-style templating engine, **Hakyllbars**
(`src/Green/Hakyllbars/`), and a companion authoring CLI (`author`) alongside
the site generator (`site`).

## ⚠️ Read this first (gotchas)

- **There is no Makefile.** `README.md` says `make init` and `bin/site` /
  `bin/author` call `make build`, but no Makefile exists — those references are
  stale. The real task runner is **`./go`** (see below).
- **`./go` and the `⚡` file are intentional.** `./go` is a "lightning runner":
  its subcommands are bash functions named with a `⚡` prefix, and it sources the
  file literally named `⚡` at the repo root. Neither is junk — don't delete them.
- **Builds are strict: `-Wall -Werror`.** Combined with `-Wunused-packages`,
  `-Wincomplete-patterns`, `-Wname-shadowing`, etc., this means an unused
  dependency, a non-exhaustive `case`, or a shadowed name **fails the build** —
  it is not just a warning.
- **Edit `package.yaml`, never `green.cabal`.** The project uses
  [hpack](https://github.com/sol/hpack); `green.cabal` is generated from
  `package.yaml`. Hand-edits to the cabal file get clobbered.
- **Two templating engines coexist.** `Green.Hakyllbars` (current) and
  `Green.Template` (legacy) are both imported by `Green/Site/*`. Migration is in
  progress — prefer Hakyllbars for new work, but expect both.
- **`.githooks` is referenced but absent.** `./go` runs
  `git config core.hooksPath .githooks`, yet no such directory exists yet. Not
  an error; just don't be surprised.

## Common commands

Everything goes through `./go <subcommand>`. Run `./go` with no argument for the
generated help listing.

| Command            | What it does                                                           |
| ------------------ | ---------------------------------------------------------------------- |
| `./go build`       | `stack build`, then `stack exec site build`, then regenerate favicons  |
| `./go watch`       | Build, then run the Hakyll preview server on <http://localhost:8000>   |
| `./go rewatch`     | `rebuild`, then run the preview server                                 |
| `./go rebuild`     | `clean` then `build`                                                   |
| `./go rebuild_all` | `clean_all` then `build`                                               |
| `./go clean`       | Remove `_cache/*` and `_site/*`                                        |
| `./go clean_all`   | `clean` + `stack clean`                                                |
| `./go prebake`     | Pre-compile dependencies only (`stack build/test --only-dependencies`) |
| `./go test`        | Run the Hspec test suite (`stack test`)                                |
| `./go kill`        | Kill a runaway preview server holding TCP port 8000                    |
| `./go datestamp`   | Print + copy an ISO-8601 timestamp (handy for frontmatter `date`)      |
| `./go favicons`    | Generate favicons / og:image from `site/images/grass.svg`              |
| `./go publish`     | Build with `SITE_ENV=prod` and publish (main-only — see **Deploy**)    |
| `./go preview`     | Build with `SITE_ENV=preview` and rsync to the preview host            |

If you'd rather bypass `./go`, the underlying commands are plain Stack:

```sh
stack build                    # compile the generator (and library)
stack exec site build          # generate the site into _site/
stack exec site watch          # generate + preview server on :8000
stack test                     # run the test suite
stack exec author -- <args>    # run the authoring CLI (see below)
```

## Toolchain

- **Build tool:** [Stack](https://docs.haskellstack.org/), resolver
  `lts-22.27` (GHC 9.6.5), `allow-newer: true`, no `extra-deps`.
- **Project format:** hpack (`package.yaml` → `green.cabal`).
- **System dependencies** (`Brewfile`): `haskell-stack`, `sass`. Install with
  `brew bundle`.
- **Undocumented extras:** `./go favicons` additionally needs `inkscape` and
  ImageMagick (`convert` / `identify`) — these are *not* in the `Brewfile`.
- **macOS note:** if `stack install hakyll` fails with
  `can't load framework: Cocoa`, see the workaround in `README.md`
  (`DYLD_INSERT_LIBRARIES=<path>/macos11ghcwa.dylib stack install hakyll`).

## Project layout

```text
.
├── go                     # ⚡ task runner (build/watch/test/publish/…) — the entrypoint
├── ⚡                      # lightning-runner harness sourced by ./go
├── config.yaml            # site config + environments (default/prod/review/preview)
├── package.yaml           # hpack project definition — SOURCE OF TRUTH
├── green.cabal            # generated from package.yaml — do not hand-edit
├── stack.yaml             # Stack resolver / build config
├── broken-links.cfg       # input for the broken-link redirect generator
├── hie.yaml               # HLS (haskell-language-server) cradle
├── Brewfile               # Homebrew system deps
├── app/
│   ├── site/Main.hs       # `site` executable → siteMain (the Hakyll generator)
│   └── author/Main.hs     # `author` executable → authorMain (authoring CLI)
├── src/Green/             # the library (green:lib)
│   ├── Green.hs           # entry points siteMain / authorMain / loadSiteConfig
│   ├── Site.hs            # master Hakyll rules — wires up every content type
│   ├── Site/*.hs          # one module per content type: Blog, Pages, HomePage,
│   │                      #   Css, Js, Images, Feed, Sitemap, Robots, Code,
│   │                      #   Static, Templates, BrokenLinks
│   ├── Hakyllbars/*       # the custom templating engine (lexer→parser→AST→eval)
│   ├── Template/*         # legacy templating layer (being phased out)
│   ├── Lens/*             # microlens helpers (TemplateHaskell-generated)
│   ├── Config.hs          # config.yaml parsing → SiteConfig
│   ├── Route.hs, Command.hs, Compiler.hs, Common.hs, Util.hs
├── test/                  # Hspec suite (hspec-discover; see Testing)
├── site/                  # *** ALL WEBSITE CONTENT & ASSETS *** (Hakyll provider dir)
├── _site/                 # generated output (gitignored; also a published git branch)
└── _cache/                # Hakyll cache (gitignored)
```

`Green/Site.hs` (`site :: SiteConfig -> Rules ()`) is the map of the whole site:
it composes the per-feature modules under `src/Green/Site/`, each of which owns
one `match` pattern (images, js, scss, templates, homepage, pages, blog, code,
static, feed, sitemap, robots).

## Content authoring

All content lives under `site/` (the Hakyll `provider-directory`):

- **Posts:** `site/_posts/YYYY-MM-DD-slug.md` (Jekyll-style naming).
- **Drafts:** `site/_drafts/` (disallowed in `robots.txt`; served only in the
  `preview` environment).
- **Standalone pages:** `site/_pages/*.md` (e.g. `contact.md`, `resume.md`) plus
  root listing pages `index.html`, `blog.html`, `archives.html`, `drafts.html`.
- **Error pages:** `site/_errors/` (`404.md`, `500.md`).
- **Assets:** `site/css/` (SCSS), `site/js/`, `site/images/` (by topic),
  `site/code/` (code listings referenced by posts).

### Post frontmatter

YAML frontmatter delimited by `---`. Real example
(`site/_posts/2022-09-05-parser-combinators.md`):

```yaml
---
title: "Parser Combinators"
author: Logan McGrath
date: 2022-06-20T10:37:30-0700       # authored time (ISO-8601 with offset)
published: 2022-09-05T12:55:18-0700  # publish time
tags: functional programming, scala, combinators, parsing   # comma-separated
description: >-
  Combining functions to build a parser…
layout: post
comments: true
code_repo: https://github.com/keywordsalad/parser-combinators/…
---
```

Common keys: `title`, `description`, `author`, `date` / `published` / `updated`,
`tags`, `layout` (`post` / `page`), `comments`. Optional: `contentClass` (maps
to a per-post SCSS class), `stylesheets:` and `preloadImages:` (lists),
`code_repo`, `changefreq`.

In the body: `<!--more-->` marks the teaser/excerpt cutoff, and inline
Hakyllbars helpers work, e.g. `{{linkedTitle "_posts/…​.md"}}` and `{{code_repo}}`.
Posts are Pandoc-flavored Markdown (fenced code, `:::{.numberLines}` attribute
blocks, etc.).

### The `author` CLI

```sh
stack exec author -- draft -t "Post Title" [-c category]   # scaffold a draft
stack exec author -- publish -f <file>                     # publish a draft
```

Note: as of now these subcommands (`Green/Command.hs`) mostly print what they
*would* do — treat them as stubs / a work in progress, not finished tooling.

## Templating (Hakyllbars)

Hakyllbars is a custom Handlebars-style engine implemented in
`src/Green/Hakyllbars/` as a pipeline: **Lexer → Parser → AST → Evaluator**
(`Source/Lexer.hs`, `Source/Parser.hs`, `Ast.hs`, `Compiler.hs`), with
`Context.hs` for name resolution and `Field*.hs` for built-in helpers (dates,
git info, HTML/URL escaping, control flow).

Template files live under `site/`:

- `site/_layouts/` — page skeletons that chain via `{{@applyLayout "…"}}`
  (e.g. `post` → `default` → `skeleton`).
- `site/_partials/` — reusable fragments included with `{{partial "…"}}`.
- `site/_templates/` — snippet templates (image figures, YouTube embeds, etc.).

Syntax cheat sheet:

```text
{{ name }}                       value / field lookup
{{ fn arg }}                     function application
{{ target.field }}              field access
{{ arg | filter }}              pipe through a filter (e.g. {{ published | dateAs shortDate }})
{{ #if x }}…{{ #else }}…{{ #end }}   conditional
{{ #for xs }}…{{ item }}…{{ #end }}  iteration
{{ partial "main-nav" }}         include a partial
{{ @applyLayout "default" }}     wrap current output in a layout
{{- … -}}                        trim surrounding whitespace
```

Prefer Hakyllbars for new template work; the legacy `Green.Template` layer is
still wired in but on its way out.

## Styling

SCSS lives under `site/css/`, entry point `main.scss`, which `@import`s
partials organized into `elements/`, `layout/`, `mixins/`, and `vendor/`, plus
per-page/per-post sheets in `pages/` and `posts/` (referenced from frontmatter
via `stylesheets:`). Compilation happens inside the generator
(`src/Green/Site/Css.hs`), not a separate build step. The `inflate-css` /
`inflate-js` flags in `config.yaml` control minification per environment (on in
dev, off in prod).

## Testing

- Framework: **Hspec**, auto-discovered. `test/Spec.hs` is just
  `{-# OPTIONS_GHC -F -pgmF hspec-discover #-}`, so any `test/**/*Spec.hs`
  exporting `spec :: Spec` is picked up automatically.
- Existing specs: `Green/RouteSpec.hs`, `Green/Site/BlogSpec.hs`,
  `Green/Hakyllbars/Source/LexerSpec.hs`,
  `Green/Hakyllbars/Source/ParserSpec.hs`.
- Shared helpers/DSL live in `Green/TestSupport*` and
  `Green/Hakyllbars/TestSupport.hs` — a small matcher DSL with infix
  combinators like `produces` and `rejectsWith` for asserting lexer/parser
  output.
- Run with `./go test` or `stack test`.

## Configuration & environments

`config.yaml` defines four environments via YAML anchors: `default`, `prod`,
`review`, `preview`. The active one is chosen by the **`SITE_ENV`** environment
variable (defaults to `default`) and loaded at runtime by `Green/Config.hs`
(`loadSiteConfig`, which also merges in process env and current time).

Key settings: `hakyll-config.provider-directory: site`,
`destination-directory: _site`, per-environment `host`, date/time display
formats, responsive `image-widths`, and the `debug-settings` inflate flags.

## Deploy

There is **no CI/CD** — deployment is manual via `./go`.

- **`./go publish`** — allowed only from `main`. It verifies the local branch is
  in sync with `origin/main`, fetches the `_site` branch into `./_site`, rebuilds
  with `SITE_ENV=prod`, commits the generated output onto the **`_site` git
  branch**, pushes it, and pushes an annotated `publish_YYYY.MM.DD_HH.MM.SS_<sha>`
  tag. (The final `rsync … bastion.keywordsalad.com` line is currently
  commented out — publishing = pushing built output to the `_site` branch.)
- **`./go preview`** — rebuild with `SITE_ENV=preview`, rsync to the preview host.

The published domain is pinned by `site/CNAME` (`keywordsalad.com`).

## Code conventions

- **Namespace:** everything under `Green.*`; the module path mirrors the
  directory tree. `Site/*` = one module per content type; `Hakyllbars/*` = the
  template engine; `*Spec` = test specs.
- **Formatting:** ormolu / fourmolu style (leading-comma lists, `where` layout)
  **by convention — there is no committed formatter config**, so match the
  surrounding code.
- **Extensions:** ~45 `default-extensions` are enabled globally in
  `package.yaml` (`OverloadedStrings`, `LambdaCase`, `BlockArguments`,
  `RecordWildCards`, `GADTs`, `TemplateHaskell`, …). Don't re-declare them per
  module.
- **Warnings are errors.** Because of `-Werror -Wunused-packages`, a dependency
  you add to `package.yaml` must actually be used, and patterns must be
  exhaustive, or the build fails.
- **Lenses:** microlens (`microlens` / `microlens-th`), generated via
  TemplateHaskell in `Green/Lens/*` and `Green/Config.hs`.
- **Spell check:** the repo tracks a [cSpell](https://cspell.org/) word list in
  `.vscode/settings.json`. When you introduce project jargon (a new tool,
  extension, identifier, or acronym) that the checker flags, add it there in the
  same change so the dictionary stays current for everyone.

## Key dependencies

`hakyll` (framework), `pandoc` (Markdown rendering), `parsec` (Hakyllbars
lexer/parser), `aeson` + `yaml` + `scientific` (config/metadata), `binary`
(template cache), `microlens(-th)` (lenses), `mtl` (monad stacks),
`optparse-applicative` (author CLI), `hjsmin` (JS minification), plus
`bytestring`, `text`, `unordered-containers`, `vector`, `network-uri`,
`process`, `time`, `filepath`, `directory`, `MissingH`, `data-default`.
