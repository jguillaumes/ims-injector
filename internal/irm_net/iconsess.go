package irm_net

import (
	"fmt"
	"net"
)

type IMSconSess struct {
	tcpAddr net.TCPAddr
	conn    *net.TCPConn
}

func NewIMSconSess(hostname string, port uint16) (*IMSconSess, error) {
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(ips[0])
	if ip == nil {
		return nil, fmt.Errorf("invalid hostname: %s", hostname)
	}
	tcpaddr := &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}
	newConn := &IMSconSess{
		tcpAddr: *tcpaddr,
		conn:    nil,
	}
	return newConn, nil
}

func (s *IMSconSess) Connect() error {
	conn, err := net.DialTCP("tcp", nil, &s.tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d: %v", s.tcpAddr.IP, s.tcpAddr.Port, err)
	}
	s.conn = conn
	return nil
}

func (s *IMSconSess) Close() error {
	if s.conn != nil {
		err := s.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection: %v", err)
		}
		s.conn = nil
	}
	return nil
}
