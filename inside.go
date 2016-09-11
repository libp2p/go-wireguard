package wireguard

import "log"
import "net"
import "runtime"

// TODO: be smarter about this
const mtu = 1500

func (f *Interface) readInsidePackets() {
	mtu := mtu
	skip := 0

	// OSX prepends the first 4 bytes received by each packet
	// with the values of AF_INET/AF_INET6 to indicate the
	// encapsulated IP packet. Unfortunately a Read()
	// must be for the entire packet content before fetching
	// subsequent packets, therefore we need to do a little
	// massaging ourselves.
	if runtime.GOOS == "darwin" {
		mtu += 4
		skip = 4
	}

	for {
		buf := make([]byte, mtu)
		log.Println("wip: f.inside.Read()\n")
		n, err := f.inside.Read(buf)
		log.Printf("wip: f.inside.Read() finished: (%d, %s)\n", n, err)
		if err != nil {
			// TODO: figure out what kind of errors can be returned
			// one would be unloading the TUN driver from underneath
			log.Printf("f.inside.Read() error: %s\n", err)
			continue
		}

		f.receiveInsidePacket(buf[skip:n])
	}
}

// extracts destination address from IPv4/IPv6 packet
func extractIP(buf []byte) (src net.IP, dst net.IP, err error) {
	ipVer := buf[0] >> 4

	if ipVer == 4 {
		src = net.IP(buf[12:16])
		dst = net.IP(buf[16:20])
	} else if ipVer == 6 {
		src = net.IP(buf[8:24])
		dst = net.IP(buf[24:40])
	} else {
		return src, dst, errInvalidIpPacket
	}

	return src, dst, nil
}

func (f *Interface) receiveInsidePacket(buf []byte) error {
	_, dst, err := extractIP(buf)
	if err != nil {
		return err
	}

	p, err := f.routetable.Lookup(dst)
	if err != nil {
		return err
	}

	if p == nil {
		// we need to generate ICMP unreachable message
		// but very tricky because we need to know our interface
		// IP address, and the whole part of construcing the ICMP
		// itself
	}

	return p.send(buf)
}
