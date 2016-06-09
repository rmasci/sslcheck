package main

import (
	"crypto/tls"
	//"flag"
	flag "github.com/spf13/pflag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

var runAs = filepath.Base(os.Args[0])

// Usage is what is run if the right parameters are not met upon startup.
func Usage() {
	// To embed the bot user and password comment the line above and uncomment the line below
	fmt.Printf("Usage: %v -i <ip address>  -p <port> -d <domain name>\n", runAs)
	flag.PrintDefaults()
}

func main() {
	var ipAddress, domainName, port string
	// define flags passed at runtime, and assign them to the variables defined above
	flag.StringVarP(&ipAddress,"ip", "i", "", "IP Address")
	flag.StringVarP(&domainName, "domain","d", "", "Domain Name")
	flag.StringVarP(&port, "port","p", "443", "Port Number")

	flag.Parse()
	if domainName == "" {
		Usage()
		os.Exit(1)
	}
	if ipAddress == "" {
		ip, err := net.LookupHost(domainName)
		if err != nil {
			fmt.Printf("Could not resolve domain name, %v.\n\n", domainName)
			fmt.Printf("Either supply a valid domain name or use the -i switch to supply the ip address.\n")
			fmt.Printf("Domain name lookups are not performed when the user provides the ip address.\n")
			os.Exit(1)
		}
		ipAddress = ip[0] + ":" + port
	} else {
		ipAddress = ipAddress + ":" + port
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
