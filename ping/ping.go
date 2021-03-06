package ping

import (
	"github.com/cooper/screenmgr/device"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

var devices []*device.Device

func startDeviceLoop(dev *device.Device) error {
	devices = append(devices, dev)
	go deviceLoop(dev)
	return nil
}

func deviceLoop(dev *device.Device) {
	p := fastping.NewPinger()
	p.MaxRTT = 5 * time.Second
	p.Network("udp")

	// resolve IP
	// TODO: ipv6 support
	ra, err := net.ResolveIPAddr("ip4:icmp", dev.Info.AddrString)
	if err != nil {
		dev.Warn("ping couldn't resolve the IP address")
		return
	}
	p.AddIPAddr(ra)

	// on receive, update last receive time
	lastTime := time.Now()
	started := false
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		dev.Debug("received ICMP packet: addr=%v, rtt=%v", addr, rtt)

		if !dev.Online || !started {
			dev.Log("device is online")
			dev.Online = true
			started = true
		}
		lastTime = time.Now()
	}

	// on idle, check if it's been a while
	p.OnIdle = func() {
		dev.Debug("ICMP idle")

		// it's been less than 10 seconds; no big deal
		if time.Since(lastTime) < 10*time.Second {
			return
		}

		if dev.Online || !started {
			dev.Log("device is offline")
			dev.Online = false
			started = true
		}
	}

	// do this continuously
	// TODO: if the device is removed or ping is disabled, stop the loop.
	p.RunLoop()
}

func init() {
	device.AddDeviceSetupCallback(startDeviceLoop)
}
