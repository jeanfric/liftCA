package handlers

import (
	"liftca"
	"strconv"
)

type JSONCAResponse struct {
	Self         string `json:"self"`
	SerialNumber string `json:"serialNumber"`
	Name         string `json:"name"`
	SubjectKeyID string `json:"subjectKeyID"`
	Visible      bool   `json:"visible"`
}

type JSONCARequest struct {
	Visible        bool   `json:"visible"`
	Name           string `json:"name"`
	PEMCertificate string `json:"pemCertificate"`
	PEMKey         string `json:"pemKey"`
	PEMKeyPassword string `json:"pemKeyPassword"`
}

type JSONCRLRequest struct {
	SerialNumber string `json:"serialNumber"`
}

type JSONCRLResponse struct {
	Self          string   `json:"self"`
	SerialNumbers []string `json:"serialNumbers"`
}

type JSONCertRequest struct {
	Host string `json:"host"`
}

type JSONCertResponse struct {
	Host           string `json:"host"`
	Self           string `json:"self"`
	SerialNumber   string `json:"serialNumber"`
	SubjectKeyID   string `json:"subjectKeyID"`
	AuthorityKeyID string `json:"authorityKeyID"`
}

func JSONCAResponseFromParcel(p *liftca.Parcel) *JSONCAResponse {
	return &JSONCAResponse{
		Name:         p.Certificate.Subject.CommonName,
		Self:         CAUrl(p.SerialNumber()),
		SerialNumber: strconv.FormatInt(p.SerialNumber(), 10),
		SubjectKeyID: p.SubjectKeyID(),
		Visible:      p.Visible,
	}
}

func JSONCertResponseFromParcel(caId int64, p *liftca.Parcel) *JSONCertResponse {
	return &JSONCertResponse{
		Host:           p.Host(),
		Self:           CertUrl(caId, p.SerialNumber()),
		SerialNumber:   strconv.FormatInt(p.SerialNumber(), 10),
		SubjectKeyID:   p.SubjectKeyID(),
		AuthorityKeyID: p.AuthorityKeyID(),
	}
}
