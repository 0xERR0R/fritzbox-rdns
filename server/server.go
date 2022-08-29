package server

import (
	"fmt"
	"github.com/0xERR0R/fritzbox-rdns/cache"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

type Server struct {
	dnsServers []*dns.Server
	cache      *cache.NamesCache
}

func NewServer(cache *cache.NamesCache) (*Server, error) {
	const address = ":53"
	dnsServers := []*dns.Server{
		createUDPServer(address),
		createTCPServer(address),
	}

	s := &Server{
		dnsServers: dnsServers,
		cache:      cache,
	}

	for _, server := range s.dnsServers {
		handler := server.Handler.(*dns.ServeMux)
		handler.HandleFunc(".", s.OnRequest)
	}

	return s, nil
}

func (s *Server) OnRequest(rw dns.ResponseWriter, msg *dns.Msg) {
	response := new(dns.Msg)
	if msg.Question[0].Qtype == dns.TypePTR {
		name := msg.Question[0].Name

		log.Debug().Str("name", name).Msg("received request")

		ip := ExtractAddressFromReverse(name)

		clientName := s.cache.Get(ip)
		response.SetRcode(msg, dns.RcodeSuccess)
		response.MsgHdr.RecursionAvailable = msg.MsgHdr.RecursionDesired
		if clientName != "" {
			rr, err := dns.NewRR(fmt.Sprintf("%s\t%d\tIN\t%s\t%s", name, 300, "PTR", dns.Fqdn(clientName)))
			if err != nil {
				response.SetRcode(msg, dns.RcodeServerFailure)
			} else {
				response.Answer = []dns.RR{rr}
			}
		}
	}
	if err := rw.WriteMsg(response); err != nil {
		log.Err(err).Msg("can't write response")
	}
}

func createUDPServer(address string) *dns.Server {
	const maxUDPSize = 65535

	return &dns.Server{
		Addr:    address,
		Net:     "udp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			log.Info().Msgf("UDP server is up and running on: '%s'", address)
		},
		UDPSize: maxUDPSize,
	}
}

func createTCPServer(address string) *dns.Server {
	return &dns.Server{
		Addr:    address,
		Net:     "tcp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			log.Info().Msgf("TCP server is up and running on: '%s'", address)
		},
	}
}

// Start starts the server
func (s *Server) Start() {
	log.Info().Msg("Starting server")

	for _, srv := range s.dnsServers {
		srv := srv

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Fatal().Msgf("start %s listener failed: %v", srv.Net, err)
			}
		}()
	}
}

// Stop stops the server
func (s *Server) Stop() {
	log.Info().Msg("Stopping server")

	for _, server := range s.dnsServers {
		if err := server.Shutdown(); err != nil {
			log.Fatal().Msgf("stop %s listener failed: %v", err)
		}
	}
}
