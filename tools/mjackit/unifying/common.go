package unifying

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

var (
	LogitechPairingAddr = Nrf24Addr{0xbb, 0x0a, 0xdc, 0xa5, 0x75}
)


type DeviceType byte

const (
	DEVICE_TYPE_UNKNOWN   DeviceType = 0x00
	DEVICE_TYPE_KEYBOARD  DeviceType = 0x01
	DEVICE_TYPE_MOUSE     DeviceType = 0x02
	DEVICE_TYPE_NUMPAD    DeviceType = 0x03
	DEVICE_TYPE_PRESENTER DeviceType = 0x04
	DEVICE_TYPE_TRACKBALL DeviceType = 0x08
	DEVICE_TYPE_TOUCHPAD  DeviceType = 0x09
)

func (t DeviceType) String() string {
	switch t {
	case DEVICE_TYPE_KEYBOARD:
		return "KEYBOARD"
	case DEVICE_TYPE_MOUSE:
		return "MOUSE"
	case DEVICE_TYPE_NUMPAD:
		return "NUMPAD"
	case DEVICE_TYPE_PRESENTER:
		return "PRESENTER"
	case DEVICE_TYPE_TRACKBALL:
		return "TRACKBALL"
	case DEVICE_TYPE_TOUCHPAD:
		return "TOUCHPAD"
	case DEVICE_TYPE_UNKNOWN:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("UNDEFINED DEVICE TYPE %02x", t)
	}
}

type UsabilityInfo byte

const (
	USABILITY_INFO_RESERVED                                    UsabilityInfo = 0x0
	USABILITY_INFO_PS_LOCATION_ON_THE_BASE                     UsabilityInfo = 0x1
	USABILITY_INFO_PS_LOCATION_ON_THE_TOP_CASE                 UsabilityInfo = 0x2
	USABILITY_INFO_PS_LOCATION_ON_THE_EDGE_OF_TOP_RIGHT_CORNER UsabilityInfo = 0x3
	USABILITY_INFO_PS_LOCATION_OTHER                           UsabilityInfo = 0x4
	USABILITY_INFO_PS_LOCATION_ON_THE_TOP_LEFT_CORNER          UsabilityInfo = 0x5
	USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_LEFT_CORNER       UsabilityInfo = 0x6
	USABILITY_INFO_PS_LOCATION_ON_THE_TOP_RIGHT_CORNER         UsabilityInfo = 0x7
	USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_RIGHT_CORNER      UsabilityInfo = 0x8
	USABILITY_INFO_PS_LOCATION_ON_THE_TOP_EDGE                 UsabilityInfo = 0x9
	USABILITY_INFO_PS_LOCATION_ON_THE_RIGHT_EDGE               UsabilityInfo = 0xa
	USABILITY_INFO_PS_LOCATION_ON_THE_LEFT_EDGE                UsabilityInfo = 0xb
	USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_EDGE              UsabilityInfo = 0xc
)

func (ui UsabilityInfo) String() string {
	ui &= 0xf
	switch ui {
	case USABILITY_INFO_RESERVED:
		return "reserved"
	case USABILITY_INFO_PS_LOCATION_ON_THE_BASE:
		return "power switch location on the base"
	case USABILITY_INFO_PS_LOCATION_ON_THE_TOP_CASE:
		return "power switch location on the top case"
	case USABILITY_INFO_PS_LOCATION_ON_THE_EDGE_OF_TOP_RIGHT_CORNER:
		return "power switch location on the edge of top right corner"
	case USABILITY_INFO_PS_LOCATION_OTHER:
		return "power switch location other"
	case USABILITY_INFO_PS_LOCATION_ON_THE_TOP_LEFT_CORNER:
		return "power switch location on the top left corner"
	case USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_LEFT_CORNER:
		return "power switch location on the bottom left corner"
	case USABILITY_INFO_PS_LOCATION_ON_THE_TOP_RIGHT_CORNER:
		return "power switch location on the top right corner"
	case USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_RIGHT_CORNER:
		return "power switch location on the bottom right corner"
	case USABILITY_INFO_PS_LOCATION_ON_THE_TOP_EDGE:
		return "power switch location on the top edge"
	case USABILITY_INFO_PS_LOCATION_ON_THE_RIGHT_EDGE:
		return "power switch location on the right edge"
	case USABILITY_INFO_PS_LOCATION_ON_THE_LEFT_EDGE:
		return "power switch location on the left edge"
	case USABILITY_INFO_PS_LOCATION_ON_THE_BOTTOM_EDGE:
		return "power switch location on the bottom edge"
	default:
		return "reserved"
	}
}

