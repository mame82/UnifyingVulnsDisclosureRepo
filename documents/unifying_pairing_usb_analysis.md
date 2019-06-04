# Send 5 arbitrary bytes via RF to target device with ID 0x03 as short HID++ 1.0 message:
```
echo -ne "\x10\x03abcde" > /dev/hidraw0
```

Note: third byte 'a' is message SubID and could be interpreted as command, so a save value should be chosen for this
in order to avoid setting registers of the real device by mistake (f.e. 0x80 would be set short). As 0x00..0x7f is
reserved for notifications and commands start with values >0x80 keeping the MSB of the third byte 0 should be save.

Out

ID: 0x10
Size: 8
Count: 6
Logical: 0x00..0xff
Usage: 0xff01 (vendor)

ID: 0x11
Size: 8
Count: 19
Logical: 0x00..0xff
Usage: 0xff02 (vendor)

ID: 0x20
Size: 8
Count: 14
Logical: 0x00..0xff
Usage: 0xff41 (vendor)

ID: 0x21
Size: 8
Count: 31
Logical: 0x00..0xff
Usage: 0xff42 (vendor)

# Device pairing, USB communication

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






