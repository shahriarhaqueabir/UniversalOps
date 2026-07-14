package netops

import "strings"

var ouiDB = map[string]string{
	"00:00:0c": "Cisco",    "00:1a:2b": "Cisco",      "00:50:b6": "Cisco",
	"00:50:56": "VMware",   "00:0c:29": "VMware",     "00:1c:14": "VMware",
	"00:1b:21": "Intel",    "00:1e:65": "Intel",      "00:08:74": "Dell",
	"00:14:22": "Dell",     "00:15:5d": "Microsoft Hyper-V",
	"00:03:ff": "Microsoft", "00:1d:d8": "Microsoft", "00:1c:42": "Lenovo",
	"00:24:d7": "TP-Link",  "50:c7:bf": "TP-Link",    "00:0a:5e": "3Com",
	"00:12:17": "3Com",     "00:1f:33": "Netgear",    "00:26:f2": "Netgear",
	"00:05:85": "Fortinet", "00:0e:8f": "Fortinet",   "70:4c:a5": "Fortinet",
	"00:22:10": "Juniper",  "28:8a:1c": "Juniper",    "04:4a:6c": "Ubiquiti",
	"f0:9f:c2": "Ubiquiti", "00:0f:20": "MikroTik",   "4c:5e:0c": "MikroTik",
	"00:1f:7b": "Aruba",    "00:24:6c": "Aruba",      "00:17:c5": "Meraki",
	"68:3a:35": "Meraki",   "00:0c:98": "HP",         "00:1b:78": "HP",
	"b8:27:eb": "Raspberry Pi", "dc:a6:32": "Raspberry Pi",
	"00:25:45": "Huawei",   "00:e0:fc": "Huawei",     "48:46:fb": "Huawei",
	"00:19:88": "ZTE",      "00:26:59": "Samsung",    "00:15:99": "Samsung",
	"00:1f:a7": "Sony",     "00:04:1f": "Sony",       "ac:22:0b": "LG",
	"00:50:c9": "LG",       "3c:5a:37": "Google",
	"00:11:32": "Synology", "00:09:6b": "IBM",        "00:03:ba": "Oracle",
}

// LookupVendor resolves the first 3 octets of a MAC address to a vendor name.
func LookupVendor(mac string) string {
	normalized := strings.ToLower(strings.ReplaceAll(
		strings.ReplaceAll(strings.ReplaceAll(mac, "-", ":"), ".", ":"), " ", ""))
	if len(normalized) < 8 {
		return ""
	}
	return ouiDB[normalized[:8]]
}
