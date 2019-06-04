# research notes

Contains unsorted notes created during research. Before diving into this, I suggest reading the reports.

## early_summary.md

Extensive summary of an early research stage (before discovery of the vulnerabilities and implementation
of the covert channel). Although this summary is outdated/incomplete, it is worth reading - it covers some very
basic topics of Unfying protocol (RF and USB layers).
The document was used in an exchange with otheer researchers interested in this topic, before the respective
github repos wen private. Because of this, the links to this repos are dead.

## covert_channel_proto.md

Analysis and first version protocol design for Unifying covert channel.

## channel_hopping_extracted_lookup_table.txt

Raw data from lookup tables of channel hopping functions, extracted during analysis of CU0007 dongle
firmware RQR12.01.B0019. Not used for tools.

## analysis_of_manually_demodulated_esb_frame_with_ack.txt

Manual analysis of demodulated ESB frames. Not very useful, when I did this, it wasn't clear to me how
the back channel in Enhanced Shockburst is working. To make a long story short: Acknowledgement frames
could carry payloads and look exactly like default ESB frames. PRX and PTX change roles for a short time
window and an ESB frame is send back to serve as ACK (empty payload). In case the PRX wants to communicate
something back, the ACK payload isn't empty. For nRF24 series, the TX FIFO of the PRX has to be filled with
the data for the ACK payload **before** the actual ESB frame arrives, otherwise an empty ACK will be sent.
This means, if higher protocol layers want the PRX to send something on-demand, the PTX has to poll constantly
in order to receive respective ACK payloads. Unifying does exactly this: Example if the Unifying dongle (PRX)
wants to send an outbound LED report to a keyboard device (PTX), this device hast to poll on ESB layer, 
otherwise it wouldn't be able to receive the ESB ACK payload carrying the LED report payload.

## encrypted_keyboard_report_type_determination.md

This documents holds some fragments, where behavior of two encrypted Unifying keyboards have been analysed 
for information leakage related to encrypted keyboard reports. The content is closely related to PoC 2 and
the respective report.

Short explanation:
Encrypted keyboard reports use a (kind of) AES128 CTR encryption. Without regard to existing weaknesses in 
the actual crypto implementation, it should be pretty clear that known plain text would be beneficial for
an attack. As pointed out above, a encrypted keyboard device (PTX) needs to poll outbound reports from the
dongle (PRX) on ESB layer, in case it wants to receive LED reports. For Logitech Unifying the closest
mechanism to this kind of polling are so called "keep alive frames" sent by the device.
As there is a trade off between constantly sending keep-alives and low power consumption, the delay between
successive keep-alives is dynamic. Dynmaic menas, there exists an additional type of RF frames, which allows
an device to advertise the intended keep-alive interval, which I called "set keep-alive".

Now for the interesting part: All keyboard devices I tested use a short keep-alive interval, after a key has 
been pressed (mostly 8ms) and toggle back to a long keep-alive interval, once all keys are released, again.
This means, when a key is pressed, the encrypted key frame is followed by frequent keep-alives. If an encrypted
keyboard report is followed by one or more "set keep-alive" frames and less frequent "keep-alive" frames, it is
a key release report.

To make a long story short, by capturing only encrypted keyboard reports and "set keep-alive" reports from RF
an attacker is able to distinguish key presses an key release. For every key release, we have known plaintext.
This is because the plaintext of a keyboard report is build like this:

- byte 0:		HID modifier byte
- byte 1..6:	HID keys 1 to 6 (0x00 means not pressed)
- byte 7:		always 0xC9 (used to sort out invalid frames after decryption)

We ultimately end up with a known plaintext of `0x00 0x00 0x00 0x00 0x00 0x00 0x00 0xc9` for every identified
key-release.

Marc Newlin already reported an issue, which allowed counter reuse for encrypted keyboards, back in 2016.
The issue was patched. For an unpatched **dongle** (the dongle decides if a report wis malformed counter is consumed, 
the device has no influence on this) it is enough to identify a single key release report, using the outlined 
technique. An attacker ultimately could resend this release report as often as he likes, while modifying the 
plain content to his needs (simple XOR keying of the encrypted part, thus the cipher for the specific counter is
known when the plaintext is known).

