// Go equivalent of the "DNS & BIND" book check-soa program.
// Created by Stephane Bortzmeyer.
package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	ioview "github.com/eogns47/NameServer_Finder/IOView"
	"github.com/miekg/dns"
	"github.com/oschwald/geoip2-golang"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	if url[ul-2] == '/' {
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
			UDPSize:     4096,
		},
	}
}

// packet limit = 512 bytes
func (zr *ZoneNsResolver) localQuery(qname string, qtype uint16, server string) (*dns.Msg, error) {

	zr.localm.SetQuestion(qname, qtype)

	r, _, err := zr.localc.Exchange(zr.localm, server+":53")
	if err != nil {
		return nil, errors.Wrap(err, "localc Exchange failed")
	}
	if r == nil || r.Rcode == dns.RcodeNameError || r.Rcode == dns.RcodeSuccess {
		return r, nil
	}

	return nil, errors.Wrap(err, "No name server to answer the question")
}

func (zr *ZoneNsResolver) Resolve(zone string, server string) ([]string, error) {
	zone = dns.Fqdn(zone)

	r, err := zr.localQuery(zone, dns.TypeNS, server)
	if err != nil || r == nil {
		return nil, errors.Wrap(err, "localQuery failed")
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
		return nil, errors.Wrap(err, "No nameservers found for "+zone)
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
		return nil, errors.Wrap(err, "LookupHost failed")
	}

	return ips, nil
}

func getCountryCode(ip string) (string, error) {
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		return "", errors.Wrap(err, "Open Geoip failed")
	}
	defer db.Close()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", errors.Wrap(err, "Invalid IP address: "+ip)
	}

	record, err := db.Country(parsedIP)
	if err != nil {
		return "", errors.Wrap(err, "db.Country failed")
	}

	return strings.ToLower(record.Country.IsoCode), nil
}

func isIPv4orIPv6(ipStr string) int {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	} else if ip.To16() != nil && ip.To4() == nil {
		return 6
	} else if ip.To4() != nil {
		return 4
	} else {
		return 0
	}
}

func fileLogger(logFolder string) *zap.Logger {
	// Check if the log folder exists, create it if not
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		err := os.Mkdir(logFolder, 0755)
		if err != nil {
			panic(err)
		}
	}
	filename := filepath.Join(logFolder, "log.log")

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

func main() {
	// initialize the rotator
	currentDir, err := os.Getwd()
	logFile := filepath.Join(currentDir, "/log")

	logger := fileLogger(logFile)

	if len(os.Args) != 2 {
		logger.Warn(os.Args[0] + "ZONE")
		return
	}

	target := os.Args[1]
	records, err := ioview.ReadCsv(target)

	if err != nil {
		logger.Warn("🚨Error with Input csv:" + err.Error())
	}

	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil || conf == nil {
		logger.Warn("Cannot initialize the local resolver: %s\n" + err.Error())
	}

	resolver := NewZoneNsResolver()
	db, err := ioview.GetDBConnect()
	if err != nil {
		logger.Warn("🚨Error with DB Connect:" + err.Error())
	}

	startTime := time.Now()
	logger.Info("Start find NS of " + target)
	for _, record := range records {
		urlcrc, err := strconv.Atoi(record[1])
		if err != nil {
			logger.Info("URL " + record[0] + " Dont have CRC")
			continue
		}
		searchId, err := ioview.InsertURLSearchDataIntoTable(db, ioview.URLSearchData{URL: record[0], URLCRC: int64(urlcrc)})
		if err != nil {
			logger.Warn("Error:" + err.Error())
			return
		}

		fmt.Println("-----------------------------------------------------------------------------------------------------------------")

		domain := dns.Fqdn(record[0])
		domain = removeHTTPPrefix(domain)

		var ns []string
		nextNs := conf.Servers[0]

		// split domain, query each part for NS records
		for i, zone := range domainToZones(domain) {
			if i > 2 {
				break
			}

			if zone != "." {
				fmt.Println("🔎Finding nameservers for zone '" + zone + "' using parent nameserver '" + nextNs + "'\n")
			}

			ns, err = resolver.Resolve(zone, nextNs)
			if err != nil {
				logger.Warn("🚨Query failed: " + err.Error())
				break
			}

			// Pick a random NS record for the next queries
			nextNs = ns[rand.Intn(len(ns))]
		}

		fmt.Println("📜nameserver List:")
		for _, nameserver := range ns {
			fmt.Println(nameserver)
		}

		var nameserverIPs []string

		for _, nameserver := range ns {
			IPs, err := getIPAddresses(nameserver)
			if err != nil {
				// 오류 처리
				logger.Warn("Error for Nameservers : " + nameserver + err.Error())
				continue
			}
			nameserverIPs = append(nameserverIPs, IPs...)

			for _, ip := range IPs {
				countryCode, err := getCountryCode(ip)
				if err != nil {
					logger.Warn("Error for Nameserver ip: " + ip + err.Error())
					continue
				}
				ipType := isIPv4orIPv6(ip)
				ioview.InsertNameServerDataIntoTable(db, ioview.NameServerData{SearchID: searchId, NameServer: nameserver, IP: ip, CountryCode: countryCode, IPType: ipType})
			}

		}

		// getIPAddresses 함수를 사용하여 URL에 대한 IP 주소 조회
		ipAddresses, err := getIPAddresses(domain)
		if err != nil {
			logger.Warn("Error for URL's Ip: " + domain + err.Error())
			return
		}

		fmt.Println("\n📜IP Addresses for nameservers:")
		for _, ip := range nameserverIPs {
			countryCode, err := getCountryCode(ip)
			if err != nil {
				logger.Warn("Error for Nameserver's countrycode:" + ip + err.Error())
				return
			}
			fmt.Println(ip, countryCode)
		}

		// 조회된 IP 주소 출력
		fmt.Println("\n📜IP Addresses for", domain, ":")
		for _, ip := range ipAddresses {
			countryCode, err := getCountryCode(ip)
			if err != nil {
				logger.Warn("Error's for URL's ip:" + ip + err.Error())
				return
			}
			fmt.Println(ip, countryCode)
			ioview.InsertWebIPDataIntoTable(db, ioview.WebIpData{SearchID: searchId, IP: ip, CountryCode: countryCode})
		}
	}
	elapsedTime := time.Since(startTime).Seconds()
	elapsedTimeStr := fmt.Sprintf("%.2f sec", elapsedTime)
	logger.Info("elapsed time for " + strconv.Itoa(len(records)) + " URLs :" + elapsedTimeStr)

}
