package unifying

import (
	"errors"
	"fmt"
	"github.com/mame82/mjackit/helper"
)

type Nrf24Addr []byte

func (a Nrf24Addr) String() (res string) {
	if len(a) == 0 {
		return
	}
	for i, o := range a {
		if i > 0 {
			res += ":"
		}
		res += fmt.Sprintf("%02x", o)
	}
	return
}

func (a Nrf24Addr) Reverse() (res Nrf24Addr) {
	res = make([]byte, len(a))
	copy(res, a)
	for i := len(res)/2 - 1; i >= 0; i-- {
		opp := len(res) - 1 - i
		res[i], res[opp] = res[opp], res[i]
	}
	return
}

func ParseNrf24Addr(s string) (nrf24Addr Nrf24Addr, err error) {
	if len(s) < 14 {
		goto error
	}

	if s[2] == ':' || s[2] == '-' {
		if (len(s)+1)%3 != 0 {
			goto error
		}
		n := (len(s) + 1) / 3
		if n != 5 && n != 4 && n != 3 {
			goto error
		}
		nrf24Addr = make(Nrf24Addr, n)
		for x, i := 0, 0; i < n; i++ {
			var ok bool
			if nrf24Addr[i], ok = helper.Xtoi2(s[x:], s[2]); !ok {
				goto error
			}
			x += 3
		}
	} else {
		goto error
	}
	return nrf24Addr, nil

error:
	return nil, errors.New("invalid Nrf24 address")

}

type DeviceCaps int

const (
	DT_UNKNOWN DeviceType = iota
	DT_LOGITECH
)
const (
	DC_NONE DeviceCaps = 0
	DC_KEYBOARD  DeviceCaps = 1 << 0
	DC_MOUSE DeviceCaps = 1 << 1
	DC_MULTIMEDIA DeviceCaps = 1 << 2
	DC_POWER DeviceCaps = 1 << 3
	DC_MEDIA_CENTER DeviceCaps = 1 << 4
	DC_SHORT_HIDPP DeviceCaps = 1 << 5
	DC_LONG_HIDPP DeviceCaps = 1 << 6
)

type LogitechDevice struct {
	Address         Nrf24Addr
	Caps            DeviceCaps
	ChannelsUsed    []byte
	LastChannelUsed byte
}

