package unifying

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/gousb"
	"time"
)

const NRF24_DEFAULT_TIMOUT = time.Millisecond * 2500

type NRF24 struct {
	ctx         *gousb.Context
	device      *gousb.Device
	epOut       *gousb.OutEndpoint
	epIn        *gousb.InEndpoint
	channels    []byte
	channel_idx int

	debug bool
}

func NewNRF24() (res *NRF24, err error) {
	res = &NRF24{
		ctx:       gousb.NewContext(),
		channels:  make([]byte, 26),
		debug: true,
	}

	for i := range res.channels {
		res.channels[i] = byte(i*3 + 2) // channel 8..71 (based on observations of pings to channel hopping Unifying dongle)
	}

	res.ctx.Debug(5)

	res.device, err = res.ctx.OpenDeviceWithVIDPID(0x1915, 0x0102)
	if err != nil {
		return
	}

	if res.device == nil {
		return nil, errors.New("NRF24 device not found")
	}

	// reset device
	res.device.Reset()

	res.device.SetAutoDetach(true)

	config, err := res.device.Config(1)
	if err != nil {
		return
	}

	// claim interface (idx 0, alt 0)
	iface, err := config.Interface(0, 0)
	if err != nil {
		return
	}

	res.epIn, err = iface.InEndpoint(1)
	if err != nil {
		return
	}
	res.epOut, err = iface.OutEndpoint(1)
	if err != nil {
		return
	}

	//fmt.Printf("%+v\n", res.device.Desc.Configs[1].Interfaces[0].AltSettings[0].Endpoints[0x81])
	fmt.Printf("EP In %+v\n", res.epIn)
	fmt.Printf("EP Out %+v\n", res.epOut)

	return res, err
}

func (d *NRF24) Close() {
	d.device.Close()
	d.ctx.Close()
}

func (d *NRF24) SetDebug(enable bool) {
	d.debug = enable
}

func (d *NRF24) Read(buf []byte, timeout time.Duration) (n int, err error) {
	/*
	self.dongle.read(0x81, 64, timeout=Nrf24.usb_timeout)
	*/

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return d.epIn.ReadContext(ctx, buf)
}

func (d *NRF24) SendCommand(command NRF24_COMMAND, data []byte, timeout time.Duration) (err error) {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	dataRaw := []byte{byte(command)}
	dataRaw = append(dataRaw, data...)

	length := len(dataRaw)

	//fmt.Printf("Writing to dongle out EP 0x01: %+v\n", dataRaw)

	for length > 0 {
		n, err := d.epOut.WriteContext(ctx, dataRaw)
		if err != nil {
			return err
		}
		length -= n
		//fmt.Println("Written", n)
	}

	return nil
}

