
# Unifying disclosure repo

*This repository was accessed by a restricted group of reviewers before beeing opened to public (including Logitech staff).
The content is mostly left untouched. Most subfolders contain a dedicated README file*

This repo will be used to discuss recent vulnerabilities in Logitech
Unifying technology, as well to share and discuss related 
proof-of-concept code.

I will build up the repo, as I have time to. In contrast to common way 
of responsible disclosure, I like to bring this to attention for a 
closed community first and **give the vendor the chance to review 
everything**.

As Logitech already agreed that I'm allowed to disclose most parts, I'll
add in material as soon as I cleaned it up. This includes:


**For questions / discussion please utilize github issues, for direct communication
use my email address**

*I try to keep this transparent, by offering repo access to Logitech staff. Respective emails
are unfortunately unanswered, so far.*


## Updates

### July 9, 2019
- press "ZDNet": https://www.zdnet.com/article/logitech-wireless-usb-dongles-vulnerable-to-new-hijacking-flaws/
- new vulnerability: Hidden pairing mode of presentation clickers allows key extraction via RF sniffnig (CVE-2019-13052, will not be patched), instead of dumping from dongle (CVE-2019-13054, affects only dongles based on Texas Instruments chip, will be patched): https://twitter.com/mame82/status/1148600800685502469

### July 8, 2019
- press "Heise": https://www.heise.de/ct/artikel/Logitech-keyboards-and-mice-vulnerable-to-extensive-cyber-attacks-4464533.html

### June 28, 2019
- encrypted Logitech presenters affected by PoC 3 (key extraction from dongle - not disclosed, yet)
- Logitech R500: https://twitter.com/mame82/status/1143093313924452353
- Logitech SPOTLIGHT: https://youtu.be/BIs2gApDoDk

### June 26, 2019
- pushed statements from Logitech concerning the talk on the topic into talk subfolder
(including my counter statements, document is German)
- invitation for closed LOGITacker testing group sent (currently covers PoC 1 and other stuff, like forced
pairing etc.)

### June 25, 2019
- handling of all vulnerabilities in scope of this repository has been handed-over to German CERT-Bund


## (planned) Repo content

- mjackit (Go code to interact with Logitech Unifying dongles/devices from RF end and carry out PoC1, PoC2 and
run the demoed covert channel)
- munifying (Go code to interact with Unifying dongle on USB end, able to carry out PoC3, will be added as soon as 
Logitech has a patch ready)
- raw documents created during info gathering/analysis phase (without further comments, partially outdated, feel free 
to ask questions)
- the reports provided to Logitech
- Client agent code for remote shell relayed through Unifying receiver

## Not in repo

