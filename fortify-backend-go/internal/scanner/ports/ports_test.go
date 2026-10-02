package ports

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTop100LengthAndUnique(t *testing.T) {
	ports := PortsForProfile(ProfileTop100)
	if len(ports) != 100 {
		t.Fatalf("top100 has %d ports, want 100", len(ports))
	}
	seen := map[int]bool{}
	for _, p := range ports {
		if p < 1 || p > 65535 {
			t.Fatalf("top100 has out-of-range port %d", p)
		}
		if seen[p] {
			t.Fatalf("top100 has duplicate port %d", p)
		}
		seen[p] = true
	}
}

func TestTop1000LengthAndUnique(t *testing.T) {
	ports := PortsForProfile(ProfileTop1000)
	if len(ports) != 1000 {
		t.Fatalf("top1000 has %d ports, want 1000", len(ports))
	}
	seen := map[int]bool{}
	for _, p := range ports {
		if seen[p] {
			t.Fatalf("top1000 has duplicate port %d", p)
		}
		seen[p] = true
	}
	// top100 must be a prefix (popularity order preserved)
	top := PortsForProfile(ProfileTop100)
	for i, p := range top {
		if ports[i] != p {
			t.Fatalf("top1000[%d] = %d, want top100 prefix %d", i, ports[i], p)
		}
	}
}

func TestFullLength(t *testing.T) {
	ports := PortsForProfile(ProfileFull)
	if len(ports) != 65535 || ports[0] != 1 || ports[65534] != 65535 {
		t.Fatalf("full profile has %d ports, want 1..65535", len(ports))
	}
}

func TestServiceAndRisky(t *testing.T) {
	if ServiceName(80) != "http" || ServiceName(443) != "https" {
		t.Fatal("well-known service names wrong")
	}
	if ServiceName(99999) != "unknown" && ServiceName(61234) != "unknown" {
		t.Fatal("unknown port should map to \"unknown\"")
	}
	for _, p := range []int{21, 23, 445, 3389, 3306, 6379, 27017} {
		if !IsRisky(p) {
			t.Fatalf("port %d should be risky", p)
		}
	}
	if IsRisky(80) || IsRisky(443) {
		t.Fatal("http/https should not be risky")
	}
}

// TestScanLocalhost spins two real listeners (one with a banner, one silent)
// and asserts the scanner finds exactly those ports.
func TestScanLocalhost(t *testing.T) {
	bannerLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer bannerLn.Close()
	silentLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer silentLn.Close()
	go func() {
		for {
			c, err := bannerLn.Accept()
			if err != nil {
				return
			}
			_ = c.SetDeadline(time.Now().Add(2 * time.Second))
			fmt.Fprint(c, "SSH-2.0-FortifyTest\r\n")
			c.Close()
		}
	}()
	go func() {
		for {
			c, err := silentLn.Accept()
			if err != nil {
				return
			}
			time.Sleep(300 * time.Millisecond) // stay open, say nothing
			c.Close()
		}
	}()

	bannerPort := bannerLn.Addr().(*net.TCPAddr).Port
	silentPort := silentLn.Addr().(*net.TCPAddr).Port
	closedPort := silentPort + 1
	// make sure the "closed" port really is closed
	for {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", closedPort), 200*time.Millisecond)
		if err != nil {
			break
		}
		c.Close()
		closedPort++
	}

	open := ScanPorts("127.0.0.1", []int{bannerPort, silentPort, closedPort})
	if len(open) != 2 {
		t.Fatalf("found %d open ports, want 2: %+v", len(open), open)
	}
	byPort := map[int]OpenPort{}
	for _, o := range open {
		byPort[o.Port] = o
	}
	if got := byPort[bannerPort].Banner; got != "SSH-2.0-FortifyTest" {
		t.Fatalf("banner = %q, want SSH banner", got)
	}
	if _, ok := byPort[silentPort]; !ok {
		t.Fatal("silent open port missing")
	}
}

func TestResolveIPLiteral(t *testing.T) {
	host, ips, err := ResolveIPs("http://127.0.0.1:8500/scan")
	if err != nil || host != "127.0.0.1" || len(ips) != 1 || ips[0] != "127.0.0.1" {
		t.Fatalf("IP literal resolve = %q %v %v", host, ips, err)
	}
}