type ReportTypes uint32

const (
	REPORT_TYPES_KEYBOARD     ReportTypes = 1 << 1
	REPORT_TYPES_MOUSE        ReportTypes = 1 << 2
	REPORT_TYPES_MULTIMEDIA   ReportTypes = 1 << 3
	REPORT_TYPES_POWER_KEYS   ReportTypes = 1 << 4
	REPORT_TYPES_MEDIA_CENTER ReportTypes = 1 << 8
	REPORT_TYPES_KEYBOARD_LED ReportTypes = 1 << 14
	REPORT_TYPES_SHORT_HIDPP  ReportTypes = 1 << 16
	REPORT_TYPES_LONG_HIDPP   ReportTypes = 1 << 17
)

func (rt ReportTypes) String() string {
	res := "Report types: "
	if rt&REPORT_TYPES_KEYBOARD != 0 {
		res += "keyboard "
	}
	if rt&REPORT_TYPES_MOUSE != 0 {
		res += "mouse "
	}
	if rt&REPORT_TYPES_MULTIMEDIA != 0 {
		res += "multimedia "
	}
	if rt&REPORT_TYPES_POWER_KEYS != 0 {
		res += "power keys "
	}
	if rt&REPORT_TYPES_MEDIA_CENTER != 0 {
		res += "media center "
	}
	if rt&REPORT_TYPES_KEYBOARD_LED != 0 {
		res += "keyboard LEDs "
	}
	if rt&REPORT_TYPES_SHORT_HIDPP != 0 {
		res += "Short HID++ "
	}
	if rt&REPORT_TYPES_LONG_HIDPP != 0 {
		res += "Long HID++ "
	}
	return res
}

func (rt *ReportTypes) FromSlice(sl []byte) (err error) {
	if len(sl) < 4 {
		return errors.New("incorrect slice length, has to be 4")
	}
	ui32 := binary.LittleEndian.Uint32(sl)
	*rt = ReportTypes(ui32)
	return nil
}


type DongleInfo struct {
	NumConnectedDevices   byte
	WPID                  []byte
	FwMajor               byte
	FwMinor               byte
	FwBuild               uint16
	LikelyProto           byte
	BootloaderMajor		  byte
	BootloaderMinor		  byte

	Serial                []byte
}

func (di *DongleInfo) String() string {
	res := fmt.Sprintf("Dongle Info\n")
	res += fmt.Sprintf("-------------------------------------\n")
	res += fmt.Sprintf("\tFirmware (maj.minor.build):  RQR%02x.%02x.B%04x\n", di.FwMajor, di.FwMinor, di.FwBuild)
	res += fmt.Sprintf("\tBootloader (maj.minor):      %02x.%02x\n", di.BootloaderMajor, di.BootloaderMinor)
	res += fmt.Sprintf("\tWPID:                        %02x%02x\n", di.WPID[0], di.WPID[1])
	res += fmt.Sprintf("\t(likely) protocol:           %#02x\n", di.LikelyProto)
	res += fmt.Sprintf("\tSerial:                      %02x:%02x:%02x:%02x\n", di.Serial[0], di.Serial[1], di.Serial[2], di.Serial[3])
	res += fmt.Sprintf("\tConnected devices:           %d\n", di.NumConnectedDevices)


	return res
}

