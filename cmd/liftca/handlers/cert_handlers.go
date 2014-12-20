package handlers

import (
	"github.com/jeanfric/liftca/ht"
	"github.com/jeanfric/liftca"
)

func GetCertificatePEM(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/x-pem-file", cert.PEMCertificate())
}

func GetCertificatePEMTXT(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("text/plain", cert.PEMCertificate())
}

func GetCertificatePrivateKeyPEM(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/x-pem-file", cert.PEMPrivateKey())
}

func GetCertificatePrivateKeyPEMTXT(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("text/plain", cert.PEMPrivateKey())
}

func GetCertificatePrivateKeyCER(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/pkix-cert", cert.DERPrivateKey())
}

func GetCertificateCER(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/pkix-cert", cert.DERCertificate())
}

func GetCerts(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	children, _ := store.GetChildren(ca.SerialNumber())
	response := make([]JSONCertResponse, 0)
	for _, s := range children {
		cert, _ := store.Get(s)
		response = append(response, *JSONCertResponseFromParcel(ca.SerialNumber(), cert))
	}
	return ht.JSONDocument(response)
}

func PostCert(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	certReq := &JSONCertRequest{}
	err := r.BodyAsJSON(certReq)
	if err != nil {
		return ht.Failure(err)
	}
	id, err := store.Add(true, ca.SerialNumber(), certReq.Host)
	if err != nil {
		return ht.Failure(err)
	}
	return ht.RedirectTo(CertUrl(ca.SerialNumber(), id))
}

func GetCert(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	return ht.JSONDocument(JSONCertResponseFromParcel(ca.SerialNumber(), cert))
}
