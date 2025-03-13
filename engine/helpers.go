package engine

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"gocryptotrader/common/file"
	"gocryptotrader/log"
)

var (
	errCertExpired     = errors.New("gRPC TLS certificate has expired")
	errCertDataIsNil   = errors.New("gRPC TLS certificate PEM data is nil")
	errCertTypeInvalid = errors.New("gRPC TLS certificate type is invalid")
)

// CheckCerts checks and verifies RPC server certificates
func CheckCerts(targetDir string) error {
	if !file.Exists(targetDir) {
		log.Warnln(log.GRPCSys, "Target directory for certificates does not exist, creating..")
		err := os.MkdirAll(targetDir, file.DefaultPermissionOctal)
		if err != nil {
			return err
		}
	}

	certFile := filepath.Join(targetDir, "cert.pem")
	keyFile := filepath.Join(targetDir, "key.pem")

	if !file.Exists(certFile) || !file.Exists(keyFile) {
		log.Warnln(log.GRPCSys, "Certificate/key file(s) do not exist, creating...")
		return genCert(targetDir)
	}

	certData, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return errCertDataIsNil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	if time.Now().After(cert.NotAfter) {
		log.Warnln(log.GRPCSys, "Certificate has expired, regenerating...")
		return genCert(targetDir)
	}

	log.Debugln(log.GRPCSys, "Certificate and key files exist and are valid.")
	return nil
}

func genCert(targetDir string) error {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	host, err := os.Hostname()
	if err != nil {
		return err
	}

	dnsNames := []string{host}
	if host != "localhost" {
		dnsNames = append(dnsNames, "localhost")
	}

	ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.IsGlobalUnicast() {
				ipAddresses = append(ipAddresses, ipnet.IP)
			}
		}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"GoCryptoTrader"},
			CommonName:   host,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddresses,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	certData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if certData == nil {
		return errCertDataIsNil
	}

	b, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return err
	}

	keyData := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if keyData == nil {
		return errCertTypeInvalid
	}

	certFile := filepath.Join(targetDir, "cert.pem")
	keyFile := filepath.Join(targetDir, "key.pem")

	err = os.WriteFile(certFile, certData, file.DefaultPermissionOctal)
	if err != nil {
		return err
	}

	return os.WriteFile(keyFile, keyData, file.DefaultPermissionOctal)
}

func verifyCert(pemData []byte) error {
	var pemBlock *pem.Block
	pemBlock, _ = pem.Decode(pemData)
	if pemBlock == nil {
		return errCertDataIsNil
	}

	if pemBlock.Type != "CERTIFICATE" {
		return errCertTypeInvalid
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return err
	}

	if time.Now().After(cert.NotAfter) {
		return errCertExpired
	}
	return nil
}
