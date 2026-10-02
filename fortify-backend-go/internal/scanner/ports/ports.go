// Package ports implements host resolution + TCP port scanning.
//
// Profiles: "top100" (100 curated commonly-open ports), "top1000" (exactly
// 1000: top100 + the 1–1024 well-known range + curated high ports),
// "full" (all 65535). Probes run over a bounded worker pool with a shared
// dialer, and open ports get a light banner grab. This is the concurrent
// counterpart to the Python backend, which had no port scanning at all.
package ports

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ProfileNone    = "none"
	ProfileTop100  = "top100"
	ProfileTop1000 = "top1000"
	ProfileFull    = "full"
)

const dialTimeout = 1200 * time.Millisecond
const bannerTimeout = 1200 * time.Millisecond
const maxWorkers = 512
const maxIPs = 8

// Top100Ports: curated most-commonly-open ports (nmap top-ports inspired).
var top100Ports = []int{
	80, 443, 22, 21, 23, 25, 53, 110, 143, 445,
	139, 135, 3389, 3306, 1433, 5432, 6379, 27017, 8080, 8443,
	8000, 8888, 5900, 111, 993, 995, 587, 465, 548, 113,
	81, 591, 199, 1720, 1723, 514, 5060, 179, 1026, 1025,
	2000, 32768, 554, 26, 49152, 1027, 2049, 210, 389, 636,
	1521, 1494, 162, 161, 69, 79, 194, 6667, 119, 512,
	513, 515, 540, 1524, 1080, 3128, 5800, 4444, 5555, 9999,
	10000, 6000, 6001, 49153, 49154, 49155, 49156, 49157, 6200, 9100,
	9200, 9300, 11211, 27018, 5000, 7001, 1812, 1813, 9535, 9595,
	3000, 4000, 5001, 8001, 8008, 8010, 8081, 8090, 8123, 8181,
}

// extraHighPorts: curated commonly-open registered/dynamic ports used to pad
// top100 up to exactly 1000 after the 1–1024 well-known range.
var extraHighPorts = []int{
	3000, 3001, 4000, 4040, 4567, 5001, 5020, 5050, 5051, 5061,
	5101, 5102, 5222, 5269, 5357, 5433, 5672, 5985, 5986, 6378,
	6380, 6481, 6514, 6556, 6666, 7000, 7002, 7070, 7080, 7214,
	7474, 8009, 8011, 8042, 8069, 8070, 8082, 8085, 8088, 8089,
	8091, 8118, 8125, 8140, 8153, 8161, 8172, 8180, 8182, 8228,
	8243, 8280, 8333, 8403, 8414, 8444, 8500, 8530, 8531, 8585,
	8649, 8686, 8767, 8834, 8880, 8883, 8889, 8983, 8999, 9000,
	9001, 9007, 9009, 9010, 9020, 9030, 9042, 9050, 9080, 9090,
	9091, 9101, 9102, 9111, 9160, 9207, 9211, 9292, 9312, 9418,
	9419, 9421, 9443, 9444, 9594, 9596, 9600, 9631, 9696, 9700,
	9800, 9876, 9877, 9888, 9929, 9939, 9943, 9944, 9966, 9998,
	10001, 10080, 10081, 10082, 10083, 10101, 10102, 10202, 10203, 10243,
	11214, 11300, 12000, 12203, 12345, 12401, 12975, 13456, 13722, 13724,
	13782, 13783, 14000, 14238, 14414, 15000, 15002, 15003, 15004, 15555,
	15672, 16000, 16010, 16666, 17185, 17500, 18080, 18081, 18082, 18083,
	18084, 18092, 18104, 18245, 18264, 18301, 18463, 18769, 18888, 19000,
	19150, 19283, 19315, 19350, 19780, 19812, 19888, 20000, 20005, 20031,
	20222, 20808, 21025, 21221, 21452, 21571, 21833, 22222, 22351, 22623,
	22763, 23053, 23546, 23750, 23820, 24284, 25000, 25003, 25565, 26000,
	26001, 26002, 26003, 26004, 26005, 26122, 27000, 27015, 27374, 27665,
	28017, 28275, 29418, 30303, 30564, 31111, 31221, 32000, 32123, 32249,
	32683, 32684, 32896, 32976, 33333, 33434, 33690, 34371, 34571, 34572,
	34573, 34787, 35357, 36330, 36567, 36963, 37651, 37777, 37810, 38102,
	38292, 40193, 40911, 41511, 42510, 43439, 44134, 44442, 44443, 44501,
	44818, 46823, 46824, 47624, 47808, 49160, 49161, 49163, 49165, 49167,
	49175, 49176, 49400, 49999, 50000, 50030, 50060, 50070, 50075, 50090,
	50100, 50111, 50234, 50236, 50636, 50800, 51102, 51103, 51111, 51234,
	51393, 51493, 51500, 51515, 52311, 52413, 52673, 52848, 52869, 54045,
	54328, 55055, 55056, 55555, 55600, 56738, 57294, 57772, 58080, 59001,
	59002, 59501, 60008, 60012, 60020, 60443, 60855, 60945, 61001, 61002,
}

