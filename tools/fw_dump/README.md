# Script to Dump firmware from Unifying dongle with Nordic chip (f.e. CU0007)

Along with the infamous `nrf-research-firmware` Marc Newlin (Bastille) released some tooling
to flash CU0007 Unifying dongles with customized firmware.

The tools could be found here: https://github.com/BastilleResearch/nrf-research-firmware/tree/master/prog/usb-flasher

`logitech-usb-dump.py` is a modified version of the logitech-usb-flash script, which dumps
the flash, instead of writing it. It depends on the `unifying.py` module from the aforementioned repo.

## Important note

In contrast to the unreleased firmware patch, this method utilizes the dongle's bootloader to
dump the firmware. The Unifying Nordic bootloader has logic in place, which hinders dumping of
keys. The tool allows extraction of the whole flash region, but there will be no valid key data
in the dumps. For questions on this, please open an issue on github. 