func (d *NRF24) SetChannel(channel byte) (err error) {
	if channel > 125 {
		channel = 125
	}

	d.SendCommand(SET_CHANNEL, []byte{channel}, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	//	fmt.Printf("Tuned to %d\n", channel)
	return err
}

func (d *NRF24) TransmitPayloadGeneric(payload []byte, addr Nrf24Addr) (err error) {
	data := []byte{byte(len(payload)), byte(len(addr))}
	data = append(data, payload...)
	data = append(data, addr...)

	d.SendCommand(TRANSMIT_PAYLOAD_GENERIC, data, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return
	}

	if buf[0] > 0 {
		return nil
	} else {
		return errors.New("Error in TransmitPayloadGeneric")
	}
}

func (d *NRF24) TransmitPayload(payload []byte, retransmitDelay byte, maxRetransmitCount byte) (ackPay []byte, err error) {
	ackPay = []byte{}
	data := []byte{byte(len(payload)), retransmitDelay, maxRetransmitCount}
	data = append(data, payload...)

	//fmt.Printf("TX data % d\n", data)

	d.SendCommand(TRANSMIT_PAYLOAD, data, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	_,err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return ackPay, err
	}

	//fmt.Printf("TX res % #x\n", buf)
	if buf[0] > 0 {
		if buf[0] > 1 {
			if buf[0] > 32 {
				return ackPay, errors.New(fmt.Sprintf("Weird USB response on TXPayload % x\n", buf))
			} else {
				ackPay = buf[1:buf[0]]
				if d.debug {
					fmt.Printf("TX send: % #x\n", payload)
					fmt.Printf("TX rcvd: % #x\n", ackPay)
				}
				return
			}

		}

		return ackPay,nil
	} else {
		return ackPay, errors.New("Error in TransmitPayload")
	}
}

func (d *NRF24) TransmitAckPayload(payload []byte) (err error) {
	data := []byte{byte(len(payload))}
	data = append(data, payload...)

	//fmt.Println("TX ack payload")

	d.SendCommand(TRANSMIT_ACK_PAYLOAD, data, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	//n, err := d.Read(buf, NRF24_DEFAULT_TIMOUT)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return err
	}

	//fmt.Printf("RxSeq buf (%d): % #x\n", n, buf[:n])

	if buf[0] > 0 {
		return nil
	} else {
		return errors.New("Error in TransmitAckPayload")
	}
}


func (d *NRF24) GetChannel() (ch byte, err error) {

	d.SendCommand(GET_CHANNEL, []byte{}, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	n, err := d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if n != 1 {
		return 0, errors.New("Error reading current channel")
	}

	return buf[0], nil
}

func (d *NRF24) ReceivePayload() (payload []byte, err error) {
	buf := make([]byte, 64)
	d.SendCommand(RECEIVE_PAYLOAD, buf, NRF24_DEFAULT_TIMOUT)

	n, err := d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return payload, err
	}

	return buf[0:n], nil
}

func (d *NRF24) EnterSnifferMode(address Nrf24Addr, enableAutoAck bool) (err error) {
	pay := []byte{0x00, byte(len(address))}
	if enableAutoAck {
			pay[0] = 0x01
	}
	pay = append(pay, address.Reverse()...)

	d.SendCommand(ENTER_SNIFFER_MODE, pay, NRF24_DEFAULT_TIMOUT)
	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)

	return err
}


func (d *NRF24) EnterPromiscuousMode() (err error) {
	err = d.SendCommand(ENTER_PROMISCUOUS_MODE, []byte{0x00}, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return
	}

	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return err
	}

	return nil
}

func (d *NRF24) EnterPromiscuousModeGeneric(prefix []byte, paylen byte, rate byte) (err error) {
	// len(prefix), rate in Mbps, payload length
	data := []byte{byte(len(prefix))}
	data = append(data, rate) // default 2 Mbps
	data = append(data, paylen) // default 32 byte (maximum for NRF24 payload)
	data = append(data, prefix...)


	err = d.SendCommand(ENTER_PROMISCUOUS_MODE_GENERIC, data, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return
	}

	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return err
	}

	return nil
}

/*
Enable Amplifier for CrazyRadio PA
 */
func (d *NRF24) EnableLNA() (err error) {
	err = d.SendCommand(ENABLE_LNA_PA, []byte{}, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return
	}

	buf := make([]byte, 64)
	_, err = d.Read(buf, NRF24_DEFAULT_TIMOUT)
	if err != nil {
		return err
	}

	return nil
}

func (d *NRF24) NextChannel() (err error) {
	d.channel_idx++
	if d.channel_idx >= len(d.channels) {
		d.channel_idx = 0
	}
	return d.SetChannel(d.channels[d.channel_idx])
}

