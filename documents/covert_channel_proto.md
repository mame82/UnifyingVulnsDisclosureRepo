# Covert channel proto

Note: 

Example payloads are for a device with RF address xx:xx:xx:xx:4C which is the 3rd device paired to the dongle.
Thus, the device index on RF is 0x02, while the device index on USB is 0x03.
The examples don't have valid header bitmask fields (wasn't implemented at time of this writing).

## RF frame (ESB, needs valid device address, device has to support HID++)

```
byte num: 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 len: 22
------------------------------------------------------------------------------------
example:  02 11 4c bb 4d 69 63 72 6f 73 6f 66 74 20 57 69 6e 64 6f 77 73 25


00:     reserved - TX: unused, RX: destination Device index (0x00..0x05)
01:     reserved - RF report ID, dongle->device 0x11 (HID++ long), device->dongle 0x51 (HID++ long, with keep-alive)
02:     reserved - Destination ID (last octet of device RF address), has to be alligned to device in use (same for RX/TX)
03:     marker for covert channel comms, agreed between PTX/PRX, normal usage is HID++ SubID, a value which isn't a valid SubID 
        has to be chosen (no interference with device functionality, other than HID error reports).
        bit0 of marker indicates if control frame. Valid values: 0xba (payload frame) 0xbb (control frame), for now.
04:     header bitmask (see bitmask)
05..14: payload
15:     Logitech CRC
```

Note: 

Supporting multiple sub channels would consume an additional header byte, but we want to use a max payload size of 16 bytes for now, 
thus there's only one channel)

### RF frame bitmask

```
bit 0..1: SeqNo (up to 4 frames in flight with accumulative Ack, but prototype uses ping-pong approach)
bit 2..3: AckNo (up to 4 frames in flight with accumulative Ack, but prototype uses ping-pong approach)
bit 4..7: payload length for payload frame / control type for control frame
```

Note:
To save a header bit, valid payload length are 0 to 15. A frame with a 16 byte payload length is valid, too, but delivered as 
a control frame of control type `0x0` (max payload size). This is indicated by setting control type bits 4..7 to 0x0 and 
frame type bit (bit0 of marker) to on.

## USB output report (results in RF frame from dongle to device)

```
byte num: 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 len: 20
-------------------------------------------------------------------------------
example:  11 03 bb 4d 69 63 72 6f 73 6f 66 74 20 57 69 6e 64 6f 77 73


00:     USB report ID, 0x11 (HID++ long)
01:     Destination Device index (0x01..0x06, 0xff for dongle) see not 1 !!
02:     marker for covert channel comms, agreed between PTX/PRX, normal usage is HID++ SubID, a value which isn't a valid SubID has 
        to be chosen (no interference with device functionality, other than HID error reports). Value is fixed to 0xbb, for now.
03:     header bitmask (see RF frame bitmask)
04..13: payload
```

Note 1:

The device index on USB is incrememented by 1 compared to RF, as counting starts from 0x01 instead of 0x00 (see HID++ 1.0 specs).
For input reports, every device index should be accepted (valid payload covert channel message is recognized using the marker byte),
in order to allow the RF part to use any of the available RF addresses (which correspond to  a single device index on USB).
For output reports, the device index of the respective input report has to be used, in order to map the resulting RF frame to the
correct RF address.

## USB input report (result of received RF frame from covert channel, device to dongle)

Same as output report. Sequence number of respective input report is reflected back, as AckNo. USB reports which haven't got the
length of a HID++ long report (20 byte) or the proper marker field (0xbb) have to be ignored.

## Communication approach

A real device connected to the dongle is mimiced (using the same RF address) in order to establish the covert channel.
Some things have to be considered to achieve this, as this is a kind ov "invalid" setup. This is because we have
two PTX (real device and covert channel device) talking to a single PRX (dongle). This, again, involves problems when
it comes to the back channel from PRX to PTX, which is handled using acknowledgement RF frames with a payload.
We have to account for those problems.

### Fixation of dongle channel

In order to avoid that the dongle starts channel hopping, the covert channel device sends keep alive reports via RF. This is achieved
by setting bit 6 of the RF report type for transmitted frames (RF report type gets 0x51 instead of 0x11). On USB end, this bit is
stripped away, thus input reports still are of type 0x11.
Sending keep alives, doesn't assure the PRX (dongle) never changes the channel. In case the dongle finds the channel poluted, it
will hop to another one. Thus the communication channel has to be checked from time to time, which could be done based on acknowledgment
frames (not receiving Acks after multiple transmissions is a strong indicator for a channel change of PRX)

### Interference with real device

As the dongle is alweays acting as PRX, while a paired device is acting as PTX bi-directionaöl communication wouldn't be possible.
In order to compensate for this, a special mode of Enhanced Shockburst is used, namely Auto-Acknowledgement with Ack payloads.
This means, that PTX (device) and PRX (dongle) change there roles for a short amount of time, in order to allow the PRX to transmit back
an acknowledgement frame to the PTX (which switched to receive mode for a short amount of time, after a frame transmission).
In order to establish a mechanism for bi-directional communication, the acknowledgement frames from PRX to PTX could carry payloads
(ack payload, frame packet format is the same as for regular ESB frames, but in opposite direction). All of this is handled
by the Nordic chip and transparent to the MCU running the vendor code. 

From perspective of MCU of a device a transmission, could be:

- successful (Ack received, during short RX phase)
- not successfulf (no Ack received, even after retransmits)
- successful with payload (Ack with payload received)

