#!/usr/bin/env bash
# This "./go" script is the build script.
# For context behind the "./go" script, please read these:
# https://blog.thepete.net/blog/2014/03/28/_-attributes-of-an-amazing-dev-toolchain/
# https://code.ofvlad.xyz/vlad/lightning-runner
set -e

_verify-prerequisites () {
  git config core.hooksPath .githooks

  if ! command -v stack &> /dev/null
  then
      _bad-message "Install haskell-stack to continue"
      exit 1
  fi

  if ! command -v hakyll-init &> /dev/null
  then
    stack install hakyll
    if [ $? -ne 0 ]; then
      _bad-message "Failed to install Hakyll, check README.md for troubleshooting"
      exit 1
    fi
  fi
}

⚡build () {
  _help-line "Compile the site generator and generate the site"
  stack build
  stack exec site build -- "$@"
  ⚡favicons
}

⚡clean () {
  _help-line "Clean generated site files"
  rm -rf _cache/* _site/*
}

⚡clean_all () {
  _help-line "Clean generated site files and site generator binaries"
  ⚡clean
  stack clean
}

⚡rebuild () {
  _help-line "Clean and then rebuild the generated site"
  ⚡clean
  ⚡build "$@"
}

⚡rebuild_all () {
  _help-line "Clean and then rebuild both the generated site and the site generator binary"
  ⚡clean_all
  ⚡build "$@"
}

⚡prebake() {
  _help-line "Compile only the site generator's and tests' dependencies"
  stack build --only-dependencies
  stack test --only-dependencies
}

⚡watch () {
  _help-line "Build the site generator, generate the site, and then run the preview server"
  ⚡build
  stack exec site watch -- "$@"
}

⚡rewatch() {
  _help-line "Rebuild the site generator, regenerate the site, and then run the preview server"
  ⚡rebuild
  stack exec site watch -- "$@"
}

⚡kill() {
  _help-line "Kill the site preview server if has gotten loose and run away!"
  lsof -ti tcp:8000 | xargs kill -9
}

⚡publish () {
  _help-line "Build the site and then publish it live"
  current_branch="$(git branch --show-current)"
  if [[ "$current_branch" != "main" ]]; then
    _bad-message "Can only publish from main branch; tried to publish from $current_branch"
    exit 1
  fi
  ⚡test_sync "main"

  sha="$(git log -1 HEAD --pretty=format:%h)"
  tag="$(date +'publish_%Y.%m.%d_%H.%M.%S')_$sha"

  git fetch origin _site
  mkdir -p _site
  rm -rf _site/* _site/.git
  cp -r .git/ ./_site/.git/
  pushd ./_site
  git switch _site
  git pull origin _site
  popd

  SITE_ENV=prod ⚡rebuild

  pushd ./_site
  # The rebuild's clean step wipes _site/*, favicons included. If favicon
  # generation was skipped (inkscape unavailable), they'd otherwise be committed
  # as deletions — restore the versions tracked on the _site branch instead.
  # No-op when favicons were regenerated (nothing shows up as deleted).
  deleted_favicons="$(git ls-files --deleted | grep -E 'favicon\.ico|images/grass-[0-9]' || true)"
  if [[ -n "$deleted_favicons" ]]; then
    _bad-message "Favicons not regenerated; retaining versions from the _site branch"
    echo "$deleted_favicons" | while IFS= read -r f; do git checkout -- "$f"; done
  fi
  git add .
  git commit -m "Build on $(date) generated from $sha"
  git push origin _site

  git tag -a "$tag" -m "Build on $(date) generated from $sha"
  git push origin "$tag"

  #rsync -ahp * bastion.thisfieldwas.green:/var/www/thisfieldwas.green/
  popd
}

⚡preview () {
  _help-line "Build the site and publish a preview build"
  SITE_ENV=preview ⚡rebuild
  rsync -ahp _site/* closet.thisfieldwas.green:/usr/share/nginx/preview.thisfieldwas.green/_site/
}

⚡test_sync () {
  _help-line "Verify that the current or specified local branch is up to date with the remote branch"

  branch=${1:-$(git branch --show-current)}
  git switch $branch
  git fetch origin $branch

  rev_parse_remote="$(git rev-parse origin/$branch)"
  rev_parse_local="$(git rev-parse $branch)"

  if [ "$rev_parse_local" != "$rev_parse_remote" ]; then
    _bad-message "Branch $branch not in sync with remote!"
    exit 1
  fi

  _good-message "Local branch $branch is up to date with remote"
}

⚡test () {
  _help-line "Run hspec tests"
  stack test
}

# ── Vicra port ──────────────────────────────────────────────────────────
#
# The Velith/Vicra rewrite of this generator lives in `vicra/` and consumes
# Vicra the way any third party would: a path dependency on a vial in another
# repository, no workspace membership, no registry.
#
# The three `VELITH_*_PATH` variables are **required**, and this is the one
# place that knows them. A release `velith` binary has no fallback at all; a
# debug build falls back to `<workspace root>/vials/...` computed from the
# *current directory*, which from this repo silently resolved to
# `thisfieldwas.green/vials/stdlib` and failed with a confusing missing-manifest
# error. Each points at a vial **root**, not its `src/`.
_vicra-env() {
  VELITH_REPO="${VELITH_REPO:-$HOME/workspace/velith-2}"
  if [ ! -d "$VELITH_REPO/vials/stdlib" ]; then
    echo "vicra: no Velith checkout at $VELITH_REPO" >&2
    echo "  set VELITH_REPO to point at it" >&2
    return 1
  fi
  VELITH="${VELITH:-$VELITH_REPO/target/debug/velith}"
  if [ ! -x "$VELITH" ]; then
    echo "vicra: no velith binary at $VELITH -- build it there with ./go build" >&2
    return 1
  fi
  export VELITH_STDLIB_PATH="$VELITH_REPO/vials/stdlib"
  export VELITH_VELDOC_PATH="$VELITH_REPO/vials/veldoc"
  export VELITH_VICRA_TEMPLATE_PATH="$VELITH_REPO/vials/vicra-template"
}

⚡vicra-check () {
  _help-line "Typecheck the Vicra port in vicra/"
  _vicra-env || return 1
  # Run from inside the vial rather than passing `--vial`: that flag needs a
  # `[workspace]` root above it, which a standalone third-party vial has no
  # reason to have. A bare `check` from the vial directory finds the manifest
  # by walking up, which is what a single-vial consumer actually does.
  ( cd vicra && "$VELITH" check )
}

⚡vicra-run () {
  _help-line "Run the Vicra port's entry point"
  _vicra-env || return 1
  "$VELITH" run vicra/src/Green/Site.vl -- "$@"
}

⚡vicra-source () {
  _help-line "Stage a rename-applied worktree of main in _source/, with its history intact"
  # `git_sha1` is `git log -1 -- <path>`: the last commit touching *that file*.
  # Committing the snake_case rename on this branch therefore changes it for
  # every template, and a page carrying `{{git_sha1}}` differs from the
  # published reference by a provenance stamp rather than by anything Vicra
  # rendered. Vicra is right and the diff is noise.
  #
  # So the comparison renders from a worktree of the commit the reference was
  # built from, with the rename applied but **not committed**: a dirty tree does
  # not change what `git log -1 -- <path>` reports. The branch keeps its
  # committed rename for development; this is what gets diffed.
  local at="${1:-main}"
  rm -rf _source
  git worktree prune
  git worktree add -q --detach _source "$at" || return 1
  python3 "${VELITH_REPO:-$HOME/workspace/velith-2}/scripts/rename-corpus-builtins.py" _source --apply > /dev/null || return 1
  echo "vicra-source: staged $at ($(git -C _source log --format=%h -1)) with the rename uncommitted"
}

⚡vicra-build () {
  _help-line "Render the staged source with Vicra into _vicra/"
  _vicra-env || return 1
  [ -d _source ] || ⚡vicra-source || return 1
  rm -rf _vicra
  mkdir -p _vicra
  # Same reason `vicra-check` runs from inside the vial: `--vial` wants a
  # `[workspace]` root, and a standalone consumer has none.
  # No `--` before the arguments: in the file-path run form they follow the
  # file directly, and `--` makes them vanish silently (main sees an empty
  # list and prints its usage).
  ( cd vicra && "$VELITH" run src/Green/Site.vl ../_source/site ../_vicra )
}

⚡vicra-diff () {
  _help-line "Render with Vicra and diff every emitted page against the published reference"
  _vicra-env || return 1
  [ -d _reference ] || ⚡vicra-reference || return 1
  ⚡vicra-build || return 1
  # Compare only what the slice emits. Diffing the whole tree would report
  # every page not yet ported as a failure, which says nothing and buries the
  # one thing this asserts.
  local status=0 compared=0 f rel
  while read -r f; do
    [ -z "$f" ] && continue
    rel="${f#_vicra/}"
    compared=$((compared + 1))
    if [ ! -f "_reference/$rel" ]; then
      echo "  NO REFERENCE: $rel" >&2
      status=1
    elif ! diff -u "_reference/$rel" "$f"; then
      echo "  DIFFERS: $rel" >&2
      status=1
    fi
  done <<< "$(find _vicra -type f | sort)"
  if [ "$compared" -eq 0 ]; then
    echo "vicra-diff: rendered nothing -- the slice is empty" >&2
    return 1
  fi
  if [ "$status" -eq 0 ]; then
    echo "vicra-diff: $compared page(s) byte-identical to the published build"
  fi
  return $status
}

⚡vicra-reference () {
  _help-line "Extract the published prod reference from the _site branch into _reference/"
  # The comparison target is the **`_site` git branch**, not the local `_site/`
  # directory. The directory is build output that Hakyll updates in place, so
  # switching SITE_ENV leaves it a mix -- it has been one, with `robots.txt`
  # carrying the prod host while `atom.xml`, `rss.xml`, `sitemap.xml` and
  # `index.html` still carried `http://localhost:8000` from a dev build. The
  # branch is a single published prod build and records which commit produced
  # it, so a diff against it is meaningful.
  local ref="${1:-origin/_site}"
  rm -rf _reference
  mkdir -p _reference
  git archive "$ref" | tar -x -C _reference || return 1
  echo "vicra-reference: extracted $ref -- $(git log --format=%s -1 "$ref")"
}

⚡datestamp () {
  _help-line "Generate ISO-8601 datestamp with time and offset"
  DATE=$(date +"%Y-%m-%dT%H:%M:%S%z")
  echo "$DATE"
  echo "$DATE" | pbcopy
  echo "Copied to clipboard: $DATE"
}

⚡favicons () {
  _help-line "Generate favicon and og:image from grass.svg"

  # Favicon generation needs inkscape (rasterize grass.svg) plus imagemagick
  # (convert/identify). These aren't always installable — e.g. locked-down work
  # machines can't install the inkscape cask. Since grass.svg changes rarely,
  # skip generation gracefully when any tool is missing and let the existing
  # favicons stand (publish restores them from the _site branch).
  for tool in inkscape convert identify; do
    if ! command -v "$tool" &> /dev/null; then
      _bad-message "$tool not found; skipping favicon generation (retaining existing favicons)"
      return 0
    fi
  done

  src_file="$(pwd)/site/images/grass.svg"
  out_dir="$(pwd)/_site/images"
  mkdir -p "$out_dir"

  sizes=(16 32 48 64 96 128 256 512 1024)
  out_files=()
  for x in ${sizes[@]}; do
    out_file="$out_dir/grass-${x}x${x}.png"
    out_files+=("$out_file")
    inkscape -w $x -h $x -o "$out_file" "$src_file"
    identify "$out_file"
  done

  convert "${out_files[@]}" "$(pwd)/_site/favicon.ico"
  identify "$(pwd)/_site/favicon.ico"
}

source ⚡
