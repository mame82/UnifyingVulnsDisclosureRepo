#!/usr/bin/env python

from unifying import *


# Instantiate the dongle
dongle = unifying_dongle()




size = 16
send = "\x10\x70\x3f" + chr(size) + "\x00"*28
logging.info(':'.join("{:02X}".format(ord(c)) for c in send))
response = dongle.send_command(0x21, 0x09, 0x0200, 0x0000, send)

firmwaredata = []

for pos in range(0, 0x7fff, 16):
	addr = chr((pos & 0xff00) >> 8) + chr(pos & 0xff)
	#print(repr(addr))
	send = "\x10" + addr + chr(size) + "\x00"*28
	response = dongle.send_command(0x21, 0x09, 0x0200, 0x0000, send)
	
	result = response[4:20]
	firmwaredata += result
	rsphx = ':'.join("{:02X}".format(c) for c in result)
	addrstr = "0x%02x:\t" % (pos)
	print(addrstr + rsphx)

with open("full_flash_dump.dat","w+") as f:
	f.write("".join(map(chr,firmwaredata)))
	f.flush()

with open("flash_dump_fw_only.bin","w+") as f:
	f.write("".join(map(chr,firmwaredata[:0x67fe])))
	f.flush()
