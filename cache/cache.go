package cache

import (
	"github.com/0xERR0R/fritzbox-rdns/fritzbox"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type NamesCache struct {
	service *fritzbox.Service
	lru     *lru.Cache
}

func NewCache(service *fritzbox.Service) *NamesCache {
	l, _ := lru.New(1000)
	nc := &NamesCache{
		service: service,
		lru:     l,
	}

	go nc.periodicUpdate()

	nc.Update()

	return nc
}

func (nc *NamesCache) periodicUpdate() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		nc.Update()
	}
}

func (nc *NamesCache) Get(ip net.IP) string {
	value, ok := nc.lru.Get(ip.String())
	if ok {
		return value.(string)
	}

	return ""
}

func (nc *NamesCache) Update() {
	log.Info().Msg("updating cache...")

	names, err := nc.service.FetchIpDeviceNames()
	if err != nil {
		log.Err(err).Msg("fetching of ip name information failed")
	}

	for ip, name := range names {
		if ip != "" {
			parsedIp := net.ParseIP(ip)

			if parsedIp != nil {
				nc.lru.Add(parsedIp.String(), name)
			} else {
				log.Warn().Str("IP", ip).Msg("IP invalid?")
			}
		}
	}

	log.Info().Int("cache_element_count", nc.lru.Len()).Msg("updating cache done!")

}
