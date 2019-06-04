# Unifying cover channel research

Hi *redacted*(no specific order ;-)),

as promised I grant access to my current state of investigations. First intention was to reimplement jackit in Go, for 
several reasons, but as discussed on Twitter I moved on to explore covert channel capabilities of the Unifying protocol.

Because of changing the goal in the middle of developing, my code ended up beeing chaotic as it contains a bunch of 
experiments, so I try to summarize my findings in this file.

This repo contains the code to interact with CrazyRadio PA (maybe other nRF24LU1+) with a modified nrf-research-firmware.

The firmware has been modified, because the one from Bastille Research has some issues in sniffing mode and could be 
found here:
https://github.com/mame82/nrf-research-firmware

For a covert channel, of course, not only the RF communication is interesting. For the USB communication, I started a
different repo, which could be found here:

https://github.com/mame82/munifying/blob/master/refs/unifying.txt

## covert channel considerations, investigations, protocol reversing

For a covert channel we need a back channel (downstream) from host perspective.

In a scenario, where the attacker works in RF range with a nRF24 or nRF51, the following conditions should be met:
- RF frames from (malicious) device to host (upstream) which could carry arbitrary data and end up as USB HID input 
report for the host are needed 
- HID output reports, from host to dongle (downstream) which could carry arbitrary data and end up as RF frames to
a non malicious device are needed. Those frames should be captured by the attacker (back channel), but not interfere 
with the functionality of the none malicious device

As we need to traverse from USB HID reports and back, first some details on the USB end of a Unifying dongle:
- the dongle presents 3 HID interfaces
- interface 1, In EP 0x81, uses 8 byte input reports described by common keyboard descriptor (1 byte modifier, 
6 bytes keys)
- interface 1 additionally defines LED output reports (5 bits used for LED, 3 bits constant, remaining 
bytes are padding) 
- interface 2, EP 0x82, uses 8 byte input/output reports described by common mouse descriptor (16 buttons in 2 byte 
bitfield, x/y axis with 12 bit resolution, z axis with 8 bit resolution, 1 byte input pan report ID 0x02)
- interface 2 defines additional input reports (ID 0x3 consumer control application:play, pause etc..., 
0x04 System Control: System sleep, System power down, System Wake Up)
- interface 3, In Ep 0x83, report len 32 bytes:
  - report ID 0x10, 6 bytes (0x00 to 0xff), Input and output --> HID++ short messages
  - report ID 0x11, 19 bytes (0x00 to 0xff), Input and output --> HID++ long messages
  - report ID 0x20, 14 bytes (0x00 to 0xff), Input and output --> DJ short messages
  - report ID 0x21, 31 bytes (0x00 to 0xff), Input and output --> DJ long messages

To understand why these interfaces exist, is recommend the following Logitech document:

https://lekensteyn.nl/files/logitech/Unifying_receiver_DJ_collection_specification_draft.pdf

In short words:
To support legacy keyboard/mice interfaces with common HID descriptors have to exist (for example controlling a
BIOS with keyboard, which has no Logitech userland code). As the dongle supports up to 6 devices, mapping everything
to a default HID keyboard/mouse is problematic, as the host wouldn't be able to distinguish the devices. Additionally
there would be no way to communicate to the devices (setting macros, request battery status etc. etc.) with the legacy
HID interfaces.

Logitech solves these two problems with the 3rd interface, which uses 2 vendor specific data exchange formats, each of
them in a long and a short version:
1) HID++ short/long for control communication between devices and dongles (reading / writing device registers, exchange
of event-like notifications)
2) DJ short/long to wrap legacy HID communication into a custom header, which includes a device ID and thus allows 
distinguishing between individual connected devices. The "DJ mode" has to be explicitly enabled, as device and dongle
have to agree on using this exchange format. The default (and fall back format) is HID++.

Additionally HID++ exists in several versions from HID++ 1.0 to HID++ 4.5 (or maybe higher).
For now, I focused onb HID++ 1.0, the respective Logitech draft (public part only) could be found here:
https://lekensteyn.nl/files/logitech/logitech_hidpp10_specification_for_Unifying_Receivers.pdf

Additional drafts on HID++ 2.0: 
https://lekensteyn.nl/files/logitech/

A very good learning resource on HID++ 1.0/2.0 is the Solaar source (especially when it comes to undocumented registers
or Wireless device IDs): 
https://github.com/pwr/Solaar/tree/master/lib/logitech_receiver

### recap

