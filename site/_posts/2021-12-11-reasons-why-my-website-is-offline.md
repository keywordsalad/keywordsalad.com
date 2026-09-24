---
title: Reasons why my website is offline
description: >-
  My website comes out of the closet. This is from where I host several of my
  services, but the uptime is terrible!
author: Logan McGrath
published: 2021-12-11T14:43:25-0800
tags: self host, configuration management, tinfoil hat, duct tape, yak shave
layout: post
comments: true
thumbnail: /images/reasons-my-website-is-offline/coming-out-of-the-closet-128x128.png
og:
  image:
    url: /images/reasons-my-website-is-offline/coming-out-of-the-closet-512x512.png
    alt: The closet computer, in context
---

My website is hosted from a 2007 HP Pavilion tower that my best friend, Vlad, a privacy buff and [low-key minimalist](https://internetwebsite.ofvlad.xyz), found next to the garbage chute at his condo building.

<!--more-->

> It appears to have been someone’s gaming rig in its prior life, as it has a decent graphics card in it and `8GB` of RAM. Its abandonment was likely due to an upgrade to Windows 10, as upon boot, the login screen was so unresponsive that I could barely get the mouse to register movement and indicate for the computer to shut down again.

I previously complained to Vlad about how untrustworthy I felt cloud providers are regarding data privacy. He finds this computer by sheer accident, and that he uses Arch Linux, btw[^archbtw], so he recommends that I breathe new life into it using Arch, and to see if I can self host some of my own services.

_I present my journey into self hosting: a rant, a recipe, and a retrospective._

## Tinfoil hat time!

This sudden urge to self host is impulsed by the advent of [GitHub Copilot](https://copilot.github.com/). The concept of it sure spooks the hell out of me. Not the idea that my professional role as a software engineer has now become so fungible that I can be replaced by a machine, I feel that is inevitable regardless, but perhaps I feel an uncomfortable rumble of an economic ripple effect.

### My revulsion is more nuanced than replacement

My use of GitHub, a long-standing freemium product, has been in the form of a paid _Pro Plan_. I pay GitHub for this tier so that I can choose to configure privileged access to certain repositories as their contents I am either paid to develop or they are personal projects that I am prototyping.

Despite paying for a _Pro Plan_, if I make those repositories publicly available, then their contents are fed into Copilot without my express consent or knowledge. These contents are then analyzed and transformed into data for which _corporations_ will be paying untold sums of money to _another, single corporation_ in order to manifest their technical means of production, _I assume_. This means that in order to gain a competitive edge, a company must use Copilot to out-produce a competitor, paying GitHub a premium in order to do so. This creates a concentration of wealth with no clear distribution mechanism returning it to lower rungs of the economic ladder.

Personally, my passion projects, my many contributions to open source, every private repository that I accidentally made public that one time... in a phrase: _my labor_, has been consented to subsidize widening profit margins in GitHub's favor, and even to the favor of companies using Copilot. An articulation of late stage capitalism's twitching proboscis has lanced into another artery from which to siphon our collective value; our dessicated tissues will remain with no further need for a paycheck.

### I guess my revulsion is still about replacement

In a sense I am training my replacement. For free. Or at least in exchange for _"free hosting"_ that I'm also partially paying for. At a less sleazy company I would be told that "Bob" is my inexpensive replacement, that for the next two months I will be training him, and that I will continue to be paid before I bounce to another company where I will become paid again. In this scenario, however, I go to another company and _"Bob" is already working there, and he's using all of my stuff from the last place_.

### Not all sunshine and rainbows in the cloud

At this same time Apple announces that they are [scanning everyone's photos](https://towardsdatascience.com/apples-neuralhash-how-it-works-and-ways-to-break-it-577d1edc9838) to make sure they don't contain CSAM. Don't worry, they aren't actually looking to see if the photo contains CSAM. Until they aren't, because someone has to verify a positive match manually. They pinky-promise they will only ever look for CSAM, even though they could look for anything they wanted, or are told to look for.

I pay Apple to store my photos in the cloud. I thought they were encrypted, right? So now they're scanning them _on my device_ because they can't look at them on the cloud, because they're encrypted. `wat`.

### Bringing the cloud home

I am very motivated to get as much of my stuff out of third party services as I can, because--to generalize based on GitHub and Apple--it's likely that no third party can be trusted with my data _even if they are already making money off of it_. I resolve to divest Apple and Google of my cloud data, GitHub of my source code, and `linode` of my miscellaneous.

I talk to Vlad about my feelings regarding GitHub Copilot, Apple, and cloud providers in general. He offers to help me find a new option for hosting my data. Just a few days later he finds that computer next to the garbage chute, and with much excitement he calls me and cheerfully announces, _“It turns on!”_

What luck!

## Keeping my data in the closet

The computer sits in many pieces in my living room as I remove dust and await several shipments of spinning platter drives. Installing the array of drives to hold my media proves difficult: HP did not make this case amenable to changes or even provide much in the way of expansion bays. As a consequence, I now have a bunch of drives that are free-floating within the case. No matter though, the drives are each `4TB`, and so heavy that they are held in place by the friction induced by their own mass and density alone. Later, for a short period, the computer carries the hostname `gravwell` because of the density these drives give it.

I craft a script so that I can reproducibly install Arch Linux with disk encryption over `ssh`, comfortably iterating from my 2019 MacBook Pro until the installation looks correct. I nuke the computer and try again, just to be sure. Then several times more.

Once the OS is installed to satisfaction, I carefully arrange the drives within the case, making sure the case can tolerate some degrees of tilt in any direction without the drives sliding out of place. To my surprise, the computer can be rotated a full 90&deg; in any direction, and the drives don't move _at all_. I place the computer inside of the bedroom closet, where I can plug it directly into the router. The computer is now _closet computer_.

I spend some number of days iterating on _closet computer_ with an ~~[Ansible playbook](#)~~ to get the configuration just right. As long as the data array remains intact, I can incrementally add new disks to it by appending the disks to the playbook configuration. If the data array is lost, then I can nuke it and build anew just by running the playbook. _(To be clear, I lose data if this happens.)_

My goal with this playbook is primarily to retain a living snapshot of _closet computer_'s configuration. A secondary goal is to be able to provision a replacement system if the computer should ever cease to work. In my head I imagine the happy path for such an event looks like this:

1. A few keystrokes.
2. I then grab a coffee while the playbook runs.
3. When the playbook is complete, I can power off the computer.
4. I plug the drives into it.
5. I turn on the new _closet computer_. Huzzah.

Configuration management is Ansible's primary domain, it does this very well, and I now use it for my current three Linux-based systems.

> I both nuke and enjoy rebuilding my Arch Linux-powered _thonkpad_ on occasion. All changes I make to it are managed through its associated playbook. I use Arch Linux, btw.

## Hosting from inside the closet

_closet computer_ hosts a small set of services:

* My website is served by [`nginx`](https://www.nginx.com/), which also terminates `ssl`.
* My source code is hosted by [`gitea`](https://gitea.io/) with [`postgres`](https://postgresql.org/) to support its data.
* My data array provides `16TB` usable storage:
  * It's [just a bunch of disks](https://en.wikipedia.org/wiki/Non-RAID_drive_architectures#JBOD) with redundancy provided by [`snapraid`](https://www.snapraid.it/).
  * It's treated as a single disk using [`mergerfs`](https://github.com/trapexit/mergerfs).
  * It’s very nice to use, and I think a choice recommendation by Vlad.
* [`pi-hole`](https://pi-hole.net/)... I thought this would mean less ads, but it hasn't really helped much.

I’m using docker containers for `gitea` and `postgres`. `nginx` runs natively and provides routing into `gitea` or static `html` for its configured domains.

I mentioned above that this computer sits in the closet. This location comes with some problems, however. For example, the closet does not have a dedicated power outlet: I have to run an extension cable from an outlet in the middle of the room and out from under the bed.

## Reasons why my website is offline

I can't express my website's uptime in a _magnitude of nines_[^aws59s] the same way that a public safety and emergency service would[^aws39s]. Having _"one nines"_ uptime is really bad, and I would like to find a more impressive metric to present to someone than `9%`.

Now that I think about it, _uptime_ may not be the right word to use. Let me instead paint a picture with a short and incomplete list of reasons for which my website is offline:

* My ISP had an outage.
* The power was out.
* I closed the sliding closet door too quickly, and it unplugged the router.
* My husband pushed a box of shoes under the bed, and it pushed the extension cable out of its plug.
* Instead of unplugging the router itself to turn it off and then on again, I flipped the switch on the power strip because I thought it would be less effort. I did not fully grasp that this turns off _everything_ in the closet. Consequently, I forgot to turn _closet computer_ back on.
* I can’t get into the `bios` to configure _closet computer_ to turn back on when the power returns.
* _closet computer_ came up before the router and `autossh` stopped trying to open the `ssh` tunnels to the _bastion server_ after a period of time, because `systemd` is designed to give up and stop trying.
* My router refuses to route any traffic whatsoever until I point DNS away from the `pi-hole` on _closet computer_.
* Hubris.

### A hosting solution held together by glue and popsicle sticks

I don’t have a static IP. My ISP’s network infrastructure simply doesn’t allow for it, as the infrastructure is based on a mesh of wireless signal that is propagated between large, unsightly antennas installed on top of each house using their service. To get around this, I swallow my pride and purchase, again, a cloud server from `linode`. I call this one my _bastion server_, and with it I have this sort of static IP address that I can configure my domains to point to.

As I still can't reach from the outside world into my closet, I have to perform a little magic...

Using `ssh`, I can open a [remote tunnel](https://www.ssh.com/academy/ssh/tunneling/example#remote-forwarding) from _closet computer_ to the _bastion server_ with this command:

```{.bash .numberLines}
ssh -R 10080:localhost:80 bastion.oflogan.xyz
```

This command binds a connection to the _bastion server_'s `localhost:10080` port and forwards it to _closet computer_'s `*:80` port. You may notice that I'm using port `10080` rather than `80`, and this is because I am using a non-priveleged user to open the tunnel, as only `root` may bind to ports under `1024`.

Because the remote tunnel binds `localhost:10080` on the _bastion server_, this means that the tunnel is not accessible publicly, and that my website is still in the closet. In order for my website to come out of the closet, I use an `nginx` [reverse proxy](https://docs.nginx.com/nginx/admin-guide/load-balancer/tcp-udp-load-balancer/) to forward the _bastion server_'s public port `*:80` to `localhost:10080`. With this final connection made, the world can reach out and see my website coming out of the closet.

I can set up reverse proxies for each port, `22`, `80`, and `443`, on the _bastion server_ into reverse tunnels from _closet computer_'s `22`, `80`, and `443` ports to the _bastion server_'s `10022`, `10080`, and `10443` ports. I then leverage `autossh` to maintain persistent remote tunnels from _closet computer_ to the _bastion server_.

**This is how it looks:**

```{.txt .nowrap}
         +-BASTION-SERVER--------------\
         | <=proxy=> | <===remote tunnel===> +-CLOSET-COMPUTER---\
SSH ---> | *:22----> | localhost:10022 | --> | *:22 SSH          |
HTTP --> | *:80----> | localhost:10080 | --> | *:80 goto 443 duh |
HTTPS -> | *:443---> | localhost:10443 | --> | *:443 my website  |
         `-----------+-----------------+     `-------------------+
```

It’s ok that I’m using a cloud server for this, I tell myself, because none of my data is stored on it. But also, I think I pulled off something kinda cool.

This is a really fragile architecture, however. If the `ssh` tunnels aren’t up, the site can’t be served. I can’t even get an error page out as there doesn't seem to be a way to query for whether the reverse proxies are passing any data through. This setup is really kinda crap like that.

Being what it is, `systemd` on _closet computer_ will stop `autossh` from retrying connections to the _bastion server_ after a certain period. This is a major problem if there’s a hiccup in the home internet connection. Sometimes my internet is out for five minutes, but that’s a tradeoff made when using a cheaper, locally-based ISP over Comcast's cartoon villain bundled contracts. At a minimum, I need to use something other than `systemd` to manage my `ssh` tunnels, as it hasn’t been reliable _at all_ and is probably one of the bigger points of failure because it cascades with any failures at the router.

Physically, the closet receives too much human traffic for _closet computer_ or really any hardware to be held there. We live in a modest century Craftsman whose modern infrastructure has been monkey-patched together, and I piggybacked on the convenient, pre-existing connections made by the previous homeowners, putting no thought into potential consequences therein. Now we have a closet acting as a server cabinet instead of a closet. We are, of course, using it as both, with frequent access throughout the day, and cables running everywhere, as if a panicked octopus were dropped on the floor and frozen in place.

## Duct tape and rainbows

I have a lot of problems hosting my website and source code from my _closet computer_ in terms of uptime. Mostly related to tripping over the power cable, though in other ways the setup isn’t resilient. It’s been a really rewarding experience nonetheless. I’m working with tech that I use all the time as a consumer, but have never myself had to engineer or maintain. It's a fun puzzle.

Hosting my website and source code myself feels very empowering. I feel as if I reclaimed a little piece of the internet. For that alone I want to see what more I can self host. The internet today clusters around a centralizing oligopoly of services, and I pine for the days when every odd person had their own small, weird website, and frequented a handful of small, niche forums, each with just enough users to be interesting. Now there's Facebook. GitHub. Google... If I name a _thing_: there's likely a single site that I would go to for that _thing_. That _thing_ has likely regressed in quality towards the mean and mediocre. And when I go to that site for that _thing_, I can't tell whether I've found the _thing_ or an ad for a subscription service that gives me access to the _thing_ at a markup.

{{imageFigure id: "closet-computer-in-context",
      title: "Cables and all, the <em>closet computer</em> in context. Corey asked me to dust first.",
      src: "/images/reasons-my-website-is-offline/coming-out-of-the-closet-1024x768.png"}}

I severely took for granted the uptime that a third party service could give me. It’s making me rethink what the total cost of what third party hosting actually looks like, not only in terms of dollars and nebulous ideals, but more importantly in terms of peace of mind. Terms such as _will the promotion committee be able to see that writeup I referred them to?_ I would absolutely benefit from hosting my website from my _bastion server_ instead, but perhaps after I reread `linode`'s service agreements to allay my confounding mind.

Regarding nebulous ideals, Copilot makes a very strong case for treating [data as a form of labor](https://www.brookings.edu/blog/techtank/2018/02/21/should-we-treat-data-as-labor-lets-open-up-the-discussion/). At the time of this writing, Wed Dec 8, 2021, Copilot is in technical review and accessible via waitlist. I can't find any mention of licensing cost, or whether there will be a distinction between paid and community editions. If Copilot is available as an exclusively free service, then I believe it stands as an excellent reinvestment back into the community and profession as a whole. If someone is paying for Copilot, then we are collectively exploited within some legally-accepted gray area, because we agreed to be.[^tos]

## Stepping back

I still use GitHub professionally, and for open source projects hosted there. I realize that being a principled curmudgeon sometimes means that I’m only that. Obviously there also exists a sort of cognitive dissonance in regards to contributing to open source works in the context of labor. Regardless, I gave myself a learning opportunity, and I consequently don’t use GitHub as a personal consumer now. I suppose that was my goal after all.

I might give Copilot a go soon. I'm reading some good things about how it cuts through boilerplate like a hot knife through butter. Reviews such as this help me conceptualize how it might fit into my own workflow, and that its role is more augmentation than replacement. Especially now that I've had time to step back. I would be over the moon to have an intelligent assist for the boring parts of my job. I suppose that was GitHub's goal, too.

On the lighter side, being professionally out and telling people that my website is coming out of the closet gives me a certain subversive glee that I never realized I needed. It's a great lead into how the computer got there, too.

~~For now, I host my source code at [bitsof.thisfieldwas.green](#). My website source repository is directly accessible from [keywordsalad/thisfieldwas.green](#).~~

[^archbtw]: Meme reference: [btw i use arch](https://knowyourmeme.com/memes/btw-i-use-arch) -- When Vlad proofread this post, he informed me that it was customary to let the reader know that one uses Arch, btw.

[^aws59s]: [AWS Public Sector Blog: Achieving "five nines" in the cloud for justice and public safety](https://aws.amazon.com/blogs/publicsector/achieving-five-nines-cloud-justice-public-safety/)

[^aws39s]: Amazon Web Services experienced a significant outage on Tue Dec 7, 2021, and it lasted some [seven hours and thirty minutes](https://www.cnbc.com/2021/12/07/amazon-web-services-outage-causes-issues-at-disney-netflix-coinbase.html), taking out large swaths of services and even preventing people from doing their jobs. This means that for 2021, AWS's US-East-1 region had an uptime of no better than 99.9%, or _"three nines"_. I wish for my home-spun home hosting solution to one day perform at least as well as AWS does on their bad days.

[^tos]: There exists an episode of South Park where Kyle accepts the Apple iTunes user agreement [without reading it](https://www.imdb.com/title/tt1884035/). It's worth a watch but the writers employ a heavy-handed storytelling device to get their point across. I liked it, but you may not.
