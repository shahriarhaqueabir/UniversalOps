package netops

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSResult holds DNS lookup results.
type DNSResult struct {
	Hostname string
	A        []string
	AAAA     []string
	MX       []string
	NS       []string
	CNAME    string
	TXT      []string
	Error    string
}

// publicDNS holds public DNS servers to query.
var publicDNS = []string{
	"8.8.8.8:53",
	"1.1.1.1:53",
}

// LookupDNS performs DNS lookups for a given hostname using public DNS servers.
func LookupDNS(hostname string) (*DNSResult, error) {
	result := &DNSResult{
		Hostname: hostname,
	}

	client := new(dns.Client)
	client.Timeout = 5 * time.Second

	// Try each DNS server until one works
	var lastErr error
	for _, server := range publicDNS {
		// A records
		records, err := queryDNS(client, server, hostname, dns.TypeA)
		if err == nil {
			result.A = records
		} else {
			lastErr = err
		}

		// AAAA records
		records, err = queryDNS(client, server, hostname, dns.TypeAAAA)
		if err == nil {
			result.AAAA = records
		}

		// MX records
		mxRecords, err := queryMX(client, server, hostname)
		if err == nil {
			result.MX = mxRecords
		}

		// NS records
		records, err = queryDNS(client, server, hostname, dns.TypeNS)
		if err == nil {
			result.NS = records
		}

		// CNAME
		cname, err := queryCNAME(client, server, hostname)
		if err == nil {
			result.CNAME = cname
		}

		// TXT records
		records, err = queryDNS(client, server, hostname, dns.TypeTXT)
		if err == nil {
			result.TXT = records
		}

		// If we got results, return them
		if len(result.A) > 0 || len(result.AAAA) > 0 || len(result.MX) > 0 {
			sort.Strings(result.A)
			sort.Strings(result.AAAA)
			sort.Strings(result.NS)
			sort.Strings(result.TXT)
			return result, nil
		}
	}

	if lastErr != nil {
		result.Error = lastErr.Error()
		return result, lastErr
	}

	return result, fmt.Errorf("no DNS servers could be reached")
}

// queryDNS performs a DNS query for a specific record type.
func queryDNS(client *dns.Client, server, hostname string, qtype uint16) ([]string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), qtype)

	reply, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, err
	}

	if reply == nil || len(reply.Answer) == 0 {
		return nil, fmt.Errorf("no records found for type %d", qtype)
	}

	var results []string
	seen := make(map[string]bool)

	for _, ans := range reply.Answer {
		switch h := ans.(type) {
		case *dns.A:
			ip := h.A.String()
			if !seen[ip] {
				results = append(results, ip)
				seen[ip] = true
			}
		case *dns.AAAA:
			ip := h.AAAA.String()
			if !seen[ip] {
				results = append(results, ip)
				seen[ip] = true
			}
		case *dns.NS:
			ns := h.Ns
			if !seen[ns] {
				results = append(results, strings.TrimSuffix(ns, "."))
				seen[ns] = true
			}
		case *dns.TXT:
			txt := strings.Join(h.Txt, " ")
			if !seen[txt] {
				results = append(results, txt)
				seen[txt] = true
			}
		}
	}

	return results, nil
}

// queryMX performs an MX record lookup.
func queryMX(client *dns.Client, server, hostname string) ([]string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeMX)

	reply, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, err
	}

	if reply == nil || len(reply.Answer) == 0 {
		return nil, fmt.Errorf("no MX records found")
	}

	type mxEntry struct {
		preference uint16
		host       string
	}

	var entries []mxEntry
	for _, ans := range reply.Answer {
		if mx, ok := ans.(*dns.MX); ok {
			entries = append(entries, mxEntry{
				preference: mx.Preference,
				host:       strings.TrimSuffix(mx.Mx, "."),
			})
		}
	}

	// Sort by preference
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].preference < entries[j].preference
	})

	var results []string
	for _, e := range entries {
		results = append(results, fmt.Sprintf("%s (priority %d)", e.host, e.preference))
	}

	return results, nil
}

// queryCNAME performs a CNAME lookup.
func queryCNAME(client *dns.Client, server, hostname string) (string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeCNAME)

	reply, _, err := client.Exchange(msg, server)
	if err != nil {
		return "", err
	}

	if reply == nil || len(reply.Answer) == 0 {
		return "", fmt.Errorf("no CNAME records found")
	}

	for _, ans := range reply.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			return strings.TrimSuffix(cname.Target, "."), nil
		}
	}

	return "", fmt.Errorf("no CNAME records found")
}