As mentioned, this is patched, thus I started to look for possibilities to gather information on the plain tet
of the encrypted "key down" frames. Luckily, the aforementioned "LED output" reports are sent over RF in unencrypted 
fashion. This means, if a key like CAPS LOCK, SCROLL LOCK or NUM LOCK is pressed, an attacker would spot the 
respective LED reports on air, which give him insights on preceding encrypted key down reports.

My test made clear, that the patch against counter re-use is weak, as only about 23 counters are cached on the dongle.
This means RF replay attacks are still possible, if the capture keyboard report sequence contains more than 23
frames with valid counters and encryption. With the knowledge up to this point, it is already possible to put 
arbitrary data into half of this frames, as we have known plaintext for every key release (assuming the whole 
sequence does not contain key combinations like `key1_down, key2_down, key_release_all`, which could easily filtered
out during sniffing). Now if somebody presses CAPS LOCK about 12 times on the keyboard under test, the received
key sequence should look something like this (without kee-alives):

1) encrypted key frame
2) encrypted key frame
3) one or more set keep-alive frames (means last encrypted key frame was a key release)
3) LED report with CAPS led changed (means encrypted key frame before the one with key release contained CAPS HID keycode)
4) encrypted key frame
5) encrypted key frame
6) one or more set keep-alive frames (means last encrypted key frame was a key release)
7) LED report with CAPS led changed (means encrypted key frame before the one with key release contained CAPS HID keycode)

... and so on ...

So in this case, we have known plaintext for all encrypted keyboard frames (same goes for NUM/SCROLL LOCK LEDs, of course).

The attack demoed in PoC2 utilizes exactly this. It requires a sequence of multiple key presses, each toggling
a LED. This means physical access is needed, unless the real user accidentally hits such an uncommon key 
sequence. Once such a sequence is recorded, all payloads could be modified to the need of the the attacker,
while keeping counters of successive encrypted frames intact (XORing encrypted part with known plaintext, followed 
XORing with new payload). The sequence could be replayed, as often as needed, as the counter cache is overflowed every time.
**This means, one-time capture of a single sequence of 23 encrypted reports fulfilling the described 
conditions, allows arbitrary keystroke injectiion. No matter if the device or the dongle is power cycled or 
different counters are in use. It actually works, till keys are re-generated, which only happens during 
pairing**   

There was an idea to overcome the need of physical access, with a brute-force approach. This could be 
accomplished by XORing the estimated plaintext on every encrypted key-down frame (incremented by 1 each time
the sequence is replayed), until an LED output report is received for the modified report. This, again, would
lead to known plaintext. This is because we know the new plaintext after report decryption was a key toggling an LED.
Let's say we produced an LED toggle of CAPS LOC, this means our `bruteforce_value_byte` produced a
`CAPS_HID_keycode` after decryption. We get:

```
HID_code_of_real_key_down ^ bruteforce_value_byte = CAPS_HID_keycode
```

... the resulting plain key, of the unmodified report could now be calculated with:

```
HID_code_of_real_key_down = CAPS_HID_keycode ^ bruteforce_value_byte
```

If the all key-down reports of a encrypted sequence are modified with `bruteforce_value_byte` at the same 
time, it would need 256 attempts in worst case, till the plaintext for all encrypted frames is known (this
assumes only the byte for key1 needs to be manipulated for each report. This again means, the sequence mustn't
contain key combinations, I already described how to prevent this).

There have been multiple reason, for which I failed with the practical implementation of the brute-force approach:
1) The XOR brute-force leads to unintended input, which couldn't be determined upfront.
2) The USB descriptor of Unifying dongles allows sending keys with high logical values (not common for a 
default keyboard descriptor). These values include key codes for multimedia-keys and keys like **power down** 
or **system sleep**. In most of my tests, resulting plaintext in brute-force phase included **power down** 
presses, which rendered the attack useless (requires a bunch of additional logic to detect when the target went 
down)
3) The attack takes too long
4) It is hard to allign LED output reports to correct encrypted key frames, while frequent input is happening.
This is because the USB layer (dongle receives LED reports via USB from host OS) enqueues outbound reports, if
they arrive faster than being processed.

