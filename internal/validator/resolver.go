package validator

import "net"

type DNSResolver interface{
	LookupIP(host string) ([]net.IP, error)
}

type RealResolver struct{}

func (r *RealResolver) LookupIP(host string) ([]net.IP, error){
	return net.LookupIP(host)
}