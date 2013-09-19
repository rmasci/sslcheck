package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"net"
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
	flag.StringVar(&ipAddress, "i", "", "IP Address")
	flag.StringVar(&domainName, "d", "", "Domain Name")
	flag.StringVar(&port, "p", "443", "Port Number")
	flag.Parse()
	if domainName == "" {
		Usage()
		os.Exit(1)
	}
	if ipAddress == "" {
		ip, err := net.LookupHost(domainName)
		if err != nil {
			fmt.Printf("Could not resolve domain name, %v.\n\n",domainName)
			fmt.Printf("Either supply a valid domain name or use the -i switch to supply the ip address.\n")
			fmt.Printf("Domain name lookups are not performed when the user provides the ip address.\n")
			os.Exit(1)
		}
		ipAddress = ip[0] + ":" + port
	} else {
		ipAddress = ipAddress + ":" + port
	}
	//Connect network
	ipConn,err:=net.DialTimeout("tcp",ipAddress,60000*time.Millisecond)
	if err != nil {
		fmt.Printf("Could not connect to %v - %v\n",ipAddress,domainName)
		os.Exit(1)
	} else {
		defer ipConn.Close()
	}
	// Configure tls to look at domainName
	config := tls.Config{ServerName: domainName}
	// Connect to tls
	conn:= tls.Client(ipConn, &config)
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
	i := 0
	for _, v := range state.PeerCertificates {
		if i == 0 {
			fmt.Println("Server key information:")
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
			i++
		} else if i == 1 {
			fmt.Printf("Issued by:\n\t%v\n\t%v\n\t%v\n", v.Subject.CommonName, v.Subject.OrganizationalUnit, v.Subject.Organization)
			i++
		} else {
			break
		}
	}

}
