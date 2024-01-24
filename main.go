// Go equivalent of the "DNS & BIND" book check-soa program.
// Created by Stephane Bortzmeyer.
package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	ioview "github.com/DNSSpy/zone-nameservers/IOView"

	"github.com/miekg/dns"
	"github.com/oschwald/geoip2-golang"
)

const (
	// DefaultTimeout is default timeout many operation in this program will
	// use.
	DefaultTimeout time.Duration = 5 * time.Second
)

func init() {
	rand.Seed(time.Now().Unix())
	// rand.NewSource(time.Now().Unix())
}

type ZoneNsResolver struct {
	localm *dns.Msg
	localc *dns.Client
}

func removeHTTPPrefix(url string) string {
	ul := len(url)
	if(url[ul-2]=='/'){
		url = url[:ul-2]
	}

	// "http://" 또는 "https://"로 시작하는지 확인
	if strings.HasPrefix(url, "http://") {
		// "http://"이 있으면 잘라서 반환
		return strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		// "https://"이 있으면 잘라서 반환
		return strings.TrimPrefix(url, "https://")
	}

	// 위의 두 경우에 해당하지 않으면 원래의 URL 반환
	return url
}


func NewZoneNsResolver() *ZoneNsResolver {
	return &ZoneNsResolver{
		&dns.Msg{
			MsgHdr: dns.MsgHdr{
				RecursionDesired: true,
			},
			Question: make([]dns.Question, 1),
		},
		&dns.Client{
			ReadTimeout: DefaultTimeout,
		},
	}
}

//packet limit = 512 bytes
func (zr *ZoneNsResolver) localQuery(qname string, qtype uint16, server string) (*dns.Msg, error) {
	zr.localm.SetQuestion(qname, qtype)

	r, _, err := zr.localc.Exchange(zr.localm, server+":53")  
	if err != nil {
		return nil, err
	}
	if r == nil || r.Rcode == dns.RcodeNameError || r.Rcode == dns.RcodeSuccess {
		return r, nil
	}

	return nil, errors.New("No name server to answer the question")
}

func (zr *ZoneNsResolver) Resolve(zone string, server string) ([]string, error) {
	zone = dns.Fqdn(zone)

	r, err := zr.localQuery(zone, dns.TypeNS, server)
	if err != nil || r == nil {
		return nil, err
	}

	var nameservers []string

	
	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.NS); ok { //if(ans.*dns.NS is false, then ok is false)
			nameserver := t.Ns
			nameservers = append(nameservers, nameserver)
		}
	}
	
	
	
	if len(nameservers) == 0 {
		// No "Answer" given by the server, check the Authority section if
		// additional nameservers were provided.
		for _, ans := range r.Ns {
			if t, ok := ans.(*dns.NS); ok {
				nameserver := t.Ns
				nameservers = append(nameservers, nameserver)
			}
		}
	}
	
	if len(nameservers) == 0 {
		return nil, errors.New("No nameservers found for " + zone)
	}

	sort.Strings(nameservers)

	return nameservers, nil
}

func domainToZones(domain string) []string {
	zones := []string{"."}

	assembled := ""
	pieces := dns.SplitDomainName(domain)
	for i := len(pieces) - 1; i >= 0; i-- {
		assembled = pieces[i] + "." + assembled
		zones = append(zones, assembled)
	}


	return zones
}

func getIPAddresses(url string) ([]string, error) {
	// Using LookupHost to lookup IP addresses of domain
	ips, err := net.LookupHost(url)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

func getCountryCode(ip string) (string, error) {
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		return "", err
	}
	defer db.Close()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("Invalid IP address: %s", ip)
	}

	record, err := db.Country(parsedIP)
	if err != nil {
		return "", err
	}

	return strings.ToLower(record.Country.IsoCode), nil
}

func validIPIncludes(nameserverIPs []string, ipAddresses []string, targetIP string) int {
	for _, ip := range nameserverIPs {
		if ip == targetIP {
			return 1
		}
	}
	for _, ip := range ipAddresses {
		if ip == targetIP {
			return 2
		}
	}
	return 0
}


func main() {
	if len(os.Args) != 2 {
		log.Fatalf("%s ZONE\n", os.Args[0])
	}

	target :=os.Args[1]
	records,err :=ioview.ReadCsv(target)

	if err != nil {
		fmt.Println("🚨Error Input csv:", err)
		return
	}

	for _, record := range records {
	fmt.Println("-----------------------------------------------------------------------------------------------------------------")

	domain := dns.Fqdn(record[0])
	domain = removeHTTPPrefix(domain)

	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || conf == nil {
		log.Fatalf("Cannot initialize the local resolver: %s\n", err)
	}


	resolver := NewZoneNsResolver()

	var ns []string
	nextNs := conf.Servers[0]

	// split domain, query each part for NS records
	for i, zone := range domainToZones(domain) {
		if(i>2){
			break
		}

		if zone != "." {
			fmt.Println("🔎Finding nameservers for zone '" + zone + "' using parent nameserver '" + nextNs + "'\n")
		}

		ns, err = resolver.Resolve(zone, nextNs)
		if err != nil {
			log.Fatalln("🚨Query failed: ", err)
		
		}


		// Pick a random NS record for the next queries
		nextNs = ns[rand.Intn(len(ns))]

		// Print the nameservers for this zone, highlight the one we used to query
		for _, nameserver := range ns {
			if nameserver == nextNs && domain != zone {
				// We'll use this one for queries
				// fmt.Println(" ➡️ " + nameserver)
			} else {
				// fmt.Println(" - " + nameserver)
			}
		}
	}

	fmt.Println("📜nameserver List:")
	fmt.Println(ns)

	var nameserverIPs []string

	for _, nameserver := range ns {
		IPs, err := getIPAddresses(nameserver)
		if err != nil {
        // 오류 처리
			fmt.Println("Error:", err)
			continue
		}
		nameserverIPs = append(nameserverIPs, IPs...)
    }


	// getIPAddresses 함수를 사용하여 URL에 대한 IP 주소 조회
	ipAddresses, err := getIPAddresses(domain)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("\n📜IP Addresses for nameservers:")
	for _, ip := range nameserverIPs {
		countryCode, err := getCountryCode(ip)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(ip, countryCode)
	}

	// 조회된 IP 주소 출력
	fmt.Println("\n📜IP Addresses for", domain, ":")
	for _, ip := range ipAddresses {
		countryCode, err := getCountryCode(ip)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(ip, countryCode)
	}

	ipBelong :=validIPIncludes(nameserverIPs, ipAddresses, record[2])

	fmt.Println("Guess IP:" ,record[2])
	fmt.Println("IP Belong:",ipBelong)


	fmt.Println("-----------------------------------------------------------------------------------------------------------------")

	}
}
