package fritzbox

type SessionInfo struct {
	SID       string `xml:"SID"`
	Challenge string `xml:"Challenge"`
}

type Network struct {
	Data struct {
		ActiveDevices []struct {
			Uid  string `json:"UID"`
			Mac  string `json:"mac"`
			Name string `json:"name"`
		} `json:"active"`
	} `json:"data"`
}

type DeviceDetails struct {
	Data struct {
		Vars struct {
			Dev struct {
				Ipv4 IpAddressInfo `json:"ipv4"`
				Ipv6 IpAddressInfo `json:"ipv6"`
			} `json:"dev"`
		} `json:"vars"`
	} `json:"data"`
}

type IpAddressInfo struct {
	Current IpAddress   `json:"current"`
	IpList  []IpAddress `json:"ipList"`
}

type IpAddress struct {
	Ip string `json:"ip"`
}
