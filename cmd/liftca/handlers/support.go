package handlers

import (
	"fmt"
	"github.com/jeanfric/liftca/ht"
	"github.com/jeanfric/liftca"
	"path"
	"strconv"
)

const (
	CaFolder   = "ca"
	CertFolder = "cert"
)

func CAUrl(caSerial int64) string {
	return path.Join("/", CaFolder, strconv.FormatInt(caSerial, 10))
}

func CACRLURL(caSerial int64) string {
	return path.Join(CAUrl(caSerial), "crl")
}

func CertUrl(caSerial, certSerial int64) string {
	return path.Join("/", CaFolder, strconv.FormatInt(caSerial, 10), CertFolder, strconv.FormatInt(certSerial, 10))
}

func ObtainCA(store *liftca.Store, r *ht.Request) (*liftca.Parcel, *ht.Answer) {
	caID, err := r.VarInt64("ca_id")
	if err != nil {
		return nil, ht.Failure(err)
	}
	auth, found := store.Get(caID)
	if !found {
		return nil, ht.NotFound()
	}

	if _, found := store.GetParent(caID); found {
		return nil, ht.NotFound()
	}

	return auth, nil
}

func ObtainCAAndCert(store *liftca.Store, r *ht.Request) (*liftca.Parcel, *liftca.Parcel, *ht.Answer) {
	ca, answer := ObtainCA(store, r)
	if answer != nil {
		return nil, nil, answer
	}

	certID, err := r.VarInt64("cert_id")
	if err != nil {
		return nil, nil, ht.Failure(err)
	}
	cert, found := store.Get(certID)
	if !found {
		return nil, nil, ht.NotFound()
	}
	parent, _ := store.GetParent(certID)
	if parent != ca.SerialNumber() {
		return nil, nil, ht.Failure(fmt.Errorf("certificate %v does not belong to CA %v", certID, ca.SerialNumber()))
	}

	return ca, cert, nil
}