// serviceNames guesses the service behind well-known ports.
var serviceNames = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
	69: "tftp", 79: "finger", 80: "http", 81: "http-alt", 110: "pop3",
	111: "rpcbind", 113: "ident", 119: "nntp", 135: "msrpc", 139: "netbios",
	143: "imap", 161: "snmp", 162: "snmp-trap", 179: "bgp", 194: "irc",
	199: "smux", 210: "z39.50", 389: "ldap", 445: "smb", 465: "smtps",
	512: "exec", 513: "login", 514: "shell", 515: "printer", 540: "uucp",
	548: "afp", 554: "rtsp", 587: "smtp-submission", 591: "http-alt", 636: "ldaps",
	443: "https",
	993: "imaps", 995: "pop3s", 1080: "socks", 1433: "mssql", 1494: "citrix-ica",
	1521: "oracle", 1524: "ingreslock", 1720: "h323", 1723: "pptp", 1812: "radius",
	1813: "radius-acct", 2000: "cisco-sccp", 2049: "nfs", 3000: "http-alt", 3128: "squid",
	3306: "mysql", 3389: "rdp", 5060: "sip", 5432: "postgres", 5800: "vnc-http",
	5900: "vnc", 6000: "x11", 6001: "x11:1", 6379: "redis", 6667: "irc",
	8000: "http-alt", 8001: "http-alt", 8008: "http-alt", 8009: "ajp", 8010: "http-alt",
	8020: "http-alt", 8080: "http-alt", 8081: "http-alt", 8088: "http-alt", 8090: "http-alt",
	8118: "privoxy", 8123: "http-alt", 8140: "http-alt", 8181: "http-alt", 8228: "http-alt",
	8243: "https-alt", 8280: "http-alt", 8333: "http-alt", 8403: "http-alt", 8414: "http-alt",
	8443: "https-alt", 8444: "http-alt", 8500: "http-alt", 8530: "http-alt", 8531: "http-alt",
	8585: "http-alt", 8649: "http-alt", 8686: "http-alt", 8767: "http-alt", 8834: "http-alt",
	8880: "http-alt", 8883: "mqtt-tls", 8888: "http-alt", 8889: "http-alt", 8983: "solr",
	8999: "http-alt", 9000: "http-alt", 9001: "http-alt", 9007: "http-alt", 9009: "http-alt",
	9010: "http-alt", 9020: "http-alt", 9030: "http-alt", 9042: "cassandra", 9080: "http-alt",
	9090: "http-alt", 9091: "transmission", 9100: "jetdirect", 9101: "jetdirect", 9102: "jetdirect",
	9111: "http-alt", 9160: "cassandra", 9200: "elasticsearch", 9207: "http-alt", 9211: "http-alt",
	9292: "http-alt", 9300: "elasticsearch", 9312: "docker", 9418: "git", 9419: "http-alt",
	9443: "https-alt", 9444: "http-alt", 9594: "http-alt", 9595: "http-alt", 9596: "http-alt",
	9600: "http-alt", 9631: "http-alt", 9696: "http-alt", 9700: "http-alt", 9800: "http-alt",
	9876: "http-alt", 9877: "http-alt", 9888: "http-alt", 9929: "nping-echo", 9939: "http-alt",
	9943: "http-alt", 9944: "http-alt", 9966: "http-alt", 9998: "http-alt", 9999: "http-alt",
	10000: "webmin", 10001: "http-alt", 10080: "http-alt", 10081: "http-alt", 10082: "http-alt",
	10083: "http-alt", 10101: "http-alt", 10102: "http-alt", 10202: "http-alt", 10203: "http-alt",
	10243: "http-alt", 11211: "memcached", 11214: "memcached", 11300: "http-alt", 12000: "http-alt",
	12203: "http-alt", 12345: "netbus", 12401: "http-alt", 12975: "http-alt", 13456: "http-alt",
	13722: "http-alt", 13724: "http-alt", 13782: "http-alt", 13783: "http-alt", 14000: "http-alt",
	14238: "http-alt", 14414: "http-alt", 15000: "http-alt", 15002: "http-alt", 15003: "http-alt",
	15004: "http-alt", 15555: "http-alt", 15672: "rabbitmq", 16000: "http-alt", 16010: "http-alt",
	16666: "http-alt", 17185: "http-alt", 17500: "dropbox", 18080: "http-alt", 18081: "http-alt",
	18082: "http-alt", 18083: "http-alt", 18084: "http-alt", 18092: "http-alt", 18104: "http-alt",
	18245: "http-alt", 18264: "http-alt", 18301: "http-alt", 18463: "http-alt", 18769: "http-alt",
	18888: "http-alt", 19000: "http-alt", 19150: "http-alt", 19283: "http-alt", 19315: "http-alt",
	19350: "http-alt", 19780: "http-alt", 19812: "http-alt", 19888: "http-alt", 20000: "http-alt",
	20005: "http-alt", 20031: "http-alt", 20222: "http-alt", 20808: "http-alt", 21025: "http-alt",
	21221: "http-alt", 21452: "http-alt", 21571: "http-alt", 21833: "http-alt", 22222: "http-alt",
	22351: "http-alt", 22623: "http-alt", 22763: "http-alt", 23053: "http-alt", 23546: "http-alt",
	23750: "http-alt", 23820: "http-alt", 24284: "http-alt", 25000: "http-alt", 25003: "http-alt",
	25565: "minecraft", 26000: "quake", 26001: "quake", 26002: "quake", 26003: "quake",
	26004: "quake", 26005: "quake", 26122: "http-alt", 27000: "http-alt", 27015: "srcds",
	27017: "mongodb", 27018: "mongodb", 27374: "sub7", 27665: "http-alt", 28017: "mongodb",
	28275: "http-alt", 29418: "git", 30303: "http-alt", 30564: "http-alt", 31111: "http-alt",
	31221: "http-alt", 32000: "http-alt", 32123: "http-alt", 32249: "http-alt", 32683: "http-alt",
	32684: "http-alt", 32896: "http-alt", 32976: "http-alt", 33333: "http-alt", 33434: "traceroute",
	33690: "http-alt", 34371: "http-alt", 34571: "http-alt", 34572: "http-alt", 34573: "http-alt",
	34787: "http-alt", 35357: "http-alt", 36330: "http-alt", 36567: "http-alt", 36963: "http-alt",
	37651: "http-alt", 37777: "dahua", 37810: "http-alt", 38102: "http-alt", 38292: "http-alt",
	40193: "http-alt", 40911: "http-alt", 41511: "http-alt", 42510: "http-alt", 43439: "http-alt",
	44134: "http-alt", 44442: "http-alt", 44443: "http-alt", 44501: "http-alt", 44818: "ethernet-ip",
	46823: "http-alt", 46824: "http-alt", 47624: "http-alt", 47808: "bacnet", 49152: "unknown",
	49153: "unknown", 49154: "unknown", 49155: "unknown", 49156: "unknown", 49157: "unknown",
	49160: "unknown", 49161: "unknown", 49163: "unknown", 49165: "unknown", 49167: "unknown",
	49175: "unknown", 49176: "unknown", 49400: "http-alt", 49999: "http-alt", 50000: "http-alt",
	50030: "http-alt", 50060: "http-alt", 50070: "hadoop", 50075: "hadoop", 50090: "http-alt",
	50100: "http-alt", 50111: "http-alt", 50234: "http-alt", 50236: "http-alt", 50636: "http-alt",
	50800: "http-alt", 51102: "http-alt", 51103: "http-alt", 51111: "http-alt", 51234: "http-alt",
	51393: "http-alt", 51493: "http-alt", 51500: "http-alt", 51515: "http-alt", 52311: "http-alt",
	52413: "http-alt", 52673: "http-alt", 52848: "http-alt", 52869: "http-alt", 54045: "http-alt",
	54328: "http-alt", 55055: "http-alt", 55056: "http-alt", 55555: "http-alt", 55600: "http-alt",
	56738: "http-alt", 57294: "http-alt", 57772: "http-alt", 58080: "http-alt", 59001: "vnc",
	59002: "vnc", 59501: "http-alt", 60008: "http-alt", 60012: "http-alt", 60020: "http-alt",
	60443: "https-alt", 60855: "http-alt", 60945: "http-alt", 61001: "http-alt", 61002: "http-alt",
}

