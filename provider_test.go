package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sops": providerserver.NewProtocol6WithError(New("test")()),
}

var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"sops": providerserver.NewProtocol6WithError(New("test")()),
	"echo": echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
}

// Shared age test key pairs. Generated for tests only — never use outside this suite.
const testAgePublicKey = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"
const testAgeSecretKey = "AGE-SECRET-KEY-18Z8D6LS5LCAZWERTYMK87NQ0N0ZEX5T50NZ9Q5XVPES2VRPWTC4SYAY5AT"

const testAgePublicKey2 = "age1qxcacnkalvdm9ky9thl8t4nwcwp5kpra6uf0w6zqn8ep7fd4q5uq655l2z"
const testAgeSecretKey2 = "AGE-SECRET-KEY-1A5M38GF766C2W0VGELHYT39J0MF2MUZ7PCQQS9587SHL4Z43X2ZS4D4DXG"
