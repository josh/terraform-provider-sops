package main

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAgeRecipient = "age1j7ce327ke8t905hr4ve97xh4jr5ujauq59nxxkr3tnz9pty78p6q26hnd0"

func TestAccEncryptDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigBasic(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_NestedStructure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigNested(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_MultipleRecipients(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigMultipleRecipients(testAgeRecipient, testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_EdgeCases(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigEdgeCases(testAgeRecipient),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_OutputTypeYAML(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptDataSourceConfigWithOutputType(testAgeRecipient, "yaml"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sops_encrypt.test", "output"),
					resource.TestCheckResourceAttr("data.sops_encrypt.test", "output_type", "yaml"),
				),
			},
		},
	})
}

func TestAccEncryptDataSource_InvalidInputTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccEncryptDataSourceConfigInvalidArray(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got \[\]interface \{\}\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptDataSourceConfigInvalidString(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got string\. SOPS can only encrypt JSON\s+objects\.`),
			},
			{
				Config:      testAccEncryptDataSourceConfigInvalidNumber(testAgeRecipient),
				ExpectError: regexp.MustCompile(`Input must be a map/object, got float64\. SOPS can only encrypt JSON\s+objects\.`),
			},
		},
	})
}

func testAccEncryptDataSourceConfigBasic(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigNested(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    database = {
      host     = "localhost"
      password = "secret123"
    }
    api_keys = ["key1", "key2"]
  }
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigMultipleRecipients(ageRecipient1, ageRecipient2 string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "multi-recipient-value"
  }
  age = ["` + ageRecipient1 + `", "` + ageRecipient2 + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidArray(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = ["not", "a", "map"]
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidString(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = "not a map"
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigInvalidNumber(ageRecipient string) string {
	return `
data "sops_encrypt" "test" {
  input = 42
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigEdgeCases(ageRecipient string) string {
	return `
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
  age = ["` + ageRecipient + `"]
}
`
}

func testAccEncryptDataSourceConfigWithOutputType(ageRecipient, outputType string) string {
	return `
data "sops_encrypt" "test" {
  input = {
    secret = "value"
    key    = "data"
  }
  age = ["` + ageRecipient + `"]
  output_type = "` + outputType + `"
}
`
}
