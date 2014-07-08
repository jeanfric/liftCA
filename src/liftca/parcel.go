package liftca

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"time"
)

var keyBarrel = NewBarrel(8, 1024)

type Parcel struct {
	Visible             bool
	Certificate         *x509.Certificate
	PrivateKey          *rsa.PrivateKey
	DERCertificateBytes []byte
}

func (p *Parcel) SerialNumber() int64 {
	return p.Certificate.SerialNumber.Int64()
}

func (p *Parcel) SubjectKeyID() string {
	return fmt.Sprintf("% x", p.Certificate.SubjectKeyId)
}

func (p *Parcel) AuthorityKeyID() string {
	return fmt.Sprintf("% x", p.Certificate.AuthorityKeyId)
}

func (p *Parcel) PublicKey() *rsa.PublicKey {
	return &p.PrivateKey.PublicKey
}

func (p *Parcel) Host() string {
	if len(p.Certificate.IPAddresses) > 0 {
		return p.Certificate.IPAddresses[0].String()
	} else {
		return p.Certificate.Subject.CommonName
	}
}

func importCAFromPEM(visible bool, serial int64, certificate, privKey, keyPassword []byte) (p *Parcel, err error) {
	certBlock, _ := pem.Decode(certificate)
	if certBlock == nil {
		err = fmt.Errorf("Invalid PEM block")
		return
	}
	if certBlock.Type != "CERTIFICATE" {
		err = fmt.Errorf("Invalid PEM block; should be a CERTIFICATE block")
		return
	}

	keyBlock, _ := pem.Decode(privKey)
	if keyBlock == nil {
		err = fmt.Errorf("Invalid PEM block")
		return
	}
	if keyBlock.Type != "RSA PRIVATE KEY" {
		err = fmt.Errorf("Invalid PEM block; should be a RSA PRIVATE KEY block")
		return
	}
	keyBlockBytes := keyBlock.Bytes
	if len(keyPassword) > 0 {
		keyBlockBytes, err = x509.DecryptPEMBlock(keyBlock, keyPassword)
		if err != nil {
			return
		}
	}

	certs, err := x509.ParseCertificates(certBlock.Bytes)
	if err != nil {
		return
	}
	if len(certs) != 1 {
		err = fmt.Errorf("Found too many certificates in PEM block")
		return
	}

	cert := &x509.Certificate{
		Subject:               certs[0].Subject,
		IsCA:                  true,
		MaxPathLen:            certs[0].MaxPathLen,
		BasicConstraintsValid: true,
		SerialNumber:          big.NewInt(serial),
		SubjectKeyId:          certs[0].SubjectKeyId,
		AuthorityKeyId:        certs[0].AuthorityKeyId,
		NotBefore:             time.Date(2010, 12, 31, 23, 59, 59, 0, time.UTC),
		NotAfter:              time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlockBytes)
	if err != nil {
		return
	}

	p = &Parcel{
		Visible:             visible,
		Certificate:         cert,
		PrivateKey:          key,
		DERCertificateBytes: certs[0].Raw,
	}

	return
}

func makeCAParcel(visible bool, serial int64, name string) (p *Parcel, err error) {
	ser := big.NewInt(serial)
	cert := makeCertTemplate(true, "", name, ser)
	key := keyBarrel.GetKey()

	hasher := sha1.New()
	bytes, err := asn1.Marshal(key.PublicKey)
	if err != nil {
		return
	}
	hasher.Write(bytes)
	h := hasher.Sum(nil)
	cert.SubjectKeyId = h
	cert.AuthorityKeyId = h

	raw, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return
	}

	p = &Parcel{
		Visible:             visible,
		Certificate:         cert,
		PrivateKey:          key,
		DERCertificateBytes: raw,
	}
	return
}

func makeParcel(visible bool, serial int64, ca *Parcel, host string) (p *Parcel, err error) {
	ser := big.NewInt(serial)
	cert := makeCertTemplate(
		false,
		host,
		ca.Certificate.Subject.CommonName,
		ser)
	key := keyBarrel.GetKey()
	hasher := sha1.New()
	bytes, err := asn1.Marshal(key.PublicKey)
	hasher.Write(bytes)
	h := hasher.Sum(nil)
	cert.SubjectKeyId = h
	cert.AuthorityKeyId = ca.Certificate.SubjectKeyId

	raw, err := x509.CreateCertificate(rand.Reader, cert, ca.Certificate, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		return
	}

	p = &Parcel{
		Visible:             visible,
		Certificate:         cert,
		PrivateKey:          key,
		DERCertificateBytes: raw,
	}
	return
}

func makeCertTemplate(isCA bool, host, name string, serial *big.Int) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore:             time.Date(2010, 12, 31, 23, 59, 59, 0, time.UTC),
		NotAfter:              time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	if !isCA {
		cert.Subject.CommonName = host
		ip := net.ParseIP(host)
		if ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		}
	}
	if isCA {
		cert.IsCA = true
		cert.MaxPathLen = 16
		cert.KeyUsage |= x509.KeyUsageCertSign
	}
	return cert
}

func (ca *Parcel) derCRLBytes(revoked []int64) ([]byte, error) {
	rev := make([]pkix.RevokedCertificate, 0)
	for _, val := range revoked {
		r := pkix.RevokedCertificate{
			SerialNumber:   big.NewInt(val),
			RevocationTime: time.Now(),
		}
		rev = append(rev, r)
	}

	return ca.Certificate.CreateCRL(
		rand.Reader,
		ca.PrivateKey,
		rev,
		time.Now(),
		time.Now().Add(time.Duration(1)*time.Minute))
}

func (ca *Parcel) DERCRL(revoked []int64) (io.Reader, error) {
	crlBytes, err := ca.derCRLBytes(revoked)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(crlBytes), nil
}

func (ca *Parcel) PEMCRL(revoked []int64) (io.Reader, error) {
	crlBytes, err := ca.derCRLBytes(revoked)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	}
	data := pem.EncodeToMemory(block)
	return bytes.NewBuffer(data), nil
}

func (p *Parcel) DERCertificate() io.Reader {
	return bytes.NewBuffer(p.DERCertificateBytes)
}

func (p *Parcel) DERPrivateKey() io.Reader {
	data := x509.MarshalPKCS1PrivateKey(p.PrivateKey)
	return bytes.NewBuffer(data)
}

func (p *Parcel) PEMCertificate() io.Reader {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: p.DERCertificateBytes,
	}
	data := pem.EncodeToMemory(block)
	return bytes.NewBuffer(data)
}

func (p *Parcel) PEMPrivateKey() io.Reader {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(p.PrivateKey),
	}
	data := pem.EncodeToMemory(block)
	return bytes.NewBuffer(data)
}
