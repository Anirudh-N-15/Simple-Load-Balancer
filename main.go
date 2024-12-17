package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	isAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type SimpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	port string
	RoundRobinCount int
	servers []Server
}

func NewLoadBalancer(port string,servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port : port,
		RoundRobinCount: 0,
		servers: servers,
	}
}

func newSimpleServer(addr string) *SimpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &SimpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *SimpleServer) Address() string {
	return s.addr
}

func (s * SimpleServer) isAlive() bool {
	return true 
}

func (s *SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb * LoadBalancer) getNextAvailableServer() Server{
	server := lb.servers[lb.RoundRobinCount % len(lb.servers)]

	for !server.isAlive(){
		server = lb.servers[lb.RoundRobinCount % len(lb.servers)]
	}
	lb.RoundRobinCount++ 
	return server
}

func (lb * LoadBalancer) serveProxy(rw http.ResponseWriter,req *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwarding Request to address %q\n",targetServer.Address())
	targetServer.Serve(rw,req)
}


func main() {
	servers := []Server{
		newSimpleServer("https://github.com"),
		newSimpleServer("https://www.gitlab.com"),
		newSimpleServer("https://www.duckduckgo.com"),
	}

	lb := NewLoadBalancer("8000", servers)

	handleRedirect := func(rw http.ResponseWriter, req *http.Request){
		lb.serveProxy(rw, req)
	}
	http.HandleFunc("/",handleRedirect)

	fmt.Printf("Serving requests at 'localhost:%s'\n",lb.port)
	http.ListenAndServe(":"+lb.port,nil)
}
