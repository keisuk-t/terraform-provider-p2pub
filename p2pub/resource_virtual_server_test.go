package p2pub

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const smallestVirtualServerDefinition = `

resource "p2pub_virtual_server" "vm1" {
    type = "VB0-1"
    os_type = "Linux"
}

`

func TestVirtualServer(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: smallestVirtualServerDefinition,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"p2pub_virtual_server.vm1", "type", "VB0-1"),
					resource.TestCheckResourceAttr(
						"p2pub_virtual_server.vm1", "os_type", "Linux"),
				),
			},
		},
	})
}
