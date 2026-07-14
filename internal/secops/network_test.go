package secops

import (
	"testing"
)

func TestTLSCertificateStruct(t *testing.T) {
	cert := TLSCertificate{
		Subject:    "CN=example.com",
		Issuer:     "CN=Let's Encrypt",
		NotAfter:   "2027-01-01",
		KeySize:    256,
		IsExpiring: false,
		DaysLeft:   170,
	}
	if cert.IsExpiring {
		t.Error("expected IsExpiring to be false")
	}
	if cert.DaysLeft != 170 {
		t.Errorf("expected DaysLeft 170, got %d", cert.DaysLeft)
	}
}

func TestPublicExposureStruct(t *testing.T) {
	pe := PublicExposure{
		Port:        22,
		Protocol:    "tcp",
		ProcessName: "sshd",
		Severity:    "critical",
	}
	if pe.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", pe.Severity)
	}
}