A quick recap:
- Interface 2 is barely usable for a covert channel as it doesn't support output data (but I had some fun sending
in System Control payloads, which shut down my Laptop when I hadn't saved my work)
- Interface 1 allows output reports for LEDs, but there's only a 8 bit field, from which only 5 bits are usable which,
again, all are mapped to real LEDs (2 of them are uncommon and shouldn't be present on most devices). Using the 3 
remaining bits defined as constant in the output report descriptor, wouldn't be reliable (if working at all), because 
they aren't defined as variable.

Interface 3 comes to help. As already stated, by default HID++ is used. **All my tests have been done on Kali Linux,
which seems to support DJ mode and thus DJ reports with ID 0x20/0x21 are working. I haven't progressed into 
investigation of Windows, yet (with and without Logitech software installed)** 


As pointed out in the 
[Unifying receiver software interface document](https://lekensteyn.nl/files/logitech/Unifying_receiver_DJ_collection_specification_draft.pdf).
raw HID reports could be wrapped into DJ reports:

Section 2.3.2ff describes the mapping between RF reports and USB reports with ID 0x20 (DJ short notifications).

A report to enable the CAPS Lock LED on a keyboard with device ID 0x03 would look like this:

```
0x20 0x03 0x0e 0x02 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00

# byte 0: report ID of DJ short report
# byte 1: device ID to address (only ends up as RF frame, if device is a keyboard or presenter)
# byte 2: USB report ID of LED output report
# byte 3: LED bitmask, 0x02 == CAPS LED on
# remaining bytes: padding to get to report length 15
```

The `send_usb.py` script is a test script to toggle the CAPS LED repeatedly. In order to make it work the correct device
ID between 0x01 and 0x6 has to be chosen and maybe a key has to be pressed to make the keyboard send RF frames and
collect ACKs (more on this later).

The same thing could have been achieved by writing a raw LED output report to the keyboard interface (EP0 with correct 
addressing or using raw HID device). The interesting thing about the DJ report format is, that the descriptor allows
logical **values between 0x00 and 0xff for those reports without violating the HID interface descriptor**, thus using 
the remaining 3 bits of the LED bitmask (byte 3) is perfectly possible (remember, the legacy keyboard LED output report
masks the unused bits as constant. The generic DJ reports, in contrast, can't do this).

The resulting RF frame would look something like this (see `sdr_analysis_led` for reference):

```
0x00 0x0e 0x02 0x00 0x00 0x00 0x00 0x00 0x00 0xf0

# byte 0: device Index (RF format)
# byte 1: USB report ID of LED output report
# byte 2: LED bitmask, 0x02 == CAPS LED on
# byte 3..8: Unused 
# byte 9: Logitech RF checksum
```

I don't want to dive to deep into RF payloads at this point, but some remarks:
- the device index isn't set (0x00), as the frame is already directed to the dedicated RF address of the keyboard 
(mapped from device index 0x03 to corresponding 5 byte RF address of paired keyboard)
- the device index of the dongle would have been 0xff
- the device index of the keyboard would have been 0x03 (in my setup), if used
- on RF end, the reports have to be pulled by the device (ACK payloads), this is why a key has to be pressed to wake up 
the keyboard

At this point, it is clear, that DJ reports allow a (very slow) outbound channel, in case a keyboard (or a presenter)
is paired to the dongle. The keyboard doesn't need to be connected or in range, in order to make the dongle produce
RF output reports (that's one of the reasons why I'm interested in the pairing process, if there's no keyboard, let's
just pair one).

Of course, in order to write raw HID reports a client side agent has to be used (deployment via keystroke injection ;-)).
As stated, I haven't started testing on Windows, but I assume the same techniques used by USaBUSe / P4wnP1 HID covert 
channel apply, to write raw reports to the HID interface on windows. The correct interface could be enumerated by
investigating the report length (32 byte in contrast to 8 byte on the other two interfaces).

The `NewUnifying()` function of `main.go` in the [munifying repo](https://github.com/mame82/munifying) does exactly
this.

### Sending more arbitrary outbound data

As mentioned, the Unifying protocols talk with devices utilizing HID++.
In HID++ 1.0 world, there has been no command to check the supported Unifying version of the devices, which has
been fixed beginning with HID++ 2.0 (see: https://lekensteyn.nl/files/logitech/logitech_hidpp_2.0_specification_draft_2012-06-04.pdf).

Each device involved has to understand a ping, based on a HID++ 1.0 short message (HID++ 2.0 onwards utilizes this
to determine the actual HID++ version, HID++ 1.0 would respond with an error). Thus I assume all devices support the
default HID++ 1.0 short message format.

In my tests, I could verify that HID++ 1.0 reports are send via RF in every case, if:
- the report ID is 0x10 (short report, didn't manage to get a long report out on Linux)
- the device Index used maps to a paired device (no matter if the device is connected)

A USB output report for a HID++ short message, which gets send to the aforementioned keyboard via RF, looks like this:

```
echo -ne "\x10\x03abcde" > /dev/hidraw0

# byte 0: report ID for HID++ short
# byte 1: device Index (paired keyboard)
# byte 2: !! HID++ command !!
# byte 3 .. 6: HID++ command parameters
``` 

The resulting RF reports look like this

```
Sniff 77:82:9a:07:f1: 0x00 0x10 0xf1 0x61 0x62 0x63 0x64 0x65 0x00 0x10
Sniff 77:82:9a:07:f1: 0x00 0x50 0xf1 0x8f 0x61 0x62 0x01 0x00 0x00 0x6c
Sniff 77:82:9a:07:f1: 0x00 0x40 0x01 0x16 0xa9 0x63 0x64 0x65 0x04 0x0c

I included the sniffed address, to pinpoint that the device ID used for the USB report (0x03)
is mapped to a dedicated RF address. The first 4 octets of the RF address are the dongle serial
(77:82:9a:07) the last octet is the device address (0xf1) which is assigned to the device during
pairing and used by both, the device and the dongle. The dongle listens on up to 6 addresses.
The direct address of the dongle would be: 77:82:9a:07:00

Report 1 (dongle to device):
byte 1: Report type (HID++ short)
byte 2: device index (RF)
byte 3: HID++ command / Message sub type (in this case: 0x61)
byte 4..7: HID++ command parameters

Report 2 (device to dongle):
byte 1: Report type: some notification (error ??)
byte 2: device index (RF)
byte 3: HID++ subtype: 0x8f == Error message (see HID++ 1.0 unifying specs)
byte 4: Sub ID of command with error (0x61 from previous report)
byte 5: Address producing the error (0x62, the first parameter send to the 0x61 command)
byte 6: Error code 0x01 == invalid command sub id

report 3:
not investigated, type 0x40 is used for various tasks (like pulling ACKs if send from dongle to device, setting keep 
alive intervals, sending keep alives ...). In this case it obviously is used to complete the error message from
report 2 and mark the end of a fragmented message. I'm not sure if the report moves from device to dongle, as at least
an empty packet would be needed in between (not forwarded by sniffer), but according to the content I assume it is a
device to dongle frame (has to be reinvestigated using an SDR if needed).

```

### recap

At this point we know, we could send up to 5 arbitrary bytes via RF to a device address which **is already paired with
the dongle**. This bytes have to be chosen carefully, as the HID++ message field (byte 3) corresponds to a command and
if the command accesses a device register, byte 4 corresponds to the register address.

It is worth mentioning, that pairing a custom device would be way more flexible, as the OS driver or dongle firmware
don't know if a HID++ command is valid, before transmitting it to the device (future proof protocol). Thus the device
has to answer the commands.

As there's a error reporting mechanism, valid commands could be enumerated for an already paired device via:
1) Client side payload sending valid/invalid HID++ messages and collecting resulting HID++ error messages
2) Malicous RF device, which sends the commands to the real device and collects error reports (there's a race condition
in response collection, because back communication is based on ACK payloads and we would interfere with the real 
dongle Logitech, more on this in RF sections)  

Anyways, unused HID++ commands (which have to be enumerated per paired device, as they could differ), are a more 
capable RF outbound channel. Sending invalid commands from USB to a device, results in outbound RF reports like shown
above. In case of an invalid command, the host receives an error message, which shouldn't be noticed by the end user, 
while the corresponding RF frames could be sniffed. For each "invalid" command, 4 arbitrary bytes in range 0x00..0xff
could be added.

Although the error messages (report 2 and 3 in the example), seem to be a nice inbound channel, I turned out that it is
possible to send in full **HID++ long reports** via RF, which could carry much larger payloads.

**At this point it should be clear, that there's definitely room for a bidirectional covert channel, which has to
be supported for a client side payload/agent**. As the USB interfaces are HID, I assume Windows grants unprivileged 
write access for arbitrary output reports.

Pairing a custom device opens additional capabilities:
- using of full range of commands of outbound HID++ short reports without issues (we don't care for valid/invalid 
commands, as there's no real device)
- no need to put additional markers into payloads in order to identify C2 traffic, as we work with a dedicated and
identifiable RF address (defined during pairing)
- possible valid outbound channels HID++ **long** messages, if a leegit looking device is paired (thinking about upload 
of keyboard makros to device, OTA of device firmware, HID++ 2.0 and greater)

So all in all, the pairing process is of great interest IMO.

Additionally, we should keep in mind, that Marc Newlin mentioned a forced pairing vulnerability for Unifying dongles 
with old firmware (which I assume could be exploited from RF end).

Additional note on none Unifying device:
I did some tests with a presenter, which couldn't be paired to a Unifying dongle, but accepts the same kind of keystroke
injection. Although "presenter" is a dedicated device type in Logitech specs, it works with standard keyboard reports.
What could be interesting about this observation ? Maybe similar techniques for outbound channel apply to none-Unifying 
devices, too.

### Pairing on USB end

Before moving on to RF, some quick notes on USB side of pairing.
The overall process is well understood, f.e. here:

https://lekensteyn.nl/logitech-unifying.html

or here:

https://github.com/pwr/Solaar

Additionally it should be noted, that dongle OTA updates are deployed using the USB HID interface (at least to boot 
the dongle in a mode with the firmware update bootloader).

Abstractly the pairing works like this:
1) USB out: Set register short (command 0x80), on dongle address (0xff), pairing register (0xb2) , open lock (0x01) to
enable pairing, for device number ??? determines last RF address octet ???, timeout to end pairing mode (default 30s) 
2) USB in: Lock changed notification (0x10 0xff 0x4a 0x01 ...)
3) USB in: Set register short response from dongle (0x10 0xff 0x80 0xb2)

4) USB in: device connection notification (0x10 deviceID 0x41 proto deviceInfo WirelessPID_LSB WirelessPID_MSB) 
<-- this indicates a connection, which could also occur when a device wakes up, in case it is a new device, the deviceID
field contains a new ID
4.1) Same notification again, but as DJ report instead of HID++
5) USB in: Lock close notification (0x10 0xff 0x4a ...), indicates end of pairing process

