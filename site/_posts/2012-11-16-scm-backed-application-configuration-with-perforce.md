---
title: SCM-Backed Application Configuration with Perforce
author: Logan McGrath
date: 2012-11-16T17:00:00-0600
published: 2012-11-16T17:00:00-0600
tags: configuration management, perforce, sinatra, angularjs
description: >-
  Continuing from my last post, I've forked Paul Hammant's original
  App-Config-App and modified it to work against Perforce. I've decided not to
  continue using Perforce Chronicle as it is primarily intended for content
  management.
comments: true
layout: post
---

Continuing from my [last post][], I've [forked][] Paul Hammant's original
App-Config-App and modified it to work against Perforce. I've decided not to
continue using Perforce Chronicle as it is primarily intended for content
management.

<!--more-->

With this version, App-Config-App is written in Ruby, mostly using [Sinatra][],
a lightweight web application framework. I'm still using [AngularJS][], but I've
also added a few other things:

* A .rvmrc file, so you automagically switch to Ruby 1.9.3
* A Gemfile, so you don't have to install everything individually :)
* [Sinatra-Contrib][] for view templating support
* [Rack Flash][] for flash messages
* [HighLine][] for masking passwords
* [json][] to manipulate JSON in native Ruby

## Getting it to work.

App-Config-App requires a couple things to work:

* Ruby 1.9.3 and Bundler
* p4 - the Perforce command line client
* p4d - the Perforce server

All installation and example setup details may be found in
[App-Config-App's README][].

## Using App-Config-App

When you login, you should see this screen:

{{imageFigure id: "img-app-start", src: "/images/app-config2/start.png"}}

You'll notice I made the extra effort to add colors and drop shadows :D The
application works from the project root in Perforce, so the files in each branch
are viewable here. Clicking on "Dev" > "aardvark_configuration.html" will bring
up a form for editing `aardvark_configuration.json` as in the previous version:

{{imageFigure id: "img-app-config", src: "/images/app-config2/aardvark_configuration.png"}}

Changes to the form data are automatically saved. After making a view edits, you
can click "View Diff" to get the diffs or "Revert" your changes. Go ahead and
change the email address and fiddle around with the banned nicks, then go click
"Pending Changes":

{{imageFigure id: "img-app-changes", src: "/images/app-config2/changes.png"}}

This screen shows all files that were changed and their diffs as well. You can
"Revert" each file individually, and if you want to commit all changes, then
enter a commit message and click "Commit Changes". If you commit the changes and
go back to "Dev" > "aardvark_configuration.html", you'll see the new values in
the form:

{{imageFigure id: "image-app-commit", src: "/images/app-config2/aardvark_configuration-changed.png"}}

## Security and Permissions

Permissions and security are managed through Perforce. For users to be able to
login, they must have a user and client configured in Perforce. Those users must
also have permissions configured in order to view or modify files.

The `setup_example.rb` script creates three test users to demonstrate branch
permissions:

```markdown
Username        Password   Write     Read
-------------------------------------------------
sally-runtime   bananas    prod      staging, dev
jimmy-qa        apples     staging   dev
joe-developer   oranges    dev
```

Logging in as any of these users will hide branches that don't have at least
read-level access, and branches that don't have write-level access won't allow
changes.

**All users created by `setup_example.rb` are intended only as examples.** In
the real world, all application users should be setup with real logins and real
permissions.

It is this support for users and per-branch permissions that I am using Perforce
as the SCM backend rather than Git.

### Application Users

The `setup_example.rb` script also sets up three application users to
demonstrate how an application would consume configuration:

```markdown
Username   Password   Read
-----------------------------
dev-app    s3cret1    dev
qa-app     s3cret2    staging
prod-app   s3cret3    prod
```

In theory, an application would periodically poll [aardvark_configuration.md5][]
until the hash value changed, then load [aardvark_configuration.json][] and
reconfigure itself.

Application user accounts are configured in Perforce like any other user. I
highly recommend that application users be given ready-only access to individual
files rather than entire branches.

## Divergence

Right now, App-Config-App offers no UI tools for managing divergence and
merging. Merges must be performed outside App-Config-App, and the specific
safety nets to prevent nefarious change vulnerabilities are dependent on your
branch specs and permissions configuration.

There are also are no tools to manage conflicts of existing edits with incoming
changes from another user. If a Perforce sync fails due to a conflict, you are
best to revert all changes and enter them again.

## @TODO

### A better model for autosave

Autosave in AngularJS isn't very good. AngularJS doesn't integrate with DOM
events the way idiomatic JavaScript does, or provide a reasonable abstraction
the way Dojo Toolkit or JQuery do. Right now, autosave in App-Config-App
triggers with every key press in the config forms, and pummels the back-end
server with ajax posts.

I've also noticed that the autosave triggers even when a value is invalid. The
first time an email address, for example, becomes invalid, AngularJS will post
back the JSON, but without the invalid email address field--the invalid field is
entirely left out of the JSON structure. After that, AngularJS will stop
autosaving until the value is valid. There are also no measures in place to
prevent a user from leaving an invalid value and saving an incomplete JSON file.

### A better model for validation

AngularJS does not offer a good validation API. The validation API is quite
opaque and I haven't found any real examples using it. The built-in form
validation is inadequate. There are few ng-* HTML attributes exposing more than
basic configuration parameters, and no hooks offered as extension points.

For example, I'm using [regular expressions][] for date validation in
App-Config-App. There isn't a hook to provide custom validation checks, and
regular expressions don't perform sanity checks. Values such as "00/00/0000"
will pass validation.

### More example clients than the Java one needed

The [App-Config-Java][] client is enough to show the basic idea behind caching
and reloading configuration from App-Config-App. I would like to create a few
more examples in a couple different platforms, possibly also showcasing
&ldquo;[hot reconfiguration][]&rdquo; for feature toggles.

### Someone should port this to Subversion or TFS

App-Config-App should be usable by the largest possible audience. For instance,
if you're using Subversion, then you should be able to take advantage of the
existing infrastructure.

The reason I point out Subversion and TFS is largely due to support of
per-branch permissions.

[last post]: {{get_url '_posts/2012-11-07-using-perforce-chronicle-for-application-configuration.md'}}
[forked]: https://github.com/lmcgrath/App-Config-App/
[AngularJS]: http://angularjs.org/
[App-Config-App's README]: https://github.com/lmcgrath/app-config-app/blob/master/README.md
[Sinatra]: http://www.sinatrarb.com/
[Sinatra-Contrib]: https://github.com/sinatra/sinatra-contrib
[Rack Flash]: http://nakajima.github.com/rack-flash/
[Highline]: http://highline.rubyforge.org/
[json]: http://rubygems.org/gems/json
[aardvark_configuration.md5]: http://localhost:9292/dev/aardvark_configuration.md5
[aardvark_configuration.json]: http://localhost:9292/dev/aardvark_configuration.json
[regular expressions]: http://docs.angularjs.org/api/ng.directive:input.text
[App-Config-Java]: https://github.com/lmcgrath/app-config-java
[hot reconfiguration]: http://paulhammant.com/2012/07/10/app-config-workflow-using-scm/