type DeviceInfo struct {
	DeviceIndex           byte
	DestinationID         byte
	DefaultReportInterval time.Duration
	WPID                  []byte
	DeviceType            DeviceType
	Serial                []byte
	RFAddr                []byte
	ReportTypes           ReportTypes
	UsabilityInfo         UsabilityInfo
	RawKeyData            []byte //applies on dongles with WPID 0x8808 (not 0x8802)
	Key                   []byte //derived from keydata

	Name string
}


func (di *DeviceInfo) String() string {
	res := fmt.Sprintf("Device Info for device index index %d\n", di.DeviceIndex)
	res += fmt.Sprintf("-------------------------------------\n")
	res += fmt.Sprintf("\tDestination ID:              %#02x\n", di.DestinationID)
	res += fmt.Sprintf("\tDefault report interval:     %v\n", di.DefaultReportInterval)
	res += fmt.Sprintf("\tWPID:                        %02x%02x\n", di.WPID[0], di.WPID[1])
	res += fmt.Sprintf("\tDevice type:                 %#02x (%s)\n", byte(di.DeviceType), di.DeviceType.String())
	res += fmt.Sprintf("\tSerial:                      %02x:%02x:%02x:%02x\n", di.Serial[0], di.Serial[1], di.Serial[2], di.Serial[3])
	res += fmt.Sprintf("\tReport types:                %08x (%s)\n", uint32(di.ReportTypes), di.ReportTypes.String())
	res += fmt.Sprintf("\tUsability Info:              %#02x (%s)\n", byte(di.UsabilityInfo), di.UsabilityInfo.String())
	res += fmt.Sprintf("\tName:                        %s\n", di.Name)
	res += fmt.Sprintf("\tRF address:                  %02x:%02x:%02x:%02x:%02x\n", di.RFAddr[0], di.RFAddr[1], di.RFAddr[2], di.RFAddr[3], di.RFAddr[4])
	//res += fmt.Sprintf("\tKeyData:                     % 02x\n", di.RawKeyData)

	if len(di.Key) > 0 {
		//res += fmt.Sprintf("\tKey:                         % 02x\n", di.Key)
		res += fmt.Sprintf("\tKey:                         % 02x **REDACTED**\n", di.Key[:3])
	} else {
		res += fmt.Sprintf("\tKey:                         none (no link encryption in use)\n")
	}

	return res
}

type SetInfo struct {
	Dongle DongleInfo
	ConnectedDevices []DeviceInfo
}

func (si *SetInfo) AddDevice(d DeviceInfo) {
	//update RF addresses
	d.RFAddr = append(si.Dongle.Serial, d.DestinationID)
	si.ConnectedDevices = append(si.ConnectedDevices, d)
}

func (si *SetInfo) String() (res string) {
	res = si.Dongle.String()
	for _,d := range si.ConnectedDevices {
		res += fmt.Sprintln()
		res += d.String()
	}
	return
}

func (si *SetInfo) StoreAutoname() (err error){
	filename := fmt.Sprintf("dongle_%02x_%02x_%02x_%02x.dat", si.Dongle.Serial[0], si.Dongle.Serial[1], si.Dongle.Serial[2], si.Dongle.Serial[3])
	err = si.Store(filename)
	if err == nil {
		fmt.Printf("Dongle data stored to file '%s'\n", filename)
	}
	return
}

func (si SetInfo) Store(filename string) (err error){
	j,eJ := json.Marshal(si)
	if eJ != nil {
		err = eJ
		return
	}

	err = ioutil.WriteFile(filename, j, 0644)

	return
}

func LoadSetInfoFromFile(filename string) (res *SetInfo, err error) {
	j,eJ := ioutil.ReadFile(filename)
	if eJ != nil {
		err = eJ
		return
	}

	res = &SetInfo{}
	err = json.Unmarshal(j, res)
	return
}