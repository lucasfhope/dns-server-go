package main

import (
	"flag"
	"github.com/codecrafters-io/dns-server-starter-go/app/mydns"
)



func main() {
	resolver := flag.String("resolver", "", "The DNS resolver to forward queries to")
	flag.Parse()

	mydns.StartDNSServer(*resolver)
}

