package crcontainer

import "testing"

func TestRegistryDomainForRegion(t *testing.T) {
	cases := []struct {
		region string
		want   string
	}{
		{"us-south", "us.icr.io"},
		{"eu-de", "de.icr.io"},
		{"eu-central", "de.icr.io"},
		{"jp-tok", "jp.icr.io"},
		{"global", "icr.io"},
		{"", defaultRegistryDomain},
		{"unknown-region", defaultRegistryDomain},
	}

	for _, tc := range cases {
		if got := registryDomainForRegion(tc.region); got != tc.want {
			t.Errorf("registryDomainForRegion(%q) = %q, want %q", tc.region, got, tc.want)
		}
	}
}

func TestContainerRegistryLoginMissingBinary(t *testing.T) {
	// containerBinary is a constant pointing at the real "container" binary
	// name; this test simply documents expected error wrapping behavior
	// when the underlying command fails. We cannot safely exec a fake
	// "container" binary here without modifying PATH, so we only assert
	// that a clearly non-existent binary produces a wrapped error.
	err := containerRegistryLoginWithBinary("does-not-exist-binary", "icr.io", "iambearer", "token")
	if err == nil {
		t.Fatal("expected an error when the container binary does not exist")
	}
}
