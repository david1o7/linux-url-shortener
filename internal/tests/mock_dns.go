package tests

import "net"

type MockResolver struct{
	IPs []net.IP
	Err error
}

func (m *MockResolver) LookupIP(host string) ([]net.IP ,error){
	return m.IPs, m.Err
}
