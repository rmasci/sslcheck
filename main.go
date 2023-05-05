package main

import (
	"crypto/tls"
	"strings"

	//"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/rmasci/verbose"
	flag "github.com/spf13/pflag"
)

var runAs = filepath.Base(os.Args[0])
var version string
var verb verbose.Verb

// Usage is what is run if the right parameters are not met upon startup.
func Usage() {
	// To embed the bot user and password comment the line above and uncomment the line below
	fmt.Printf("Usage: %v -i <ip address>  -p <port> -d <domain name>\n", runAs)
	flag.PrintDefaults()
	fmt.Println(version)
}

func main() {
	var ipAddress, domainName, port, passedPort string
	var help bool
	// define flags passed at runtime, and assign them to the variables defined above
	flag.StringVarP(&ipAddress, "ip", "i", "", "IP Address")
	flag.StringVarP(&domainName, "domain", "d", "", "Domain Name")
	flag.StringVarP(&port, "port", "p", "443", "Port Number")
	flag.BoolVarP(&help, "help", "h", false, "Help")
	flag.BoolVarP(&verb.V, "verbose", "v", false, "Verbose")
	flag.Parse()
	if help {
		Usage()
		os.Exit(0)
	}
	// IF domain name is empty -- look for it without the -d User can run sslcheck https://google.com
	if domainName == "" {
		d := flag.Args()
		if len(d) <= 0 {
			Usage()
			os.Exit(1)
		} else {
			domainName = d[0]
		}
	}
	if port != "" {
		passedPort = port
	}
	// TODO make this work.

	// Here we strip off the https:// part - User might have cut / pasted it from browser. If port is not set, set it to 443
	if strings.HasPrefix(domainName, "https://") {
		if passedPort == "" {
			port = "443"
		}
		domainName = strings.TrimLeft(domainName, "https://")
		verb.Printf("Strip https: Domain: %s\n", domainName)
	}
	// Even if the port was set above to 443, the user specified a specific  port.
	if strings.Contains(domainName, ":") {
		tmp := strings.Split(domainName, ":")
		domainName = tmp[0]
		// passedPort is what the user passed using -p <port>  We use that here because user could specify https://google.com:8443 -p 4443  In both cases the port is not touched if the user passed -p <port>
		// If the user did not pass -p <port> but put a :<port> that overrides the https:// above. Order should be get the port from -p, then from :<port> then from https://.
		if passedPort == "" {
			port = tmp[1]
		}
	}
	// in this case the IP Address overrides the domain.
	if strings.Contains(ipAddress, ":") {
		tmp := strings.Split(ipAddress, ":")
		ipAddress = tmp[0]
		if passedPort == "" {
			port = tmp[1]
		}
	}

	if ipAddress == "" {
		ip, err := net.ResolveIPAddr("ip4", domainName)
		if err != nil {
			fmt.Printf("Could not resolve domain name, %v.\n\n", domainName)
			fmt.Printf("Either supply a valid domain name or use the -i switch to supply the ip address.\n")
			fmt.Printf("Domain name lookups are not performed when the user provides the ip address.\n")
			os.Exit(1)
		}
		ipAddress = ip.IP.String() + ":" + port
	} else {
		ipAddress = ipAddress + ":" + port
	}
	// If all else failed -- use 443. User can pass sslcheck google.com and it will assume 443
	if port == "" {
		port = "443"
	}
	//Connect network
	ipConn, err := net.DialTimeout("tcp", ipAddress, 60000*time.Millisecond)
	if err != nil {
		fmt.Printf("Could not connect to %v - %v\n", ipAddress, domainName)
		os.Exit(1)
	} else {
		defer ipConn.Close()
	}
	// Configure tls to look at domainName
	config := tls.Config{ServerName: domainName}
	// Connect to tls
	conn := tls.Client(ipConn, &config)
	defer conn.Close()
	// Handshake with TLS to get cert
	hsErr := conn.Handshake()
	if hsErr != nil {
		fmt.Printf("Client connected to: %v\n", conn.RemoteAddr())
		fmt.Printf("Cert Failed for %v - %v\n", ipAddress, domainName)
		os.Exit(1)
	} else {
		fmt.Printf("Client connected to: %v\n", conn.RemoteAddr())
		fmt.Printf("Cert Checks OK\n")
	}
	state := conn.ConnectionState()
	for i, v := range state.PeerCertificates {
		switch i {
		case 0:
			fmt.Println("Server key information:")
			switch v.Version {
			case 3:
				fmt.Printf("\tVersion: TLS v1.2\n")
			case 2:
				fmt.Printf("\tVersion: TLS v1.1\n")
			case 1:
				fmt.Printf("\tVersion: TLS v1.0\n")
			case 0:
				fmt.Printf("\tVersion: SSL v3\n")
			}
			fmt.Printf("\tCN:\t %v\n\tOU:\t %v\n\tOrg:\t %v\n", v.Subject.CommonName, v.Subject.OrganizationalUnit, v.Subject.Organization)
			fmt.Printf("\tCity:\t %v\n\tState:\t %v\n\tCountry: %v\n", v.Subject.Locality, v.Subject.Province, v.Subject.Country)
			fmt.Printf("SSL Certificate Valid:\n\tFrom:\t %v\n\tTo:\t %v\n", v.NotBefore, v.NotAfter)
			fmt.Printf("Valid Certificate DNS:\n")
			if len(v.DNSNames) >= 1 {
				for dns := range v.DNSNames {
					fmt.Printf("\t%v\n", v.DNSNames[dns])
				}
			} else {
				fmt.Printf("\t%v\n", v.Subject.CommonName)
			}
		case 1:
			fmt.Printf("Issued by:\n\t%v\n\t%v\n\t%v\n", v.Subject.CommonName, v.Subject.OrganizationalUnit, v.Subject.Organization)
		default:
			break
		}
	}

}
