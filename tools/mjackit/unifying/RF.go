package unifying

/*
RF frame type

0x01	b0000 0001	Keyboard report
0x02	b0000 0010	Mouse report
0x03	b0000 0011  Media ??
0x04	b0000 0100	System Control report (Sleep, PowerDown, WakeUp)
0x0e	b0000 1110	LED report (host to device)


0x10	b0001 0000	HID++ short
0x11	b0001 0001	HID++ long

0xd3	b1101 0011	(0x13 after masking) Keyboard report encrypted

0x5f	b1001 1111	Pairing request, with follow up frames -> keep dongle alive (no channel hop)
0x1f	b0001 1111	Pairing response, with follow up frames --> keep device alive (active listening on current channel)

0x0f	b0000 1111	Pairing response, without follow up frames

0x40	b0100 0000	I'm alive notification
0x4f	b0100 0000	Set Keep Alive Notification / Pairing completed notification

0x08	b0000 1000	valid after masking with 0x1f, unknown type
0x15	b0000 1000	valid after masking with 0x1f, unknown type, only newer firmwares
0x16	b0000 1000	valid after masking with 0x1f, unknown type, only newer firmwares

 */


type RFFrameType int

const (
	FT_UNKNOWN RFFrameType = iota
	FT_NOT_LOGITECH
	FT_KEYBOARD
	FT_KEYBOARD_ENCRYPTED
	FT_MOUSE
	FT_SYSTEM_CONTROL
	FT_MEDIA
	FT_LED_REPORT

	FT_NOTIFICATION_KEEP_ALIVE
	FT_SET_KEEP_ALIVE
	FT_INVALID_CHKSM

	FT_PAIRING_REQ_PHASE1
	FT_PAIRING_RSP_PHASE1
	FT_PAIRING_ACK_PULL_PHASE1
	FT_PAIRING_REQ_PHASE2
	FT_PAIRING_RSP_PHASE2
	FT_PAIRING_ACK_PULL_PHASE2
	FT_PAIRING_REQ_PHASE3
	FT_PAIRING_RSP_PHASE3
	FT_PAIRING_ACK_PULL_PHASE3
)

func (t RFFrameType) String() string {
	switch t {
	case FT_UNKNOWN:
		return "UNKNOWN"
	case FT_NOT_LOGITECH:
		return "NOT LOGITECH"
	case FT_INVALID_CHKSM:
		return "INVALID CHECKSUM"
	case FT_KEYBOARD:
		return "UNENCRYPTED KEYBOARD REPORT"
	case FT_MOUSE:
		return "UNENCRYPTED MOUSE REPORT"
	case FT_MEDIA:
		return "UNENCRYPTED MEDIA KEY REPORT"
	case FT_KEYBOARD_ENCRYPTED:
		return "ENCRYPTED KEYBOARD KEY REPORT"
	case FT_PAIRING_REQ_PHASE1:
		return "Pairing request phase 1"
	case FT_PAIRING_REQ_PHASE2:
		return "Pairing request phase 2"
	case FT_PAIRING_REQ_PHASE3:
		return "Pairing request phase 3"
	case FT_PAIRING_RSP_PHASE1:
		return "Pairing response phase 1"
	case FT_PAIRING_RSP_PHASE2:
		return "Pairing response phase 2"
	case FT_PAIRING_RSP_PHASE3:
		return "Pairing response phase 3"
	case FT_PAIRING_ACK_PULL_PHASE1:
		return "Pairing pull ack payload phase 1"
	case FT_PAIRING_ACK_PULL_PHASE2:
		return "Pairing pull ack payload phase 2"
	case FT_PAIRING_ACK_PULL_PHASE3:
		return "Pairing pull ack payload phase 3"
	case FT_NOTIFICATION_KEEP_ALIVE:
		return "NOTIFICATION KEEP ALIVE"
	case FT_SET_KEEP_ALIVE:
		return "SET KEEP ALIVE"
	case FT_LED_REPORT:
		return "LED REPORT"
	case FT_SYSTEM_CONTROL:
		return "UNENCRYPTED SYSTEM CONTROL REPORT"
	}

	return "No type string defined"
}

func (t RFFrameType) ShortString() string {
	switch t {
	case FT_INVALID_CHKSM:
		return "X"
	case FT_KEYBOARD:
		return "K"
	case FT_MOUSE:
		return "M"
	case FT_MEDIA:
		return "UNENCRYPTED MEDIA KEY REPORT"
	case FT_KEYBOARD_ENCRYPTED:
		return "k"
	case FT_NOTIFICATION_KEEP_ALIVE:
		return "A"
	case FT_SET_KEEP_ALIVE:
		return "S"
	case FT_LED_REPORT:
		return "L"
	default:
		return "?"
	}

	return "No type string defined"
}

func ClassifyRFFrame(pay []byte) (ftype RFFrameType) {
	l := len(pay)
	if l != 5 && l != 10 && l != 22 {
		//fmt.Println("None Logitech frame")
		return FT_NOT_LOGITECH
	}

	//validate Logitech checksum
	chksm_given := pay[len(pay)-1]
	LogitechChecksum(pay)
	chksm_calculated := pay[len(pay)-1]
	if chksm_given != chksm_calculated {
		//fmt.Printf("Chksm invalid given %#02x calculated %#02x: % #02x\n", chksm_given, chksm_calculated, pay)
		return FT_INVALID_CHKSM
	}

	//dev id is pay[0]
	rfTypeByte := pay[1]

	switch {
	case rfTypeByte == 0x40 && l == 5:
		return FT_NOTIFICATION_KEEP_ALIVE
	case rfTypeByte == 0x4f && l == 10:
		//		if pay[2] == 0x06 {
		//			return FT_PAIRING_REQ_PHASE3
		//		}
		return FT_SET_KEEP_ALIVE
	case rfTypeByte == 0x5f && l == 22:
		if pay[2] == 0x01 {
			return FT_PAIRING_REQ_PHASE1
		}
		if pay[2] == 0x02 {
			return FT_PAIRING_REQ_PHASE2
		}
		if pay[2] == 0x03 {
			return FT_PAIRING_REQ_PHASE3
		}
	case rfTypeByte == 0x1f && l == 22:
		if pay[2] == 0x01 {
			return FT_PAIRING_RSP_PHASE1
		}
		if pay[2] == 0x02 {
			return FT_PAIRING_RSP_PHASE2
		}
	case rfTypeByte == 0x0f && l == 10 && pay[2] == 0x06 && pay[3] == 0x02:
		return FT_PAIRING_RSP_PHASE3
	case rfTypeByte & 0x1f == 0x0E:
		return FT_LED_REPORT
	case rfTypeByte & 0x1f == 0x13 && l == 22:
		return FT_KEYBOARD_ENCRYPTED
	case rfTypeByte & 0x1f == 0x01:
		return FT_KEYBOARD
	case rfTypeByte & 0x1f == 0x02:
		return FT_MOUSE
	case rfTypeByte & 0x1f == 0x03:
		return FT_MEDIA
	case rfTypeByte & 0x1f == 0x04:
		return FT_SYSTEM_CONTROL
	}

	return FT_UNKNOWN
}

