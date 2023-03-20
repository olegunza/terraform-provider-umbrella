package umbrellaprovider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSiteResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSiteResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("umbrella_site.test", "name", "one"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("umbrella_site.test", "site_id"),
					resource.TestCheckResourceAttrSet("umbrella_site.test", "origin_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "umbrella_site.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: testAccSiteResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("umbrella_example.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSiteResourceConfig(sitename string) string {
	return fmt.Sprintf(`
	terraform {
		required_providers {
		  umbrella = {
			source  = "OZAKHARO-M-22C5.local/local/umbrella"
			version = "1.0.0"
		   
		  }
		}
	  }
	  
resource "umbrella_site" "test" {
  name = %[1]q
}
`, sitename)
}
