package p2pub

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const smallestSystemStorageDefinition = `

resource "p2pub_system_storage" "storage1" {
    type = "S30GB_CENTOS7_64"
}

`

func TestSystemStorage(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: smallestSystemStorageDefinition,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"p2pub_system_storage", "type", "S30GB_CENTOS7_64"),
				),
			},
		},
	})
}