So now what's the actual document about ? LED output reports, which are needed by all described attacks, are
produced by the host OS, not by the dongle (the dongle relays them from USB to RF). Different OS-device 
combinations behave differently in terms of "When is the LED report produced". An LED toggle could either 
occur after the `key down` or the `key up`, but there is always a minimum of two encrypted key reports per single
LED report (key down and key up). Finding a generic description for a valid sequence of key reports and LED 
reports, would eliminate the need to account for host OS and "set keep-alive" frames. 

The document contains the relevant observations, to find such a generic description.

## flash_extracted_device_data.txt

Ra device extracted from flash of a CU0007 dongle running firmware RQR17.01.B0019.
The data has been extracted utilizing the flash read methods of the Logitech bootloader. As it turned out 
later, the bootloader mirrors in another flash page on attempts to read this data. This lead to some confusion
during analysis of the key generation algorithm (the offset of flash should contain key data, not the device data which 
is mirrored in). The data could be accessed using publicly documented HID++ 1.0 commands. The actual content
of the flash region is the result of pairing virtual devices (emulated pairing). This helped to get a sense 
of how data changes of RF payloads during pairing lead to data changes in the respective flash regions (with
known meaning, as they are related to the output of HID++ 1.0 commands documented in Unifying specs).

## analysis_of_mirrored_flash(fake)_and_conception_of_firmware_mod_for_mem_dump.txt

t.b.d. 

This patch would allow reading devices AES keys from dongles based on Nordic chips and thus isn't included, for
now.

## mouse_pairing_real_flash_and_ram_dumps.txt

t.b.d.

Contains raw output of `munifying` including memdumps using the vulnerability of PoC 3, which allows extraction
of all AES device keys from Unifying dongle with Texas Instruments chip and thus isn't included, for now.

## old_notes_on_unifying_reverse_approach_incomplete.txt

Raw (outdated) notes taken during inspection of firmware RQR12.01.B0019. This firmware has no single patch 
applied, which allowed some basic research. The notes are included because they contain pseudo code for
the "encryption algorithm" in use.

## pairing_live_capture_using_crazyradio_k400p.txt

As the name states. A full pairing sequence between CU0007 and Logitech K400+, captured with CrazyRadioPA
and a modified nrf-research-firmware for `mjackit`.

## pairing_phase_1_requests_from_various_devices.txt

Initial pairing requests from multiple devices, to analyse differences.

**Worth mentioning (not sure if I included it in the reports):** Each device send its current RF address with the
first pairing address. As the RF address contains the 4-byte base address of the dongle, this leaks the actual
dongle address of a device. New devices replace this address with a random address for each request.

A device send a pairing request on every channel, in case it is powered on but doesn't receive an ACK frame
from the respective dongle, after a second or so. Pairing requests are always sent to the global unique pairing
address **bb:0a:dc:a5:75** (already published by Marc Newlin). It is always worth monitoring this pairing address 
(as it is worth monitoring 2.405 GHz, as it seems devices without a dongle in range send on this channel anyways)

## pairing_sdr_capture_full_analysis_incomplete.txt

Partial analysis of data transmitted during pairing (outdated).

## sdr_analysis_led_output_report_ack_payload.txt

Manual analysis of demodulated SDR capture from early analysis phase.

## logitech_usb_hid_reports.md

Early analysis of some USB HID reports of different Unifying devices.

**Worth mentioning:** The section `Manual LED output via hidraw USB interface` shows how DJ reports
written to the USB "DJ" HID interface of a Unifying dongle allow to send arbitrary RF frames
(LED reports in the example). The covert channel for the remote shell utilizes HID++ reports, instead of DJ 
reports, though.

## unifying_pairing_usb_analysis.md

Analysis of USB traffic (HID++ 1.0 commands and notifications) during device pairing.
Setting a dongle into pairing mode and un-pairing of existing devices is re-implemented in `munifying` 

## xxxxxxx_memread.txt 

t.b.d.
    
 