6) USB out: Get register long (command 0x83), on dongle address (0xff), pairing info register (0xb5) , get device name 
of index n (0x4n) 
7) USB in: Get register long response (0x11 0xff 0x83 0xb5 0x4n nameLength remaining_bytes_are_name_in_ascii)

8) USB out: Get register long (command 0x83), on dongle address (0xff), pairing info register (0xb5) , get extended 
pairing info of index n (0x3n) 
9) USB in: Get register long response (0x11 0xff 0x83 0xb5 0x3n serial0..serial3 powerSwitchPos ... unknown params)


10) USB in: second device connection notification (0x41), if first one hasn't link established set

11) USB out: special ping command to request HID++ version
12) USB in: Uknown command notification (if device is HID++ 1.0)
13) USB in: error response for ping command (invalid value)


A captured full USB pairing session looks like this (partially interpreted)
```
> 10ff83b5030000 //Short report, dev idx: 0xff, SubID: 0x83 (get long register req), register address: 0xb5 (pairing info), params: 0x03,0x00,0x00
< 11ff83b50377829a0717060a0000000000000000 //LongRegResp addr 0xb5, params: 0x03 ??, 0x77:0x82:0x9a:0x07 (dongle RF address), 0x17:0x06:0x0a ???

> 10ff81f1010000 //Get register 0xf1 (version info), params: 0x01, 0x00, 0x00 (dev idx 0x01) //First device, part 1
< 10ff81f1011201 //Get register resp, 0xf1 (version info), params: 0x01 (index), 0x12 (major), 0x01 (minor)

> 10ff81f1020000 //Get register 0xf1 (version info), params: 0x02, 0x00, 0x00 (dev idx 0x02) //First device part 2
< 10ff81f1020019 //Get register resp, 0xf1 (version info), params: 0x02 (index), 0x0019 (patch)

> 10ff81f1030000 //Get register 0xf1 (version info), params: 0x03, 0x00, 0x00 (dev idx 0x03)
< 10ff8f81f10300 //Error (0x8f; invalid value --> 0x03)

> 10ff81f1040000 //Get register 0xf1 (version info), params: 0x04, 0x00, 0x00 (dev idx 0x04)
< 10ff81f1040214 //???

> 10ff8100000000 //Get short register at address 0x00, params: 0x00,0x00,0x00
< 10ff8100000000 //

> 10ff8000000100 //Get short register at address 0x00, params: 0x00,0x01,0x00
< 10ff8000000000

> 10ff8102000000 //Get short register at address 0x02, params: 0x00,0x00,0x00
< 10ff8102000100 //Get Short reg resp, address 0x02, params: 0x00,0x01,0x00 //HID++ version ??

> 10ff8002020000 //Set register req, addr: 0x02, params: 0x02,0x00,0x00
< 10034104611120 //dev idx: 0x03, sub: 0x41 (device connection), r0: 0x04 (protocol type = unifying), r1: 0x61 (dev info: 0110 0001, keyboard, link encrypted, link not established, reason: packet without payload), r2: 0x11 (wireless PID LSB), r3: 0x20 ((wireless PID MSB)
< 10ff8002000000 //Set reg resp, addr: 0x02, params: 0x00, 0x00, 0x00 

> 10ff83b5420000 //Get long register req, addr: 0xb5 (Pairing info), Param: 0x42 (proto unifying+dev id ???),0x00,0x00
< 11ff83b542044b35323020202020202020202020 //Get long reg resp: addr: 0xb5, p0: 0x42 (proto unifying + dev id ???), p1: 0x04 (name length == 4), p2..p5: "K520"

> 10ff83b5320000
< 11ff83b5322a2236d81a40000002000000000000

> 10ff81f1010000
< 10ff81f1011201

> 10ff81f1020000
< 10ff81f1020019

> 10ff81f1030000
< 10ff8f81f10300 //102

> 10ff81f1040000
< 10ff81f1040214

> 10ff83b3000000 //Short report, dev idx: 0xff, SubID: 0x83 (get long register req), register address: 0xb3 (device activity), params: 0x00,0x00,0x00
< 11ff83b300000000000000000000000000000000 //Long report, dev idx: 0xff, SubID: 0x83 (long register resp), register address: 0xb3 (device activity), params: 0x00 * 16

... repeated 8 times


=========================
real pairing starts here
==========================

> 10ff80b201123c // enable pairing (p0: 0x01 - Open Lock, p1: 0x12 - Device Number ??, p2: 0x3c == 60 sec) //timeout of 0 would be default of 30sec
< 10ff4a01000000 //Notif Lock open (pairing on)
< 10ff80b2000000 //SetRegResp for enable pairing

> 10ff83b3000000
< 11ff83b300000000000000000000000000000000

... repeated multiple times

> 10ff83b3000000
< 11ff83b300000000000000000000000000000000

< 10014104421710 //Notif device connection: p0: 0x04 - protocol type unifying, p1: 0x42 - mouse, link not established, link not encrypted, packet without payload, p2: WPID LSB 0x17, p3: WPID MSB 10)
//Note WPID 1017 is MX Anywhere

> 10ff83b5400000 //Request device name
< 200141001710040000000000000000 //Short DJ packet, not HID++ (contains HID++ 0x41 Notif with WPID 1017)
< 10ff4a00000000 //Notif Lock closed (pairing off)
< 11ff83b5400b416e797768657265204d58000000 //Device name response (Anywhere MX)

> 10ff83b5300000 //Request extended pairing
< 11ff83b530cd6b955a0400000001000000000000 //Ext pairing response, Serial CD:6B:95:5A, Power Switch position 0x01 (on the base)

> 10ff83b3000000 
< 11ff83b300000000000000000000000000000000

> 10ff83b3000000
< 11ff83b300000000000000000000000000000000

< 10014104821710 //Notif, devID 0x01, Proto: Unifying, DevInfo: Mouse + packet with payload, WPID: 1017

> 10010012000000 //Ping to test for HID++ 2.0 support (see HID++ 2.0 draft)
< 10014b01000000 //Unknown notification 0x4b ??pairing success??
< 10018f00120100 //Error response for ping

> 1001810d000000 //Read reg 0x0d ?? of deviceID 0x01
< 1001810d5a6430

> 10ff83b3000000
< 11ff83b303000000000000000000000000000000

> 100181f1010000
< 100181f1011601

> 100181f1020000
< 100181f1020040

> 100181f1030000
< 100181f1030006

> 100181f1040000
< 100181f1040210

> 10ff81f1010000
< 10ff81f1011201

> 10ff81f1020000
< 10ff81f1020019

> 10ff81f1030000
< 10ff8f81f10300

> 10ff81f1040000
< 10ff81f1040214

> 10ff83b3000000 //Same polling request as before, but with new result
< 11ff83b307000000000000000000000000000000

... repeated multiple times ...

... till first mouse reports arrive ...

< 0200000320000000

```