// RiskyPorts: remotely-administered or data-store services whose exposure is
// a high-severity finding on its own (scored higher in the analyzer).
var RiskyPorts = map[int]bool{
	21: true, 23: true, 69: true, 79: true, 111: true,
	512: true, 513: true, 514: true, 139: true, 445: true,
	1433: true, 1434: true, 1521: true, 2049: true, 2375: true,
	2376: true, 2379: true, 3306: true, 3389: true, 5432: true,
	5900: true, 6000: true, 6001: true, 6379: true, 9042: true,
	9200: true, 9300: true, 10250: true, 11211: true, 27017: true,
	27018: true,
}

// IsRisky reports whether an open port is sensitive enough for a high score.
func IsRisky(port int) bool { return RiskyPorts[port] }

// ServiceName guesses the service behind a port ("unknown" fallback).
func ServiceName(port int) string {
	if s, ok := serviceNames[port]; ok {
		return s
	}
	return "unknown"
}

// ValidProfile reports whether p is a known port-scan profile.
func ValidProfile(p string) bool {
	switch p {
	case ProfileNone, ProfileTop100, ProfileTop1000, ProfileFull:
		return true
	}
	return false
}

// PortsForProfile returns the probe list: exactly 100 / exactly 1000 / all.
func PortsForProfile(profile string) []int {
	switch profile {
	case ProfileTop100:
		return append([]int(nil), top100Ports...)
	case ProfileTop1000:
		seen := make(map[int]bool, 1000)
		out := make([]int, 0, 1000)
		for _, p := range top100Ports {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
		for p := 1; p <= 1024 && len(out) < 1000; p++ {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
		for _, p := range extraHighPorts {
			if len(out) >= 1000 {
				break
			}
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
		return out
	case ProfileFull:
		all := make([]int, 0, 65535)
		for p := 1; p <= 65535; p++ {
			all = append(all, p)
		}
		return all
	default:
		return nil
	}
}

// OpenPort is one reachable TCP port with an optional service banner.
type OpenPort struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
	Banner  string `json:"banner,omitempty"`
}

// HostResult is the per-IP outcome of a port scan.
type HostResult struct {
	IP           string     `json:"ip"`
	OpenPorts    []OpenPort `json:"open_ports"`
	PortsScanned int        `json:"ports_scanned"`
}

// Result mirrors the JSON the dashboard renders under results.ports.
type Result struct {
	Host         string       `json:"host"`
	IPs          []string     `json:"ips"`
	Profile      string       `json:"profile"`
	PortsScanned int          `json:"ports_scanned"`
	Hosts        []HostResult `json:"hosts"`
	DurationMs   int64        `json:"duration_ms"`
	Error        string       `json:"error,omitempty"`
}

// ResolveIPs maps a URL hostname to scan targets: IP literals pass through,
// names go through DNS (capped at maxIPs).
func ResolveIPs(rawURL string) (host string, ips []string, err error) {
	u, perr := url.Parse(rawURL)
	if perr != nil || u.Hostname() == "" {
		return "", nil, fmt.Errorf("invalid URL")
	}
	host = u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		return host, []string{host}, nil
	}
	addrs, lerr := net.LookupIP(host)
	if lerr != nil || len(addrs) == 0 {
		return host, nil, fmt.Errorf("DNS resolution failed for %s", host)
	}
	seen := map[string]bool{}
	// IPv4 first — the usual scan target; then IPv6.
	sort.Slice(addrs, func(i, j int) bool {
		i4, j4 := addrs[i].To4() != nil, addrs[j].To4() != nil
		if i4 != j4 {
			return i4
		}
		return addrs[i].String() < addrs[j].String()
	})
	for _, a := range addrs {
		s := a.String()
		if !seen[s] {
			seen[s] = true
			ips = append(ips, s)
		}
		if len(ips) >= maxIPs {
			break
		}
	}
	return host, ips, nil
}

func probe(ip string, port int) (OpenPort, bool) {
	target := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		return OpenPort{}, false // closed or filtered — not a finding
	}
	defer conn.Close()
	// Light banner grab: many services (SSH, FTP, SMTP, …) speak first.
	// HTTP-style services stay silent — that just yields an empty banner.
	_ = conn.SetReadDeadline(time.Now().Add(bannerTimeout))
	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	banner := strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, string(buf[:n])))
	return OpenPort{Port: port, Service: ServiceName(port), Banner: banner}, true
}

