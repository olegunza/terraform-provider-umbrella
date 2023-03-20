package umbrellaprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVADataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccVADataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.umbrella_va.test", "vas.health", "ok"),
				),
			},
		},
	})
}

const testAccVADataSourceConfig = `
terraform {
	required_providers {
	  umbrella = {
		source  = "OZAKHARO-M-22C5.local/local/umbrella"
		version = "1.0.0"
	   
	  }
	}
  }
data "umbrella_va" "test" {
}`