All in all, pairing could be reduced to opening the Lock (with a timeout), wait for connect notifications followed by
lock closed notification and finally polling the device information.

My `munifying` repo contains a partial implementation of the process in the main method, but I have paused working on it, 
in order to do more reserach on RF end:
https://github.com/mame82/munifying/blob/master/main.go 

### RF - generic info

Unifying uses the Enhanced Shockburst (ESB) format by Nordic to communicate (known fact). There're various micro controllers
in use across different device-dongle combinations, but all are capable of communicating via ESB. This is even the case
if ESB isn't mentioned in the specs of a specific controller (like CC2544, which has way more options for packet format, 
but could be configured to align to ESB ... I believe this is the controller used in the CU0012 dongle). Even though, 
newer controllers support extended features (longer payloads, more data pipes == concurrent devices), the newest devices
should be compatible to the oldest Unifying dongles and vice versa.

Before getting into detail, let me share some of my pain with you. When started to investigate Unifying, I relied on
an nRF24LU1+ (CrazyRadio) and still do. This is great for sniffing/interacting with a known device address, but a pain
in this ass when it comes to device discovery or inspections of raw RF frames. Travis Goodspeed's work on pseudo 
promiscuous mode is awesome, but didn't help in all situations. I had the following questions on RF lay, when I started
- How is channel hopping done ? (I had BLE in mind and thought device and dongle take an agreement on how to hop channels,
which has been proven wrong ... reading specs could have helped in this)
- How is data communicated back from the PRX to the PTX (upstream traffic) ? I wasn't able to see any back 
communication from host to device, using CrazyRadio in Sniffing mode! (Turned out it was a mix of wrong assumptions
and an error in the nRF research firmware, which hinders sniffing multiple frames which occur with very short delays in
between)

