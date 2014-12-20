package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jeanfric/liftca"
	"github.com/jeanfric/liftca/cmd/liftca/handlers"
	"github.com/jeanfric/liftca/ht"
)

func main() {
	quit := make(chan bool)

	var addressArg string
	var storeFileArg string

	flag.StringVar(&addressArg, "a", ":8080", "listen address")
	flag.StringVar(&storeFileArg, "s", "store.gob", "path to state storage file")
	flag.Parse()

	storeFile := filepath.Clean(storeFileArg)
	backingFile, err := os.OpenFile(storeFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer backingFile.Close()
	store := liftca.LoadStore(backingFile)
	storeChanged := make(chan struct{})
	store.Updates(storeChanged)
	go func(c <-chan struct{}) {
		for {
			<-c
			store.DumpStore(backingFile)
		}
	}(storeChanged)

	r := ht.NewRouter()
	fileServer := ht.NewLoggingFileServer(http.Dir("static"))

	r.Handle("GET", "/ca", ht.NewHandler(store, handlers.GetCAs))
	r.Handle("POST", "/ca", ht.NewHandler(store, handlers.PostCA))
	r.Handle("GET", "/ca/{ca_id}-certificate.cer", ht.NewHandler(store, handlers.GetCACertificateCER))
	r.Handle("GET", "/ca/{ca_id}-crl.crl", ht.NewHandler(store, handlers.GetCACRLCER))
	r.Handle("GET", "/ca/{ca_id}-private-key.cer", ht.NewHandler(store, handlers.GetCAPrivateKeyCER))
	r.Handle("GET", "/ca/{ca_id}-certificate.pem", ht.NewHandler(store, handlers.GetCACertificatePEM))
	r.Handle("GET", "/ca/{ca_id}-certificate.pem.txt", ht.NewHandler(store, handlers.GetCACertificatePEMTXT))
	r.Handle("GET", "/ca/{ca_id}-private-key.pem", ht.NewHandler(store, handlers.GetCAPrivateKeyPEM))
	r.Handle("GET", "/ca/{ca_id}-private-key.pem.txt", ht.NewHandler(store, handlers.GetCAPrivateKeyPEMTXT))
	r.Handle("GET", "/ca/{ca_id}-crl.pem", ht.NewHandler(store, handlers.GetCACRLPEM))
	r.Handle("GET", "/ca/{ca_id}-crl.pem.txt", ht.NewHandler(store, handlers.GetCACRLPEMTXT))
	r.Handle("GET", "/ca/{ca_id}", ht.NewHandler(store, handlers.GetCA))
	r.Handle("GET", "/ca/{ca_id}/cert", ht.NewHandler(store, handlers.GetCerts))
	r.Handle("POST", "/ca/{ca_id}/cert", ht.NewHandler(store, handlers.PostCert))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-certificate.pem", ht.NewHandler(store, handlers.GetCertificatePEM))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-certificate.pem.txt", ht.NewHandler(store, handlers.GetCertificatePEMTXT))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-private-key.pem", ht.NewHandler(store, handlers.GetCertificatePrivateKeyPEM))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-private-key.pem.txt", ht.NewHandler(store, handlers.GetCertificatePrivateKeyPEMTXT))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-private-key.cer", ht.NewHandler(store, handlers.GetCertificatePrivateKeyCER))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}-certificate.cer", ht.NewHandler(store, handlers.GetCertificateCER))
	r.Handle("GET", "/ca/{ca_id}/cert/{cert_id}", ht.NewHandler(store, handlers.GetCert))
	r.Handle("POST", "/ca/{ca_id}/crl", ht.NewHandler(store, handlers.PostCRL))
	r.Handle("GET", "/ca/{ca_id}/crl", ht.NewHandler(store, handlers.GetCRL))
	r.Handle("DELETE", "/ca/{ca_id}/crl/{cert_id}", ht.NewHandler(store, handlers.DeleteCRL))
	r.Handle("GET", "/", fileServer)
	r.Handle("GET", "/{f}", fileServer)
	r.Handle("GET", "/js/{f}", fileServer)
	r.Handle("GET", "/fonts/{f}", fileServer)
	r.Handle("GET", "/css/{f}", fileServer)
	r.Handle("GET", "/img/{f}", fileServer)
	r.Handle("GET", "/partials/{f}", fileServer)

	s := &http.Server{
		Addr:           addressArg,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}
	go s.ListenAndServe()

	log.Printf("liftCA engaged at '%v', data file '%v'", addressArg, storeFile)
	<-quit
}
