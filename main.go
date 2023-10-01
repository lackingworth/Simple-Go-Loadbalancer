package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() 		string
	IsAlive() 		bool
	Serve(w http.ResponseWriter, r *http.Request)		
}

type simpleServer struct {
	addr 		string
	proxy 		*httputil.ReverseProxy
}

type LoadBalancer struct {
	port		string
	rrc			int
	servers		[]Server
}

func (s *simpleServer) Address() string {return s.addr}

func (s *simpleServer) IsAlive() bool {return true}

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.rrc % len(lb.servers)]
	for !server.IsAlive() {
		lb.rrc++
		server = lb.servers[lb.rrc % len(lb.servers)]
	}
	lb.rrc++
	return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwarding request to address: %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr: 		addr,
		proxy: 		httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalanser(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port: 		port,
		rrc: 		0,
		servers: 	servers,
	}
}


func handleErr(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	servers := []Server {
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.duckduckgo.com"),
	}
	lb := NewLoadBalanser("8000", servers)
	handleRedirect :=  func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Serving request at 'localhost: %s'\n", lb.port)
	http.ListenAndServe(":" + lb.port, nil)
}