func (d *NRF24) Scan(timeout time.Duration) (err error) {
	err = d.EnterPromiscuousMode()
	if err != nil {
		return
	}

	err = d.SetChannel(d.channels[d.channel_idx])
	if err != nil {
		return
	}

	ch, err := d.GetChannel()
	if err != nil {
		return err
	}
	fmt.Println("Channel set to:", ch)

	startTime := time.Now()
	lastSwitchTime := startTime

	channelSwitchTime := time.Millisecond * 100 //change channel after 100 ms
	for time.Since(startTime) < timeout {
		dt := time.Since(lastSwitchTime)
		if dt > channelSwitchTime {
			d.NextChannel()
			dt -= channelSwitchTime
			lastSwitchTime = time.Now()
			lastSwitchTime.Add(dt) //add overhead
		}

		// ReceivePayload uses the last channel which was set
		// in promiscuous mode it tries to determine valid packets by bit shifting raw rx data + CRC check
		payload, errP := d.ReceivePayload()
		if errP != nil {
			return errP
		}

		if len(payload) >= 5 {
			fmt.Printf("==========\nReceived (channel %d): % #X\n", d.channels[d.channel_idx], payload)
		}
	}

	fmt.Println("Scan ended")
	return err
}



func (d *NRF24) PingSweep(addr Nrf24Addr, retransmitDelay byte, retransmits byte) (channel byte, err error) {
	d.EnterSnifferMode(addr, false)
	for _,ch := range d.channels {
		d.SetChannel(ch)
		if _,err := d.TransmitPayload([]byte{ 0x0f, 0x0f, 0x0f, 0x0f }, retransmitDelay, retransmits); err == nil {
			return ch,nil
		}
	}

	return 0,errors.New("Device address not found with ping")
}


type Payload struct {
	Channel byte
	Timestamp time.Time
	Data []byte
}

func (p Payload) String() string {
	return fmt.Sprintf("Channel %d, data: % +x", p.Channel, p.Data)
}

func (d *NRF24) Scan3(dwell time.Duration, abortOnReceive bool) (payloads []Payload, err error) {
	err = d.EnterPromiscuousMode()
	if err != nil {
		return
	}

	payloads = make([]Payload,0)

	for _, ch := range d.channels {
		err = d.SetChannel(ch)
		if err != nil {
			return
		}
		//fmt.Println("Scanning on channel ", ch)
		channelNotScanned := true
		startTime := time.Now()
		for channelNotScanned || (time.Since(startTime) < dwell) {
			// ReceivePayload uses the last channel which was set
			// in promiscuous mode it tries to determine valid packets by bit shifting raw rx data + CRC check
			data, errP := d.ReceivePayload()
			if errP != nil {
				return payloads,errP
			}

			channelNotScanned = false

			if len(data) > 1 {
				payload := Payload{
					Data: data,
					Timestamp: time.Now(),
					Channel: ch,
				}

				fmt.Println(payload)
				payloads = append(payloads, payload)
				if abortOnReceive {
					return
				}
			}
		}
	}

	return
}


type NRF24_COMMAND byte

/*
#define TRANSMIT_PAYLOAD               0x04
#define ENTER_SNIFFER_MODE             0x05
#define ENTER_PROMISCUOUS_MODE         0x06
#define ENTER_TONE_TEST_MODE           0x07
#define TRANSMIT_ACK_PAYLOAD           0x08
#define SET_CHANNEL                    0x09
#define GET_CHANNEL                    0x0A
#define ENABLE_LNA                     0x0B
#define TRANSMIT_PAYLOAD_GENERIC       0x0C
#define ENTER_PROMISCUOUS_MODE_GENERIC 0x0D
#define RECEIVE_PACKET                 0x12
#define LAUNCH_LOGITECH_BOOTLOADER     0xFE
#define LAUNCH_NORDIC_BOOTLOADER       0xFF
 */

const (
	TRANSMIT_PAYLOAD               = 0x04
	ENTER_SNIFFER_MODE             = 0x05
	ENTER_PROMISCUOUS_MODE         = 0x06
	ENTER_TONE_TEST_MODE           = 0x07
	TRANSMIT_ACK_PAYLOAD           = 0x08
	SET_CHANNEL                    = 0x09
	GET_CHANNEL                    = 0x0A
	ENABLE_LNA_PA                  = 0x0B
	TRANSMIT_PAYLOAD_GENERIC       = 0x0C
	ENTER_PROMISCUOUS_MODE_GENERIC = 0x0D
	RECEIVE_PAYLOAD                = 0x12
	//ENQUEUE_ACK_PAYLOAD            = 0x13
)

