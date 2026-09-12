---
title: Highly available ssh tunnels
description: Duct tape performs better than systemd for keeping my webserver's ssh tunnels open
author: Logan McGrath
date: 2022-02-08T14:27:57-0800
published: 2022-02-09T16:02:28-0800
tags: self host, duct tape, yak shave
layout: post
comments: true
thumbnail: /images/tags/duct-tape/duct-tape-128x128.png
og:
  image:
    url: /images/tags/duct-tape/duct-tape-512x512.png
    alt: I duct-taped my ssh tunnels open!
---

In my previous post, {{linked_title "_posts/2021-12-11-reasons-why-my-website-is-offline.md"}}, I complained about `systemd` giving up when it fails to maintain `ssh` tunnels. In this post, I complain about `systemd` a bit more and how I gave up and stopped using it for managing my `ssh` tunnels.

<!--more-->

## Recall from my last post

1. My website is hosted from my _closet computer_, an old PC tower that sits in my bedroom closet
2. I have no static IP address
3. I maintain a remote `ssh` tunnel from a _bastion server_ hosted on a [`linode`](https://linode.com) to my _closet computer_
4. My host names' DNS point to the IP address of the _bastion server_
5. The _bastion server_ acts as a proxy into the `ssh` tunnels to the _closet computer_
6. The _closet computer_ handles the traffic it receives

**This is how it looks:**

```{.txt .nowrap}
         +-BASTION-SERVER--------------\
         | <=proxy=> | <===remote tunnel===> +-CLOSET-COMPUTER---\
SSH ---> | *:22----> | localhost:10022 | --> | *:22 SSH          |
HTTP --> | *:80----> | localhost:10080 | --> | *:80 goto 443 duh |
HTTPS -> | *:443---> | localhost:10443 | --> | *:443 my website  |
         `-----------+-----------------+     `-------------------+
```

## Failure, `systemd`, and duct tape

When my home internet connection is down, the remote `ssh` tunnels into the _closet computer_ fail to reopen. I have been using `systemd` on _closet computer_ to manage these tunnels, and `systemd` will eventually stop trying to open the tunnels if the network remains inaccessible. This retry-then-give-up behavior that `systemd` exhibits is to safeguard against situations such as the [thundering herd problem](https://en.wikipedia.org/wiki/Thundering_herd_problem) and [this Stackoverflow question](https://unix.stackexchange.com/questions/289629/systemd-restart-always-is-not-honored) highlights the specific levers used to configure `systemd`'s retry behavior. To paraphrase, `systemd` tries to recover within a fixed amount of tries before it gives up. Given that there is a retry limit, I can't reasonably extend the interval between tries so that my webserver recovers from a multi-hour outage without personally intervening. Moreover, I have to be physically next to the computer to recover it when this happens.

I can't make any guarantees about my internet connection's uptime and that I'm not at risk of unleashing a thundering herd against my own infrastructure. As a solution, `systemd` therefore does not make sense for keeping my `ssh` tunnels tunnels open.

As a better solution, one both held together by duct tape and recommended by [Vlad](https://internetwebsite.ofvlad.xyz), I have instead leveraged a simple cronjob:

```{.txt .nowrap}
*/5 * * * * ssh closet.thisfieldwas.green true || ssh -fN bastion.thisfieldwas.green
```

And added the following `ssh` configuration to the `autossh` user that opens the tunnels:

```{.txt .nowrap}
Host bastion.thisfieldwas.green
    Port 22
    RemoteForward 10022 localhost:22
    RemoteForward 10080 localhost:80
    RemoteForward 10443 localhost:443

Host closet.thisfieldwas.green
    ProxyCommand ssh jump_user@bastion.thisfieldwas.green -W 127.0.0.1:10022
```

This simple setup works by testing every five minutes whether the `ssh` tunnel to _closet computer_ is open by `ssh`'ing in by its external hostname via `ssh closet.thisfieldwas.green true`. There is an account named `jump_user` that is allowed to `ssh` into the tunnels open on the _bastion server_ and all `ssh` requests to _closet computer_ proxy through this user account. If an `ssh` connection cannot be made, then the remote tunnels are opened via `ssh -fN bastion.thisfieldwas.green`. This setup works when I reboot the _closet computer_, if I unplug my router, or if my internet connection drops.

I don't have to worry about my husband Corey accidentally tripping over the power cable because I know that he will plug it back in. When the _closet computer_ has powered back on, the tunnels open shortly after, and I'm offline for about five minutes for a small hiccup. For a self-hosted webserver, I think this is acceptable.

## @TODO

### A better way to recover from bad configuration

I've managed to disable my `ssh` tunnels three times now by changing host keys, the `autossh` user's keys, or the `jump_user`'s allowed keys. I don't have a quick solve for these instances beyond manually stepping through and fixing each `ssh` key error as it comes up, and I would much prefer a solution that's no more manual than running a single command.

### Better uptime, but from the closet

While using a cronjob to keep `ssh` tunnels open is a step towards higher availability for my website, I still have some distance to cover before my maximum downtime is one minute or less per incident. For instance, last Sunday, Feb 6, 2022 at 12pm PST, a 3.1 earthquake struck; its epicenter barely a mile from my home and it felt like a large truck had collided with the house! That earthquake could have cut my internet connection for multiple days.

> I am also a bad Californian and underprepared for a worse earthquake, but that's another matter.

I could simply host my website from the _bastion server_, but physically it's located in Fremont, CA, where it could still be knocked offline by a Bay Area earthquake. And sure, `linode` has other locations, but simply hosting from anywhere that isn't my closet would deprive me of a good yak shave.

As an exercise, I want to duct tape my own high availability solution together while still serving my website from my _closet computer_ as the primary node. It is fun to tell people that my professional site is coming out of the closet, after all.
