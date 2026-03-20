package revocation

import (
	"testing"
)

func TestInfoArchival_AddCRL_AddOCSP(t *testing.T) {
	var ia InfoArchival
	if err := ia.AddCRL([]byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := ia.AddOCSP([]byte{4, 5}); err != nil {
		t.Fatal(err)
	}
	if len(ia.CRL) != 1 || len(ia.CRL[0].FullBytes) != 3 {
		t.Fatalf("CRL: %+v", ia.CRL)
	}
	if len(ia.OCSP) != 1 || len(ia.OCSP[0].FullBytes) != 2 {
		t.Fatalf("OCSP: %+v", ia.OCSP)
	}
}