- recent work on LOGITacker (brings most of this to a nRF52840 dongle, acting as stand-alone device, exception see issue #2)
- raw SDR captures
- modified firmwares / project files from firmware analysis

## The new vulns

### 1) PoC1 - Sniff pairing and recreate AES keys for a Unifying device, in order to live decrypt keyboard RF traffic (CVE-2019-13052)

- PoC video: https://youtu.be/1UEc8K_vwJo
- Demo shown in the video is part of `mjackit` (in `tools` folder of this repo)
- Applies to all Unifying dongles and latest keyboard devices (and some others)
- will not be patched, because of limitations of available Logitech hardware and backwards compatibility requirement
of Unifying technology
- covered in ["vulnerability report 1"](https://github.com/mame82/UnifyingVulnsDisclosureRepo/raw/master/vulnerability_reports/report1_git.pdf)

### 2) PoC2 - Keystroke injection for encrypted devices, without knowledge of keys + bypass of counter reuse mitigation (CVE-2019-13053)

- PoC video: https://youtu.be/EksyCO0DzYs
- Demo shown in the video is part of `mjackit` (in `tools` folder of this repo)
- Applies to all Unifying dongles and latest keyboard devices
- Exploits a combination of known plaintext (requires one time physical access to device) and improper counter re-use 
protection (even if patches for counter re-use vulnerability, reported by Bastille are applied)
- will not be patched, because of limitations of available Logitech hardware and backwards compatibility requirement
of Unifying technology
- there exists a theoretical attack, which doesn't require physical access to a device - my attempts to implement a
proper PoC have failed (works too slow, leads to unpredictable and unintended input on target, reference: https://twitter.com/mame82/status/1117248244478894080)
- covered in ["vulnerability report 2"](https://github.com/mame82/UnifyingVulnsDisclosureRepo/raw/master/vulnerability_reports/report2_git.pdf)

*Note: In feedback for press releases, I was asked why this attack needs physical access, at all. In its nature, the crypto implementaion is vulnerable to known-plaintext attacks. "Plaintext" in this context means: pressed keys. A potential attacker needs about 12 known key-presses to attack inject arbitrary keystrokes without knowledge of the encryption keys (in fact, known plaintext for 24 encrypted reports is needed, but 12 out of 24 are key-releases in most cases). Of course, it doesn't matter how an potential attacker gets knowledge of the twelve pressed keys. Anyways, the Proof-of-Concept for the vulnerability was extended (as highlighted in "vulnerability report 2"), in order to showcase that the communication protocol could leak additional information via RF, ultimately leading to known plaintext. This happens, if the user presses a key which toggles a keyboard LED (CAPS, SCROLL, NUM) and ultimately an unencrypted LED report is sent over RF. An automated attack could be deployed on top of this, which could derive plaintext of twelve successive presses to LED togelling keys. As it is unlikely that a normal user presses such a key 12-times in a sequence (which must not be interrupted by non-LED-togelling keys, in order to get a continuous counter seqeunce), Report 2 and the respective Proof-of-Concepts state that an attacker needs physical access to press the "magic key sequence" once. This does not apply if the keys could be obtained in another fashion (f.e. watching a presentaion, where an affected clicker is used). In addition there exists a theoretical bruteforce approach, which allows to get known plaintext for unknown key presses (described in report 2). Approaches to implement a reliable PoC for the bruteforce failed due to different reasons (mostly because PowerDown keys where send during bruteforce attempts). If such a bruteforce succeeds (RF only, no physical access, but aggressive interaction with target host) the former unknown key-presses of the encrypted reports are known to the attacker (bruteforce of plaintext, without key knowledge). There are some slides in the talk hosted in this repo, which visualize the approach - a picture says more than 1000 words.*


### 3) PoC3 - AES key extraction from Unifying dongles with one-time physical access 

- PoC video: https://twitter.com/mame82/status/1101635558701436928 (high res: https://youtu.be/5z_PEZ5PyeA)
- Demo shown in the video is part of `munifying`, which will be added to this repo as soon as the respective Logitech 
patch is issued
- once keys are extracted `mjackit` could be used to eavesdrop all devices or inject keystrokes
- covered in "vulnerability report 3" (will be added, as soon as the Logitech patch is available)

### Remote shell

- PoC video: https://twitter.com/mame82/status/1104044796761595904 (high res: https://youtu.be/OBqvcAbMkRk)
- not considered as vulnerabilities - utilizes features of Unifying, not bugs
- server side is part of `mjackit` (in `tools` folder of this repo)
- deployment of client side payload, utilizing PoC2 is part of `mjackit` (toy demo, typing out the payload 
takes about 2 minutes, if a single character goes wrong - everything fails ... inteded way for client agent deployment 
is a down&exec stage) - PoC video of toy version: https://twitter.com/mame82/status/1128392333165256706

### Misc

- forced pairing as published by Bastille is part of `mjackit` (video: https://twitter.com/i/status/1124767990300409856)
- forced pairing prank (removed from code): https://twitter.com/mame82/status/1086253615168344069
- `mjackit` could emulate a dongle in pairing mode. This was mainly used in research phase, but could possibly be
used to pair non-Unifying Logitech devices to Unifying dongles (mentioned in one of the reports)

## ORIGINAL CONTENT DIRECTED TO EARLY REVIEWERS: Twitter invitation message for this repo (for those who missed it)

```
Hi folks,

as most of you know, I did some research in the area of security of wireless input devices manufactured by Logitech. 
The core technology is called "Unifying", but some of the already known vulnerabilities apply to non-Unifying products 
or products from other vendors.

The initial idea was to improve available tooling for exploitation of MouseJack-like vulnerabilities (see public 
Material from Bastille and SySS for reference), in order get those tools working reliable enough to be used for live 
demos in awareness talks/trainings.
While gathering information on the specifics of the Unifying protocol implementation, it became obvious that there was 
room for more than improvements of known exploits.

A new idea ivolved: Utilizing an unmodified Unifying receiver to work as relay between USB HID and RF for a covert 
channel driving a remote shell (unprivileged user, target OS Windows, no modification of Unifying dongle or firmware, 
latest patches applied to all Unifying hardware and target host).

It quickly turned out that implementing such a covert channel is possible, but as for all other covert channels I know, 
there is a constraint: A client agent has to be deployed, which is able to understand and translate the communication. 
Targeting wireless input devices, keystroke injection (typing out the client side agent - or at least some download&exec 
cradle) seemed to be the obvious way to go. Unfortunately related vulnerabilities had been patched by Logitech back in 
2016. I didn't want to gave up at this point and moved over the research to private area (in my limited spare time ... 
lack of time is a huge constrained for me). The result of this "digging deeper idea":

1) A discovered vulnerability in AES key exchange/generation, which allows a passive remote attacker (monitoring RF) to 
steal the crypto keys if a pairing is sniffed. Ultimately the attacker is able to decrypt all keystrokes monitoring RF 
traffic. Such a pairing has to be initialized by a user, but there are several ways to extend the attack in order to 
fulfill this requirement (DoS the device till the user re-pairs, attacker gets physical access to dongle+device and 
re-pairs hisself etc.). Of course, knowledge of encryption keys allows keystroke injection, too.

2) A discovered weakness in crypto implementation. Crypto relies on a "kind of" AES CTR implementation. This is worth 
nothing, if the implementation allows counter reuse an attacker is able to gain knowledge of plain text. Luckily, such 
an attack was already presented by Bastille and ultimately patched by Logitech. Unfortunately the mitigation approach 
still had issues and I was able to improve the idea to make it exploitable again. An attacker requires one time physical 
access to a keyboard device, to generate the known plaintext (about 10 to twenty key presses, which only take some 
seconds), while the automated RF equipment generates enough material to exploit the crypto implementation. This kind of 
physical access is needed exactly once, afterwards the attacker could inject as many keystrokes as he likes, as often as 
he likes. This requires no knowledge of encryption keys.
There exists a theoretical attack, which works without physical access at all. Unfortunately I wasn't able to come up 
with a reliable PoC for the second approach (the failed attempt was documented in one of my tweets, without further 
explanations).

3) A discovered weakness in special versions of latest Unifying dongles, which allows extraction of all AES keys in less 
than a second, with physical access. Impact is the same as for (1), but there's no requirement to discover devices on 
air, anymore (a keystroke injection would be possible if the respective device even isn't in range of the dongle or 
disabled).

And finally:
4) The initial goal to develop a protocol stack which relays a shell through a Unifying dongle has been achieved.

Vulnerabilities 1 to 3 have been reported to Logitech (in fact they got in touch with me before I had the first report 
ready). We walked together to half of the "responsible disclosure" process (although I had some concerns, I didn't 
involve one of the existing disclosure or bounty programs to handle this ... beside the failed attempt to use h1 
submission form provided by Logitech). The other half of the disclosure process - the actual disclosure - hasn't 
happened yet.

And exactly this is, why you are reading this:

As mentioned, bringing up the time to work on this is a big issue for me. Anyways, I submitted for exactly one 
conference CFP. This was BlackHat USA and yesterday I received the information that the talk isn't accepted. I'm totally 
fine with this (and not really surprised ... the submission was quickly hacked together some hours before CFP closed and 
the content isn't "fresh"). So I'm back to my main problem:

How to bring the new issues to broad attention in security community and make customers aware, without investing more of 
my limited time. I decided to take a less common approach: offering the material to security researchers, tool developers 
and other folks doing work in this specific domain.

As stated, my intention is to bring the problem to broad attention in a responsible way. Additionally I could only rely 
on folks, which are able to deal with the technical parts of the material. The reason for this is, that most material 
hasn't reached release state, as I'm again suffer from lack of time, but most of you should be able to bring it into 
production.

So what is the material I'm talking about:
- modified version of Bastille's nrf-research-firmware for nRF24LU1+ based USB dongles (fixes some minor issues, 
required for attacks)
- "mjackit" Go based tool to carry out all mentioned attacks utilizing a CrazyRadio PA (missing: proper documentation, 
proper CLI, code cleaning to get rid of experimental code)
- "mjackit" additionally includes the server side code for the covert channel, code for keystroke injection with respect 
to target keyboard language layout, code for live decryption of RF keyboard frames, some protocol details in comments
- "munifying" Go based tool to interface with a Unifying dongle via USB on Linux. Allows dumping of key and device data 
for affected dongles, unpairing of paired devices, putting dongle into pairing mode (same issues as with other tools... 
no CLI, no documentation etc.)
- ClientAgent for covert channel (.NET code written in C# and meant to be deployed as in-memory PowerShell payload, 
tested against Win10 and Win7)
- raw notes: SDR based protocol analysis, USB based protocol analysis, analysis of pairing from both ends, several 
outdated notes
- the notes include a document, with assembler code for a mod of the very first CU0007 firmware. This patch adds 
functionality which allows dumping of IRAM, external RAM, Flash and SFRs - at runtime - using a modified HID++1.0 
command (pure HID, no high privileges). I developed this small patch in order to dump AES keys for further analysis. 
I haven't taken a final decision, if I really gonna include this, as it would allow key extraction from all dongles with 
nordic chip, which aren't patched to a signed DFU bootloader. On the other hand, the patch doesn't add value to the 
research content.
- the writeups for the 3 vulnerabilities reported to Logitech

Currently, I'm working on a new tool which incorporates most of the aforementioned functionality, thus I put no further 
effort into development of the provided tools/docs.

What do I want from you?

Money!!! ... Just joking. I mentioned it several times, I want you to bring this to broad attention, to avoid spending 
more time on this myself. This could be achieved in several ways:
- Creating videos explaining the issues
- incorporation of new approaches into existing tooling
- public blog posts/papers covering the content
- maybe in another way, I currently can't think of

So please let me know if and how you are willing to support this and I will prepare a closed github repo (or repos) 
during next weeks, to grant you access to this content.

Of course this requirement doesn't apply to Logitech, as they should get access to all of this.

Note 1: I can't grant you access to "munifying" before the respective patch is released by Logitech (should happen 
during next weeks)

Note 2: This is for Logitech: I grant access to all mentioned content, as soon as the repos are ready. Please get in 
touch if there are any implications. I don't plan to give (public) talks on the topic in near future.


Regards

MaMe82
```