Under normal circumstances, we don't have to take care of the chip's low level functionality of handling acknowledgment frames,
but because of our "invalid setup", we have to.

Obviously from PTX end, it is easy to detect if a transmission failed, for both the valid device and the covert channel device
(which clones this device). From perspective of the PRX (dongle), payload transmission is only possible using acknowledgments.
The dongle needs a way to assure, that a payload added to an acknowledgement frame has been received by the respective device,
somehow, in order to account for packet loss. Once more, this is handled by the chip in low level. To avoid interference with
the real device, as far as possible, it is crucial to understand the low level process.

The wireless chip in PRX mode automatically sends ack, if auto-ack is enabled (which is the case in this scenario). If the
TX FIFO of the PRX contains outbound data, this data is send along with the next acknowledgment for an arriving frame. This 
means the payload has to reside in the RX FIFO **before an RF frame arrives** (we don't care for this, but it should be known.
that multiple device to dongle transmissions could be needed to fetch an acknowledgement payload on PTX end, which was produce 
in response to a payload transmitted earlier). More important is to understand, how the dongle knows that an acknowledgment
payload was successfully transmitted back, in order to decide if this payload has to be send again or not (in contrast, 
retransmission from PTX to PRX are automated and handled by the chip). From dongle perspective, it looks like this: the chip
produces an interrupt, which indicates that an acknowledgment with payload was transmitted successfully. This interrupt fires,
when the following conditions are met:

1) The PRX (dongle) has transmitted the ack payload from TX FIFO via RR
2) A new RF frame (different content or different sequence id) arrived after the ack ha been send

To be clear: Two a second RF frame has to be received, in order to allow the PRX to detect if an ACK has been delivered. If
the conditions highlighted above are met, the dongle considers an ack payload as delivered and **won't send it again**.
If the dongle re-transmits acknowledgment payloads, at all, depends on the vendor implementation.

The arising problem for the covert channel gets more obvious, now:

Both, the real device and the covert channel device, consume ack payloads and send successive frames. In an ideal scenario,
the covert channel device would only consume ack payloads which carry "covert channel data" and the real device would consume
all other ack payloads. The problem could be described more easily with an example, so let's assume the dongle has a pending
ack payload, which is an outbound LED report destinated to the real device. Multiple things could happen now:

1) The covert channel device transmits a frame. Once the dongle receives the transmission, it switches to TX mode and sends back 
the LED report. The cover channel device toggled to RX mode for a short period, after transmission receives the respective ack 
payload (which was destinated to the real device). The real device hasn't tranmitted, thus it hasn't changed to RX mode and 
ultimately misses the ACK payload. The dongle, at this point in time, doesn't know if the ACK arrived, but as soon as either
the real device or the covert channel device would transmit a new RF frame, the dongle assumes the ack payload was delivered 
correctly. Even if we don't allow the covert channel device to send a successive frame to the dongle, the real device will do
so and as this frame differs from the previous one, the dongle's RF chip fires the interrupt for a successful ack transmission.
A possible (but not implemented) way  to deal with that, would be to change the RF mode of the covert channel device to PRX
and reque the received payload (destinated to real device) in the TX FIFO in order to send it as ack when the next TX from the
real device happens. Anyways, this setup involves new problems, because now there're two PRX trying to deliver ACKs.

2) The real device transmits a frame. It receives the ack payload in response and everything is fine.

The same issues apply, if the dongle has an outbound payload pending, which is destinated to the covert channel device, but
transmitted to the real device.

It should be noted, that on a higher level those transmissions to the wrong receiver are less problematic. If the real device
receives a "covert channel" payload, it will regard it as invalid, because the HID++ SubID is chosen in a way to be invalid.
There is a small interference, because the real device would generate an HID++ error, which is send back in response (consumes
channel capacity). If the covert channel device receives a wrong frame, it is simply discarded.

### Dealing with interference

To account for the outlined problems, the covert channel device stays in PRX mode. It only switchs over to PTX if something
has to be send. This means, changing from PTX to PRX and back is done manually. A transmission of a new frame is only started,
if an ACK payload with covert channel data has been received. This reduces transmissions to the minimum and thus reduces
consumption of ack payloads. Ack payloads with covert channel data are send back anyways, but in response to transmission
received from the real device. The real device considers those ack payloads invalid, but they are received by the covert 
channel device, too (which is acting as PRX, but doesn't send ACKs).

So the problem can't be solved fully, but could be reduced. Still, if the covert channel device transmitts, it could consume
enqueued outbound data from the dongle in form of an ack payload, which was destinated to the real device (which misses
this ack, as it is not in RX mode).

As a rule of thumb: The more transmission are done from the covert channel device, the higher is the likelyhood of fetching away
ack payloads destinated to the real device. Transmitting less frequently means reducing the covert channel bandwith.

Taking a look at the described issue from application layer, instead of RF level, makes it even less problematic: Situations
in which the dongle has to transmit data back to the device are rare. For pure HID traffic, this only applies to LED outbound
reports to keyboards. Wrongly consuming those reports results in missed LED toggles on the real device (if there are LEDs
at all). Otherwise, outbound traffic from dongle to device is only involved for HID++ communication. This for example happens
if the dongle requests battery status from a device or additional device information, which should be shown in a dedicate
software (not cached by the dongle itself). During normal operation, this doesn't happen to often.

For a prototype covert channel, we account to this problem only as much as needed. This means if a keyboard is able to type
and a mouse is able to act like a mouse, while covert cjhannel communication is happening on the same RF address, we are fine.


