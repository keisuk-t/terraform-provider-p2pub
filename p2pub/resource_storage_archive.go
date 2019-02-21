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
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceStorageArchiveCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.StorageArchiveAdd{
		GisServiceCode: gis,
		ArchiveSize: d.Get("archive_size").(string),
	}
	var res = protocol.StorageArchiveAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId(res.ServiceCode)
	d.Set("archive_size", res.ArchiveSize)

	return nil
}

func resourceStorageArchiveRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.StorageArchiveGet{
		GisServiceCode: gis,
		IarServiceCode: d.Id(),
	}
	var res = protocol.StorageArchiveGetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.Set("archive_size", res.ArchiveSize)

	return nil
}

func resourceStorageArchiveUpdate(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	d.Partial(true)

	if d.HasChange("archive_size") {
		args := protocol.StorageArchiveSizeChange{
			GisServiceCode: gis,
			IarServiceCode: d.Id(),
			ArchiveSize: d.Get("archive_size").(string),
		}
		var res = protocol.StorageArchiveSizeChangeResponse{}

		if err := p2pubapi.Call(*api, args, &res); err != nil {
			return err
		}
		d.SetPartial("archive_size")
	}

	d.Partial(false)

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
