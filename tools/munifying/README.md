I agreed with Logitech not to disclose detailed information on CVE-2019-13054 / CVE-2019-13055 (USB based AES key extraction from Logitech wireless receivers).

As `munifying` source code contains relevant information, only a pre-compiled binary is linked for now.

The binary is compiled for Linux, x64 and depends on libusb-1.0.0.
Features related to CVE-2019-13054/13055 are disabled in this pre-release version.

Still, it could be used to reproduce other PoC like this one: https://twitter.com/mame82/status/1148600800685502469

# munifying pre-release repository

https://github.com/mame82/munifying_pre_release
