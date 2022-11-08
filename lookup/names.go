package lookup

import (
	"context"
	"net"
	"time"

	"github.com/0xERR0R/fritzbox-rdns/fritzbox"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	ctx = context.Background()

	// keep ip -> name pairs for one week in redis
	ttl = time.Duration(7 * 24 * time.Hour)

	updatePeriod = 15 * time.Minute
)

type NamesLookupService struct {
	service *fritzbox.Service
	rdb     *redis.Client
}

func NewNamesLookupService(service *fritzbox.Service, rdb *redis.Client) *NamesLookupService {
	n := &NamesLookupService{
		service: service,
		rdb:     rdb,
	}

	go n.periodicUpdate()

	n.Update()

	return n
}

func (n *NamesLookupService) periodicUpdate() {
	ticker := time.NewTicker(updatePeriod)
	defer ticker.Stop()

	for {
		<-ticker.C
		n.Update()
	}
}

func (n *NamesLookupService) Get(ip net.IP) string {
	val, err := n.rdb.Get(ctx, ip.String()).Result()

	if err == redis.Nil {

		return ""
	} else if err != nil {
		log.Err(err).Msg("can't fetch value from redis")

		return ""
	}

	return val
}

func (n *NamesLookupService) Update() {
	log.Info().Msg("updating cache...")

	names, err := n.service.FetchIpDeviceNames()
	if err != nil {
		log.Err(err).Msg("fetching of ip name information failed")
	}

	for ip, name := range names {
		if ip != "" {
			parsedIp := net.ParseIP(ip)

			if parsedIp != nil {
				err := n.rdb.Set(ctx, parsedIp.String(), name, ttl).Err()

				if err != nil {
					log.Err(err).Msg("can't store value in redis")
				}
			} else {
				log.Warn().Str("IP", ip).Msg("IP invalid?")
			}
		}
	}

	size, err := n.rdb.DBSize(ctx).Result()

	if err != nil {
		log.Err(err).Msg("can't retrieve the db size from redis")
	}

	log.Info().Int64("db_total_size", size).Msg("updating cache done!")
}
