package goclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type TLSOptions struct {
	CaCertFile         string
	InsecureSkipVerify bool
}

func NewTLSConfig(conf *TLSOptions) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: conf.InsecureSkipVerify,
	}

	if conf.CaCertFile == "" || conf.InsecureSkipVerify {
		return tlsConfig, nil
	}

	caCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile(conf.CaCertFile)

	if err != nil {
		return nil, err
	}

	if !caCertPool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append cert from PEM file : %s", conf.CaCertFile)
	}

	tlsConfig.RootCAs = caCertPool
	return tlsConfig, nil
}