No pain so far, but here it comes ... for a long time I considered to start with SDR (I had no real idea of RF comms
up to this point) and guess what, those questions forced me to buy an SDR (LimeSDR mini, which I got for only twice the
price it should have. If I could decide again, I would go for larger LimeSDR or a BladeRF).

Two days later, my new toy arrived (that's why the price doubled, didn't want to wait several weeks). I managed to get
the software/driver stack up in about a day and had to wait two more days, after learning SMA != RP-SMA and ordering new
antennas. It took me one more day, to learn the very basics of GNU radio and capture some ESB traffic for visual 
inspection. 

What I didn't know before: 

RF data exchange in 2.4GHz band with 2Mbps transfer rate, transmitted in bursts (no constant stream, as loved by GNU 
radio), spread across 125MHz bandwidth (theoretical sample rate of LimeSDR mini is roughly above 30MHz) ... these
are conditions/specs which don't lead to a real SDR beginners task. So I struggled for about two weeks, trying to convert 
what I could visually inspect (GFSK demodulated ESB frames) to a - more or less - realtime output of 1s and 0s. I tried 
to learn about Clock Recovery, when I identified it as my main problem, but yeah ... there seems to be no short path 
into protocol reversing, without getting in-depth knowledge on RF. 
I managed to bring up a setup, which could dump demodulated ESB bits (each bit represented by a byte 0x00 or 0x01) into 
a raw file, but this was a continuous stream, not reduced to the relevant bursts. 
Im aware of GNU blocks for nRF and various decoders, but I really thought there must be a quick way to bring
0s and 1s to the display during live capture. I haven't found an easy one and ended up analysing raw captures of
with 20 MHz sample rate with Inspectrum, which is great when it comes to (offline) decoding GFSK bursts into binary data.

