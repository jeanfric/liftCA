package handlers

import (
	"fmt"
	"github.com/jeanfric/liftca/ht"
	"github.com/jeanfric/liftca"
	"strconv"
)

func GetCACertificateCER(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/pkix-cert", ca.DERCertificate())
}

func GetCACRLCER(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	revoked := store.GetRevokedChildren(ca.SerialNumber())
	crl, err := ca.DERCRL(revoked)
	if err != nil {
		return ht.Failure(err)
	}
	return ht.Read("application/pkix-crl", crl)
}

func GetCAPrivateKeyCER(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/pkix-cert", ca.DERPrivateKey())
}

func GetCACertificatePEM(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/x-pem-file", ca.PEMCertificate())
}

func GetCACertificatePEMTXT(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("text/plain", ca.PEMCertificate())
}

func GetCAPrivateKeyPEM(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("application/x-pem-file", ca.PEMPrivateKey())
}

func GetCAPrivateKeyPEMTXT(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	return ht.Read("text/plain", ca.PEMPrivateKey())
}

func GetCACRLPEM(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	revoked := store.GetRevokedChildren(ca.SerialNumber())
	crl, err := ca.PEMCRL(revoked)
	if err != nil {
		return ht.Failure(err)
	}
	return ht.Read("application/x-pem-file", crl)
}

func GetCACRLPEMTXT(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	revoked := store.GetRevokedChildren(ca.SerialNumber())
	crl, err := ca.PEMCRL(revoked)
	if err != nil {
		return ht.Failure(err)
	}
	return ht.Read("text/plain", crl)
}

func GetCRL(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	revoked := store.GetRevokedChildren(ca.SerialNumber())
	output := make([]string, len(revoked))
	for i, e := range revoked {
		output[i] = strconv.FormatInt(e, 10)
	}

	return ht.JSONDocument(&JSONCRLResponse{
		Self: CACRLURL(ca.SerialNumber()),
		SerialNumbers: output,
	})
}

func PostCRL(store *liftca.Store, r *ht.Request) *ht.Answer {
	req := &JSONCRLRequest{}
	r.BodyAsJSON(req)
	certID, err := strconv.ParseInt(req.SerialNumber, 10, 64)
	if err != nil {
		return ht.Failure(err)
	}
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	if p, _ := store.GetParent(certID); p != ca.SerialNumber() {
		return ht.Failure(fmt.Errorf("certificate %v does not belong to CA %v", certID, ca.SerialNumber()))
	}
	store.SetRevoked(certID, true)
	return ht.RedirectTo(CACRLURL(ca.SerialNumber()))
}

func DeleteCRL(store *liftca.Store, r *ht.Request) *ht.Answer {
	_, cert, answer := ObtainCAAndCert(store, r)
	if answer != nil {
		return answer
	}
	store.SetRevoked(cert.SerialNumber(), false)
	return ht.NoContent()
}

func GetCAs(store *liftca.Store, r *ht.Request) *ht.Answer {
	response := make([]JSONCAResponse, 0)
	for _, s := range store.GetCAs() {
		auth, _ := store.Get(s)
		if auth.Visible {
			response = append(response, *JSONCAResponseFromParcel(auth))
		}
	}
	return ht.JSONDocument(response)
}

func GetCA(store *liftca.Store, r *ht.Request) *ht.Answer {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return answer
	}
	auth, _ := store.Get(ca.SerialNumber())
	return ht.JSONDocument(*JSONCAResponseFromParcel(auth))
}

func PostCA(store *liftca.Store, r *ht.Request) *ht.Answer {
	caReq := &JSONCARequest{}
	err := r.BodyAsJSON(caReq)
	if err != nil {
		return ht.Failure(err)
	}
	var id int64
	if caReq.PEMCertificate != "" || caReq.PEMKey != "" || caReq.PEMKeyPassword != "" {
		id, err = store.AddExistingCA(caReq.Visible, []byte(caReq.PEMCertificate), []byte(caReq.PEMKey), []byte(caReq.PEMKeyPassword))
	} else {
		id, err = store.AddCA(caReq.Visible, caReq.Name)
	}

	if err != nil {
		return ht.Failure(err)
	}
	return ht.RedirectTo(CAUrl(id))
}
