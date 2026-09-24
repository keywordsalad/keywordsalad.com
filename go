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

  #rsync -ahp * bastion.keywordsalad.com:/var/www/keywordsalad.com/
  popd
}

⚡preview () {
  _help-line "Build the site and publish a preview build"
  SITE_ENV=preview ⚡rebuild
  rsync -ahp _site/* closet.keywordsalad.com:/usr/share/nginx/preview.keywordsalad.com/_site/
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