... okay, enough background story, hope your feeling with me now.

Now for the relevant info on RF:

- the unifying Hardware (at leat the one inspected by myself) uses up to 26 channels with 3 MHz distance (a single nRF24 
channel at 2Mbps transfer rate consumes a 2MHz, even though the step size is 1 MHz)
- seems there're dongles using less than 26 channels (compare FCC data for CU0012, only 12 channels) 
- channels in use are 5,8,11,17..71,74,77 (3 MHz steps, starting at 2405 MHz). This observation aligns with FCC docs I 
wasn't aware of, at the point I analyzed the channels in use.
- Unifying RX/TX is always at 2Mbps
- The dongle acts as PRX, the devices as PTX
- The dongle listens for multiple devices (dedicated pipes, represented by different addresses)
- The device only uses one address at a given time
- **the dongle never transmits to a device on RF layer. TX from dongle to device occurs only in form of ACKs or ACKs
with payload (the device has to poll for upstream data)**

### RF - packet format

Unifying always uses Enhanced Shockburst (not legacy Shockburst), which adds dynamic payload length and automatic ACK
handling (latter is important).

A packet/RF frame is built like this

1) minimum 8 bit preamble of alternating 0 and 1. If the preamble ends with 1 or 0 depends on the next field, which is
the address
2) 40 address bits (5 byte address): The address corresponds to the sync field (exactly this is the problem with 
promiscuous mode). For Unifying, address bytes 0..3 consist of the dongle serial (like a MAC), byte 4 is a unique index
assigned to each device during pairing. The dongle **always** uses a device index of 0x00. According to Marc Newlin's
research I assumed device indices reach from 0x07 to 0x0c, which was wrong. A device index could have all values except
0x00, depending on the outcome of the pairing process. When a new device is paired, the device index is always 
incremented by one (even if the device was paired before). ** The dongle listens on the address of every paired device,
which means paired device addresses could be enumerated, even in absence of the device, as the dongle responds to 
received frames with an ACK for every listening address**
3) 9 bit header field (starting from MSB):
   1) first 6 bit: payload length, all packets spotted on Unifying have a payload length of 5, 10 or 22 bytes (could be
   used to verify Unifying traffic, as done by Jackit)
   2) 2 bit: PID, used as 2 bit sequence number, to distinguish ongoing traffic from retransmits
   3) 1 bit: No auto ack - if set to 1, the PRX doesn't send an ACK, for Unifying this is always 0 (PRX always sends ACK)
4) Payload, 0 to 32 byte, depending on the length in header field
   1) For Unifying, the last byte is always a checksum. The algorithm was published with Key Keriki v2 research and is
   implemented in the `LogitechChecksum()` function in `hid_logitech.go`
5) 2 byte CRC16 checked by ESB implementation to filter out noise on RF layer (Unifying has CRC always enabled, with 
16bit, although there's a custom 8 bit checksum in the payload)

If the nRF is used normally, only the raw payload is presented to/consumed from the host (or mcu). This means, we not
only use the CRC and header field, but also the address field. Thus, to eavesdrop RF traffic, the correct address to
listen on (sync field) has to be known upfront.

MouseJack solves this by using the pseudo promiscuous mode of Travis Goodspeed, which is capable of sniffing traffic
including the address bytes. This is done by (mis)configuring the Nordic chip to use a 2 byte address (read address 
length according to ESB spec has to be between 3 and 5 bytes) and setting a address of `00:aa` or `00:55` which
corresponds to a typical preamble prepended by a "low" signal state. Luckily the Nordic chip enables processing of
frames, right after it sees this fake-sync field, even if it is a preamble in reality. The absence of a real preamble
(frame start should look like `55:00:55` is those two byte are an address) isn't a problem, as the possible frame is 
only processed backwards up to the sync field, before being pushed to the higher layers. The rest is done by disabling
CRC check and sorting out noise on the MCU (moving CRC check to MCU to drop invalid packets). In result, one could sniff
ESB frames, including the address, which is crucial for real sniffing/injection.

This approach has several shortcomings:
1) A preamble isn't necessarily prepended by `0x00`, which is a condition which has to be fulfilled, using pseudo 
promiscuous mode. Goodspeed opted to use 0x00, as it occurs in noise pretty often. He manages to read up to about 50% 
of the frames with this approach, for a known channel.
2) Assuming the channel in use is known, one has to wait till the device transmits to the dongle, in order to receive a frame
or an ACK. The fact that devices could take a very long delay between transmissions of bursts (> 1 second when idle),
further decreases chances of sniffing a valid address. It should be noted, that it is much more likely to capture a 
frame from a Logitech mouse (8ms delay between successive frames when moving), than from a keyboard (keep alives occur
in low frequency)
3) As the address field is now part of the effective payload, full size packets couldn't be read completly (5 address
byte + 9 bit header + 32 byte payload + 2 byte CRC > 32 byte payload pushed to upper layer). This is less of a problem,
because we are only interested in the address in use and up to 22 payload bytes for Unifying. 

