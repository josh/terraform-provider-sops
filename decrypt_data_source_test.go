package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAgePublicKey = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"
const testAgeSecretKey = "AGE-SECRET-KEY-18Z8D6LS5LCAZWERTYMK87NQ0N0ZEX5T50NZ9Q5XVPES2VRPWTC4SYAY5AT"

func testAccDecryptPreCheck(t *testing.T) {
	testAccPreCheck(t)
}

func TestAccEncryptDecryptIntegration_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigBasic(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.secret", "value"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.key", "data"),
				),
			},
		},
	})
}

func TestAccEncryptDecryptIntegration_NestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigNested(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.database.%"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.api_keys.#"),
				),
			},
		},
	})
}

func TestAccEncryptDecryptIntegration_EdgeCases(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigEdgeCases(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.string_val", "hello world"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.int_val", "42"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.float_val", "3.14159"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.bool_true", "true"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.bool_false", "false"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.array_strings.#", "3"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.nested_deep.%"),
				),
			},
		},
	})
}

func testAccEncryptDecryptIntegrationConfigBasic(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age_recipients = ["` + ageRecipient + `"]
}

data "sops_decrypt" "test" {
  input = data.sops_encrypt.test.output
}
`
}

func testAccEncryptDecryptIntegrationConfigNested(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    database = {
      host     = "localhost"
      password = "secret123"
    }
    api_keys = ["key1", "key2"]
  }
  age_recipients = ["` + ageRecipient + `"]
}

data "sops_decrypt" "test" {
  input = data.sops_encrypt.test.output
}
`
}

func testAccEncryptDecryptIntegrationConfigEdgeCases(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    string_val    = "hello world"
    int_val       = 42
    float_val     = 3.14159
    bool_true     = true
    bool_false    = false
    null_val      = null
    array_strings = ["one", "two", "three"]
    array_numbers = [1, 2, 3, 4, 5]
    array_mixed   = ["string", 123, true, 45.67]
    empty_array   = []
    empty_object  = {}
    nested_deep   = {
      level1 = {
        level2 = {
          level3 = {
            value = "deep"
          }
        }
      }
    }
  }
  age_recipients = ["` + ageRecipient + `"]
}

data "sops_decrypt" "test" {
  input = data.sops_encrypt.test.output
}
`
}

func TestAccEncryptDecryptIntegration_YAMLInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigYAMLInput(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.secret", "yaml-value"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.key", "yaml-data"),
				),
			},
		},
	})
}

func TestAccEncryptDecryptIntegration_YAMLNestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigYAMLNested(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.database.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.database.host", "yaml-localhost"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.database.port", "5432"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.api_keys.#", "3"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.api_keys.0", "yaml-key1"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.nested.level1.level2.%"),
				),
			},
		},
	})
}

func TestAccEncryptDecryptIntegration_YAMLEdgeCases(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigYAMLEdgeCases(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.string_val", "yaml string"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.int_val", "123"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.float_val", "45.67"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.bool_true", "true"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.bool_false", "false"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.array_strings.#", "3"),
				),
			},
		},
	})
}

func TestAccEncryptDecryptIntegration_DefaultInputTypeJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccDecryptPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDecryptIntegrationConfigDefaultInputType(testAgePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttrSet("data.sops_decrypt.test", "output.%"),
					resource.TestCheckResourceAttr("data.sops_decrypt.test", "output.secret", "default-value"),
				),
			},
		},
	})
}

func testAccEncryptDecryptIntegrationConfigYAMLInput(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    secret = "yaml-value"
    key    = "yaml-data"
  }
  age_recipients = ["` + ageRecipient + `"]
  output_type    = "yaml"
}

data "sops_decrypt" "test" {
  input      = data.sops_encrypt.test.output
  input_type = "yaml"
}
`
}

func testAccEncryptDecryptIntegrationConfigYAMLNested(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    database = {
      host     = "yaml-localhost"
      port     = 5432
      password = "yaml-secret123"
    }
    api_keys = ["yaml-key1", "yaml-key2", "yaml-key3"]
    nested = {
      level1 = {
        level2 = {
          value = "deep-yaml"
        }
      }
    }
  }
  age_recipients = ["` + ageRecipient + `"]
  output_type    = "yaml"
}

data "sops_decrypt" "test" {
  input      = data.sops_encrypt.test.output
  input_type = "yaml"
}
`
}

func testAccEncryptDecryptIntegrationConfigYAMLEdgeCases(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    string_val    = "yaml string"
    int_val       = 123
    float_val     = 45.67
    bool_true     = true
    bool_false    = false
    null_val      = null
    array_strings = ["a", "b", "c"]
    array_numbers = [10, 20, 30]
    empty_array   = []
    empty_object  = {}
  }
  age_recipients = ["` + ageRecipient + `"]
  output_type    = "yaml"
}

data "sops_decrypt" "test" {
  input      = data.sops_encrypt.test.output
  input_type = "yaml"
}
`
}

func testAccEncryptDecryptIntegrationConfigDefaultInputType(ageRecipient string) string {
	return `
provider "sops" {
  age_identity_value = "` + testAgeSecretKey + `"
}

data "sops_encrypt" "test" {
  input = {
    secret = "default-value"
  }
  age_recipients = ["` + ageRecipient + `"]
}

data "sops_decrypt" "test" {
  input = data.sops_encrypt.test.output
}
`
}