// ScanPorts probes an explicit port list on one IP over a bounded pool.
func ScanPorts(ip string, ports []int) []OpenPort {
	if len(ports) == 0 {
		return []OpenPort{}
	}
	workers := maxWorkers
	if len(ports) < workers {
		workers = len(ports)
	}
	jobs := make(chan int, len(ports))
	for _, p := range ports {
		jobs <- p
	}
	close(jobs)

	var mu sync.Mutex
	open := []OpenPort{}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				if found, ok := probe(ip, p); ok {
					mu.Lock()
					open = append(open, found)
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	sort.Slice(open, func(i, j int) bool { return open[i].Port < open[j].Port })
	return open
}

// ScanHost resolves the URL to IPs and scans the profile's ports on each.
// Independent IPs are scanned concurrently.
func ScanHost(rawURL, profile string) Result {
	start := time.Now()
	host, ips, err := ResolveIPs(rawURL)
	if err != nil {
		return Result{Host: host, Profile: profile, Hosts: []HostResult{}, IPs: []string{}, Error: err.Error()}
	}
	ports := PortsForProfile(profile)

	var mu sync.Mutex
	hosts := make([]HostResult, 0, len(ips))
	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			open := ScanPorts(target, ports)
			if open == nil {
				open = []OpenPort{}
			}
			mu.Lock()
			hosts = append(hosts, HostResult{IP: target, OpenPorts: open, PortsScanned: len(ports)})
			mu.Unlock()
		}(ip)
	}
	wg.Wait()
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].IP < hosts[j].IP })
	return Result{
		Host: host, IPs: ips, Profile: profile,
		PortsScanned: len(ports), Hosts: hosts,
		DurationMs: time.Since(start).Milliseconds(),
	}
}
