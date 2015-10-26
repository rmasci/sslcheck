package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var runAs = filepath.Base(os.Args[0])
var verb bool

// Usage is what is run if the right parameters are not met upon startup.
func Usage() {
	// To embed the bot user and password comment the line above and uncomment the line below
	fmt.Printf("Usage: %v -i <ip address>  -p <port> -d <domain name>\n", runAs)
	flag.PrintDefaults()
}

const shortForm = "2006-01-02 15:04:05"

func main() {
	osExit := 0
	timeNowSec := time.Now().UTC().Unix()
	var ipAddress, domainName, port string
	// define flags passed at runtime, and assign them to the variables defined above
	flag.StringVar(&ipAddress, "i", "", "IP Address")
	flag.StringVar(&domainName, "d", "", "Domain Name")
	flag.BoolVar(&verb, "v",false, "Verbose")
	flag.StringVar(&port, "p", "443", "Port Number")
	flag.Parse()
	if domainName == "" {
		Usage()
		os.Exit(1)
	}
	if ipAddress == "" {
		if verb {
			fmt.Println("looking up: ", domainName)
		}
		ip, err := net.LookupHost(domainName)
		if err != nil {
			fmt.Printf("Could not resolve domain name, %v.\n\n", domainName)
			fmt.Printf("Either supply a valid domain name or use the -i switch to supply the ip address.\n")
			fmt.Printf("Domain name lookups are not performed when the user provides the ip address.\n")
			Usage()
			os.Exit(1)
		}
		if verb {
			fmt.Printf("IP Addresses: %v\n", ip)
			fmt.Println("IP Address:",ip[0])
		}
		ipAddress = ip[0] // + ":" + port
	}
	//
	//Connect network
	if len(ipAddress) > 16 {
		ipAddress = "[" + ipAddress + "]:" + port
	} else {
		ipAddress = ipAddress + ":" + port
	}
	if verb {fmt.Println("IPAddress Length",len(ipAddress))}
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
		fmt.Printf("Cert Failed for %v - %v\n", ipAddress, domainName)
		os.Exit(1)
	}
	fmt.Printf("Client connected to: %v\n", conn.RemoteAddr())
	state := conn.ConnectionState()
	fmt.Printf("Cert Checks OK\n")
	i := 0
	for _, v := range state.PeerCertificates {
		if i == 0 {
			sslFrom := fmt.Sprint(v.NotBefore)
			sslTo := fmt.Sprint(v.NotAfter)
			fmt.Println("Server key information:")
			fmt.Printf("\tCN:\t %v\n\tOU:\t %v\n\tOrg:\t %v\n", v.Subject.CommonName, strings.Trim(fmt.Sprintf("%v",v.Subject.OrganizationalUnit),"[]"), strings.Trim(fmt.Sprintf("%v",v.Subject.Organization),"[]"))
			fmt.Printf("\tCity:\t %v\n\tState:\t %v\n\tCountry: %v\n", v.Subject.Locality[0], v.Subject.Province[0], v.Subject.Country[0])
			fmt.Printf("SSL Certificate Valid:\n\tFrom:\t %s\n\tTo:\t %s\n", sslFrom, sslTo)

			//check date...
			ts := strings.Split(sslTo, " ")
			timeExp, err := time.Parse(shortForm, ts[0]+" "+ts[1])
			if err != nil {
				fmt.Printf("Could not parse time: %v %v", ts[0], err)
				osExit = 1
			}
			timeExpSec := timeExp.Unix()
			if timeExpSec < timeNowSec {
				fmt.Printf("\tWARNING:\t Cert has expired.\n")
				osExit = 1
			}
			time2Exp := (((timeExpSec - timeNowSec) / 60) / 60) / 24
			if time2Exp < 31 {
				fmt.Printf("\tWARNING:\t Cert Expires in %v days\n", time2Exp)
				osExit = 1

			} else {
				fmt.Printf("\tOK: \tCert Expires in %v days\n", time2Exp)
				//fmt.Printf("TimeNow: %v, TimeCert: %v\n", timeNowSec, timeExpSec)
			}
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
			fmt.Printf("Issued by:\n\t%v\n\t%v\n\t%v\n", v.Subject.CommonName, strings.Trim(fmt.Sprintf("%v",v.Subject.OrganizationalUnit),"[]"), strings.Trim(fmt.Sprintf("%v",v.Subject.Organization),"[]"))
			i++
		} else {
			break
		}
	}
	os.Exit(osExit)

}
