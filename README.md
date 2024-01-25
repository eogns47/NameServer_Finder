# [![baby-gopher](https://raw.githubusercontent.com/drnic/babygopher-site/gh-pages/images/babygopher-logo-small.png)](http://www.babygopher.org) NameServer_Finder

Walk the DNS tree to find which name servers a particular zone uses. Mimics "dig +trace", reads the URL from a CSV file and stores the IPs of the entered URL, the IP of the URL's nameserver, and the country code using GEOIP.<br><br>
This repository is based on the following repository.<br>
➡️https://github.com/DNSSpy/zone-nameservers

# Building

After a git clone;

```
$ go build
$ ./NameServer_Finder {csvfile name}
```

# Examples

Here's what it looks like for test.csv

```
[test.csv]
url,url_crc,ip
https://intercrypto.com,143273094,192.99.100.45
https://download.mywebface.com,855349189,35.201.91.40
https://m14.xozejjt.com,2315816021,157.56.160.177
```

output

```
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'com.' using parent nameserver 'c.root-servers.net.'

🔎Finding nameservers for zone 'intercrypto.com.' using parent nameserver 'm.gtld-servers.net.'

📜nameserver List:
[ns3.intercrypto.com. ns4.intercrypto.com.]

📜IP Addresses for nameservers:
192.99.100.45 ca
192.99.100.45 ca

📜IP Addresses for intercrypto.com. :
192.99.100.45 ca
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'com.' using parent nameserver 'f.root-servers.net.'

🔎Finding nameservers for zone 'mywebface.com.' using parent nameserver 'm.gtld-servers.net.'

📜nameserver List:
[dns1.p03.nsone.net. dns2.p03.nsone.net. dns3.p03.nsone.net. dns4.p03.nsone.net.]

📜IP Addresses for nameservers:
198.51.44.3 us
2620:4d:4000:6259:7:3:0:1 us
198.51.45.3 us
2a00:edc0:6259:7:3::2 us
198.51.44.67 us
2620:4d:4000:6259:7:3:0:3 us
198.51.45.67 us
2a00:edc0:6259:7:3::4 us

📜IP Addresses for download.mywebface.com. :
35.201.91.40 us
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'com.' using parent nameserver 'c.root-servers.net.'

🔎Finding nameservers for zone 'xozejjt.com.' using parent nameserver 'e.gtld-servers.net.'

📜nameserver List:
[ns102a.microsoftinternetsafety.net. ns102b.microsoftinternetsafety.net.]

📜IP Addresses for nameservers:
13.107.222.41 us
13.107.206.41 us

📜IP Addresses for m14.xozejjt.com. :
157.56.160.177 us
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'online.' using parent nameserver 'l.root-servers.net.'

🔎Finding nameservers for zone 'gnldr.online.' using parent nameserver 'f.nic.online.'

📜nameserver List:
[ns3.abovedomains.com. ns4.abovedomains.com.]

📜IP Addresses for nameservers:
103.224.212.45 au
103.224.182.45 au
103.224.182.46 au
103.224.212.46 au

📜IP Addresses for gnldr.online. :
103.224.212.238 au
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'me.' using parent nameserver 'i.root-servers.net.'

🔎Finding nameservers for zone '6porn.me.' using parent nameserver 'b2.nic.me.'

📜nameserver List:
[hans.ns.cloudflare.com. tricia.ns.cloudflare.com.]

📜IP Addresses for nameservers:
173.245.59.175 us
108.162.193.175 us
172.64.33.175 us
2803:f800:50::6ca2:c1af cr
2606:4700:58::adf5:3baf us
2a06:98c1:50::ac40:21af us
108.162.192.232 us
173.245.58.232 us
172.64.32.232 us
2803:f800:50::6ca2:c0e8 cr
2606:4700:50::adf5:3ae8 us
2a06:98c1:50::ac40:20e8 us

📜IP Addresses for www.6porn.me. :
172.67.214.223 us
104.21.16.169
2606:4700:3032::ac43:d6df us
2606:4700:3036::6815:10a9 us
-----------------------------------------------------------------------------------------------------------------
🔎Finding nameservers for zone 'cc.' using parent nameserver 'i.root-servers.net.'

🔎Finding nameservers for zone 'yzjkzii.cc.' using parent nameserver 'ac1.nstld.com.'

📜nameserver List:
[ns102a.microsoftinternetsafety.net. ns102b.microsoftinternetsafety.net.]

📜IP Addresses for nameservers:
13.107.222.41 us
13.107.206.41 us

📜IP Addresses for m38.yzjkzii.cc. :
157.56.160.177 us
```

The arrow represents which nameserver from the parent was used to query for details of the child zone.

# Why?

Why not just query directly for `NS` records, you ask? Not everyone keeps those up-to-date and they often return outdated or wrong information, as nameservers change without modifying the `NS` records to reflect that.

In other words: the only _absolutely_ way to find our which nameservers a particular zone uses, you have to walk the DNS tree.

# Credits

This code is initially based on the [check-soa](https://github.com/miekg/exdns/tree/master/check-soa) script by [miekg](https://github.com/miekg)
