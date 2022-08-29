package fritzbox

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"time"
)

type Service struct {
	session    SessionInfo
	client     *http.Client
	fbHost     string
	fbUser     string
	fbPassword string
}

func NewService(host, user, password string) *Service {
	return &Service{
		client:     createHttpClient(),
		fbHost:     host,
		fbPassword: password,
		fbUser:     user,
	}
}

func (s *Service) FetchIpDeviceNames() (map[string]string, error) {
	err := s.performLogin()
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	log.Debug().Msg("fetching network details")
	network, err := s.fetchNetwork()
	if err != nil {
		return nil, fmt.Errorf("fetch network failed: %w", err)
	}

	log.Info().Msgf("found %d active devices", len(network.Data.ActiveDevices))
	ipToDeviceNameMap := make(map[string]string, 0)

	log.Debug().Msg("fetching all device details")

	for _, dev := range network.Data.ActiveDevices {
		log.Debug().Str("name", dev.Name).Str("uuid", dev.Uid).Msg("fetching device details")
		details, err := s.fetchDeviceDetails(dev.Uid)
		if err != nil {
			return nil, fmt.Errorf("fetching device details for device '%s' failed: %w", dev.Name, err)
		}
		var deviceIps []string

		deviceIps = append(deviceIps, details.Data.Vars.Dev.Ipv4.Current.Ip, details.Data.Vars.Dev.Ipv6.Current.Ip)

		for _, addr := range details.Data.Vars.Dev.Ipv6.IpList {
			deviceIps = append(deviceIps, addr.Ip)
		}

		log.Debug().Str("name", dev.Name).Strs("fetched_IPs", deviceIps).Msg("fetched data")

		for _, ip := range deviceIps {
			ipToDeviceNameMap[ip] = dev.Name
		}
	}

	return ipToDeviceNameMap, nil

}

func (s *Service) fetchNetwork() (Network, error) {
	resp, err := s.client.PostForm(s.fbHost+"/data.lua", map[string][]string{
		"xhr":   {"1"},
		"sid":   {s.session.SID},
		"lang":  {"de"},
		"page":  {"netDev"},
		"xhrId": {"all"},
	})
	if err != nil {
		return Network{}, fmt.Errorf("can't get network page: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Network{}, fmt.Errorf("can't read body: %w", err)
	}
	network := Network{}

	err = json.Unmarshal(body, &network)
	if err != nil {
		return Network{}, fmt.Errorf("can't unmarshal body: %w", err)
	}

	return network, nil
}

func (s *Service) fetchDeviceDetails(uuid string) (DeviceDetails, error) {
	resp, err := s.client.PostForm(s.fbHost+"/data.lua", map[string][]string{
		"xhr":   {"1"},
		"sid":   {s.session.SID},
		"lang":  {"de"},
		"page":  {"edit_device"},
		"xhrId": {"all"},
		"dev":   {uuid},
	})
	if err != nil {
		return DeviceDetails{}, fmt.Errorf("can't get device details page: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DeviceDetails{}, fmt.Errorf("can't read body: %w", err)
	}

	deviceDetails := DeviceDetails{}

	err = json.Unmarshal(body, &deviceDetails)
	if err != nil {
		return DeviceDetails{}, fmt.Errorf("can't unmarshal body: %w", err)
	}

	return deviceDetails, nil
}

func (s *Service) performLogin() error {
	log.Debug().Msg("performing login")
	session, err := fetchSessionInfo(s.client, s.fbHost+"/login_sid.lua")
	if err != nil {
		return err
	}

	response := buildResponse(session.Challenge, s.fbPassword)

	session, err = fetchSessionInfo(s.client, s.fbHost+"/login_sid.lua?&username="+s.fbUser+"&response="+response)
	if err != nil {
		return err
	}
	if session.SID == "0000000000000000" {
		return errors.New("login not successful")
	}

	s.session = session

	log.Debug().Str("SID", session.SID).Msg("login successful")

	return nil
}

func createHttpClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   2 * time.Minute,
	}
	return client
}

func fetchSessionInfo(client *http.Client, url string) (SessionInfo, error) {
	resp, err := client.Get(url)
	if err != nil {
		return SessionInfo{}, err
	}

	defer resp.Body.Close() // nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SessionInfo{}, err
	}

	var sessionInfo SessionInfo
	err = xml.Unmarshal(body, &sessionInfo)
	if err != nil {
		return SessionInfo{}, err
	}

	return sessionInfo, nil
}

func buildResponse(challenge string, password string) string {
	challengePassword := utf8ToUtf16(challenge + "-" + password)

	md5Response := md5.Sum([]byte(challengePassword)) // nolint: gas

	return challenge + "-" + fmt.Sprintf("%x", md5Response)
}

func utf8ToUtf16(input string) string {
	e := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	t := e.NewEncoder()

	outstr, _, err := transform.String(t, input)
	if err != nil {
		log.Fatal().Err(err).Msgf("can't convert utf8 to utf16")
	}

	return outstr
}
