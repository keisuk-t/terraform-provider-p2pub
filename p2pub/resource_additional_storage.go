package p2pub

import (
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourceAdditionalStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdditionalStorageCreate,
		Read:   resourceAdditionalStorageRead,
		Update: resourceAdditionalStorageUpdate,
		Delete: resourceAdditionalStorageDelete,

		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"storage_group": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"os_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"storage_size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"encryption": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

//
// api
//

func getAdditionalStorageInfo(api *p2pubapi.API, gis, ib string) (*protocol.StorageGetResponse, error) {
	args := protocol.StorageGet{
		GisServiceCode:     gis,
		StorageServiceCode: ib,
	}
	var res = protocol.StorageGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func setAdditionalStorageLabel(api *p2pubapi.API, gis, ib, label string) error {
	args := protocol.StorageLabelSet{
		GisServiceCode:     gis,
		StorageServiceCode: ib,
		Name:               label,
	}
	var res = protocol.StorageLabelSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	return nil
}

func isExtendedAdditionalStorage(stype string) bool {
	if strings.Index(stype, "BX") == 0 || strings.Index(stype, "GX") == 0 {
		return true
	}

	return false
}

//
// resource operations
//

func resourceAdditionalStorageCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.StorageAdd{
		GisServiceCode: gis,
		Type:           d.Get("type").(string),
		StorageGroup:   d.Get("storage_group").(string),
	}

	if isExtendedAdditionalStorage(d.Get("type").(string)) {
		if d.Get("encryption") != nil && d.Get("encryption").(string) != "" {
			args.Encryption = d.Get("encryption").(string)
		} else {
			args.Encryption = "No"
		}
	}

	var res = protocol.StorageAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	ib := res.ServiceCode

	if err := p2pubapi.WaitDataStorage(api, gis, ib,
		p2pubapi.InService, p2pubapi.NotAttached, d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	if d.Get("label") != nil && d.Get("label").(string) != "" {
		if err := setAdditionalStorageLabel(api, gis, ib, d.Get("label").(string)); err != nil {
			return err
		}
	}

	d.SetId(ib)

	return resourceAdditionalStorageRead(d, m)
}

func resourceAdditionalStorageRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	res, err := getAdditionalStorageInfo(api, gis, d.Id())
	if err != nil {
		return err
	}

	d.Set("type", res.Type)
	d.Set("storage_group", res.StorageGroup)
	d.Set("os_type", res.OSType)
	d.Set("storage_size", res.StorageSize)
	d.Set("label", res.Label)
	d.Set("mode", res.Mode)

	if isExtendedAdditionalStorage(res.Type) {
		d.Set("encryption", res.Encryption)
	}

	return nil
}

func resourceAdditionalStorageUpdate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	d.Partial(true)

	if d.HasChange("label") {
		if err := setAdditionalStorageLabel(api, gis, d.Id(), d.Get("label").(string)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	d.Partial(false)

	return nil
}

func resourceAdditionalStorageDelete(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	if err := p2pubapi.WaitDataStorage(api, gis, d.Id(),
		p2pubapi.InService, p2pubapi.NotAttached, d.Timeout(schema.TimeoutDefault)); err != nil {
		return err
	}

	args := protocol.StorageCancel{
		GisServiceCode:     gis,
		StorageServiceCode: d.Id(),
	}
	var res = protocol.StorageCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