### recap

Focusing on Unifying, the facts up to this point could be used to improve implementations like `Jackit` in order to find
valid devices for keystroke injection.

**Assumption:**

We grabbed an address in promiscuous looking like `aa:bb:cc:dd:08`. If this would be a Logitech address, the now known 
dongle address is `aa:bb:cc:dd:00`. Possibly paired devices are in range `aa:bb:cc:dd:01..ff`.

We could first test, if the dongle itself is reachable on every relevant channel (5,8,1 ... 77) by sending a random
payload to this address. If the dongle is in range it would send back an ACK, which is indicated by a successful 
transmission on higher layer. So a simple ping sweep could be implemented, to:
1) Test if the dongle is in range (target for keystroke injection)
2) Find the current channel used by the dongle

Important: If no device is communicating with the dongle, the dongle starts hopping channels like this:
`5,14 ... 8,17 ... 11,20 ... 23,29 ...` (in 6 MHz steps, I haven't take notes on the hopping frequency, but it is fast).

3) because of this fact, we know that if a dongle stays on a channel for a longer time, a device is connected 
4) according to the public Logitech documents, the dongle tries to remain on a channel as long as possible. But we have
no guarantee. Thus if a transmitted (injected) payload doesn't get acknowledged, a new ping sweep should be started

5) once the channel of the dongle is known, the pingsweep could be repeated for all possible device addresses 
`aa:bb:cc:dd:01..ff`. Even if the respective device isn't in range (not connected according to dongle's internal state), 
the dongle would acknowledge a frame received on an address of a paired device. It has to be taken into account, that 
frames could get lost (proper retransmit setup) **and that the radios of PTX and PRX need a minimum of 130us to swap 
their roles + the time needed to transmit the ACK**

The last fact is important, when working with the nRF research firmware. The `TransmitPayload` function (emulating PTX)
mode, accepts two parameters. One is the `retransmit count` for automatic retransmission of frames which didn't receive 
an ACK, the other one is the retransmit delay (called timeout, which is misleading).

According to the specs, the PTX switches over to RX mode, for the period defined by retransmit delay (which is has to
be handed in as integer multiplier for a base unit of 150ms). There is an additional timeout, to abort the overall 
transmission, which isn't used by the research firmware (as far as I could see). There is an overall timeout for the
`TransmitPayload` method, which isn't consumed as parameter and is related to USB, not to RF layer (error timeout
if the USB request doesn't return a result in time. This mustn't be confused, as the transmission could still be 
ongoing, while the caller already returned an error due to USB response timeout).

6) At this point (assuming all timing parameters are chosen wisely) we are able to enumerate valid device addresses 
for a known dongle address. A dongle address could be derived from a address captured in promiscuous mode, which is
likely the device address of a mouse sending with short delay between bursts. Further assuming, that most dongles should
have a mouse **and a keyboard** paired, we only need to test if a device accepts keystrokes, to know if it is capable
of injection.

I implemented such a test, which is a bit naive, as it relies on the capability of injecting unencrypted keystrokes to
a device which otherwise communicates encrypted (a vulnerability presented by Bastille, my test dongle uses an outdated
firmware at version 012.001.0019 and is a CU0007).

The `LEDTest` function of `main.go` in this repo implements a test, to determine if a given device address a) is 
vulnerable to keystroke injection and b) communicates back LED reports. The function only succeeds if both conditions
are met and works like this:

1) Ping Sweep through all relevant channels, to determine on which channel the dongle is listening (based on ACK of
the dongle transmitted back to the device address)
2) Transmit a CAPSLOCK keydown report on the given channel, with a retransmit delay long enough to collect an ACK with 
payload (in every second iteration of the framing loop a keyup event is sent)
3) Read RX buffer (to check if an ACK payload has arrived during retransmit delay)
4) If an ACK payload arrived, check if it is an LED report (we don't check for CAPS, because everything else is unlikely)
5) Repeat step 1..4. If enough LED reports arrived, we could be sure that we have a injectable keyboard device address
(and hopefully the CAPS state is reverted to the initial one, if not ... who cares)
     
### More facts on ACK payloads

As pointed out, the only way to receive data from the dongle are ack payloads. A real device sends keep alives from time
to time, in order to collect them. From a PTX point of view (emulating a device), my first implementation for collecting
outbound USB reports to a specific device via RF, was to send empty ESB packets to the dongle in high frequency, form the
respective device address. This allowed me to read back ACKs, but interrupts outbound communication to the real device.
The reason is, that the real device is hardly able to receive data on the now polluted channel and that an ACK payload
isn't retransmitted once consumed by the PTX.

I chose this approach, because I wasn't able to capture payloads transmitted with ACKs passively, using the sniffer mode
of the research firmware. Reinspecting everything, I spotted a very rare amount of LED reports, when toggling CAPS on the
host keyboard while sniffing on the keyboard's RF address (roughly 5 percent, compared to the LED state changes on the
wireless keyboard). It turns out, that the research firmware flushed the RX queue on every read to the RX buffer (Which
happens far less frequently, than packets are arriving, iéven if the read loop has no delays. Remember, the delay 
between a finished TX and beginning of the respective ACK is about 130us). This could easily be fixed and it turns out
that the patched research firmware is able to receive bidirectional communications without issues. In fact about twice
as much packets could be spotted during sniffing now.

