package p2pub

import (
	"time"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourceStorageArchive() *schema.Resource {
	return &schema.Resource{
		Create: resourceStorageArchiveCreate,
		Read:   resourceStorageArchiveRead,
		Update: resourceStorageArchiveUpdate,
		Delete: resourceStorageArchiveDelete,

		Timeouts: &schema.ResourceTimeout{
		        Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"archive_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceStorageArchiveCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.StorageArchiveAdd{
		GisServiceCode: gis,
	}
	var res = protocol.StorageArchiveAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId(res.ServiceCode)

	return nil
}

func resourceStorageArchiveRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceStorageArchiveUpdate(d *schema.ResourceData, m interface {}) error {
	return nil
}

func resourceStorageArchiveDelete(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.StorageArchiveCancel{
		GisServiceCode: gis,
		IarServiceCode: d.Id(),
	}
	var res = protocol.StorageArchiveCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