Changing the perspective to PRX (emulating a dongle), things get more complicated. Things are working like this:
- for a PRX, an ACK payload has to be loaded to the TX fifo **before a transmission from the PTX arrives**, otherwise
an empty ACK is send (this is relevant for the RF side of the pairing, as the device has to transmit packets to the dongle
in order to pull the ACK payloads, which contain the responses to requests of earlier transmissions).
- the research firmware provides a `TransmitAckPayload` method to achieve this. This method tries to  send an ACK, for
about 500ms (blocking method) and only succeeds without error, if the ACK has been sent. This of course could only happen
if a transmission has been received during the 500ms period. 

The `TransmitAckPayload` method is hard to use, in order to emulate a dongle in pairing mode, for multiple reasons.
The method flushes RX and TX fifo on every call. This is problematic if issued in a loop which alternates with the
method to read RX (`ReceivePayload`). According to the Nordic specs, the interrupt for a successful ACK transmission
doesn't toggle after sending the ACK. Instead it toggles when the next frame arrives from the PTX (if the ACK wouldn't
have been transmitted successfully, the PTX would retransmit the old frame again). This means, when `TransmitAckPayload`
returns success, more than one payload has been received. To know what this payload was, we have to read it **after the
ACK has already been send and a second payload arrived**. Even worse, the second payload wouldn't be ACK'ed if the 
`TransmitAckPayload` method isn't called again fast enough, as the method disables auto acknowledgments before returning
(assuming we are in passive sniffer mode, this is correct behaviour - otherwise the dongle and the sniffer, both would
send ACKs). Without going into further details, for proper dongle emulation, this firmware part has to be reimplemented
with a read-method like `ReceivePayloadAndAck(ackPayloadData []byte) error`.

I started implementing an emulation of a dongle in pairing mode, which works more or less robust (due to implications
of `TransmitAckPayload`). I haven't moved on to device emulation, yet (which should be easier, as ACK payloads could
be read right after TX). At time of this writing, the pairing works only with one of my devices, as I haven't fully
reversed all parts, yet, and there are some device specific parameters I need to investigate.

The aforementioned method could be found in `main.go` and is called `SimulatePairingDongle`.

The corresponding method `SimulatePairingDevice` isn't working, as it based on outdated assumptions.

The RF part of the pairing process will be described in the next section.

### RF - pairing

It took me the whole day to write up till here, thus this part is t.b.d.

As it is the most interesting one currently, let me give some rough details:
- when a dongle is set into pairing mode (Unlock, as described in the respective USB section), the RF listening address 
is changed to `bb:0a:dc::a5:75`. This global address is shared by all Unifying devices (as already presented by Marc 
Newlin)
- the dongle stays on this address, for the time defined in the repective HID++ command to unlock (30s if called with 
timeout 0x00, about 60 seconds if initiated from Unifying software)
- if the dongle doesn't receive RF frames from a device trying to pair, it hops channels
- a device, when turned on, tries to communicate to its paired address, if no ACKs are received from the dongle, it 
switches over to the gobal pairing address, starts channel hopping and sending frames on every channel, till an ACK
is received from a dongle in pairing mode (residing on this address)
- I forgot to measure, how long it takes for a device to change to pairing mode after enabling, but it should be less
than a second

Note: If a device is turned on, while the dongle isn't in range, an attacker could pair the device to a different dongle 
address (naive DoS). If an **old** devices is spotted, which tries to pair, it transmits its current address during
the pairing process, which could be used to determine a dongle address. This would happen, if a Unifying device is turned
on, while the dongle is out of range or not powered. For newer devices the "currently paired address" is chosen randomly.

- once the device in pairing mode receives an ACK from a dongle in pairing mode, both stop channel hopping and the 
pairing process starts
- the process involves several stages, but for now you have to live with the provided source of `SimulatePairingDongle`   


# Final notes

Don't get confused on the code structure, as there currently is no code structure. I started experimenting, putting code
in commenting code out, adding multiples structs (Go pendant of classes) with the same purpose, coded based on wrong 
assumptions, had frequent changes never committed ... so yeah:

- a usable file is nrf24.go, as it is the main interface to the CrazyRadio (which I'm using)
- other file contain varios predefined addresses and constants, mostly specific to my test devices

I tried to put all information into this summary

The modified nRF research firmware for CrazyRadio is here:
https://github.com/mame82/nrf-research-firmware

The code to interface with a Unifying dongle from Linux is here (needs libusb-1.0):
https://github.com/mame82/munifying


Add up: I forgot to mention mouse injection, injection of system control keys, but this is easy and basically done like
normal keystroke injection, but with different report IDs. I hvae some templates somewhere in the code and functionality 
to create proper mouse reports. Not much to say on this, one could move mice and shut down systems, but nothing new
(injection should be used to deploy client side agent, I'm not investigating this to end up with a "Shut Down Host" 
prank)