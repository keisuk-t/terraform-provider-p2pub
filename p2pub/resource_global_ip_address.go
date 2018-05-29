package p2pub

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourceGlobalIPAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlobalIPAddressCreate,
		Read:   resourceGlobalIPAddressRead,
		Update: resourceGlobalIPAddressUpdate,
		Delete: resourceGlobalIPAddressDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"address_num": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceGlobalIPAddressCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.GlobalAddressVAdd{
		GisServiceCode: gis,
		AddressNum: d.Get("address_num").(string),
	}
	var res = protocol.GlobalAddressVAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId(res.ServiceCode)

	return nil
}

func resourceGlobalIPAddressRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.GlobalAddressVGet{
		GisServiceCode: gis,
		IgaServiceCode: d.Id(),
	}
	var res = protocol.GlobalAddressVGetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.Set("address_num", res.AddressNum)

	return nil
}

func resourceGlobalIPAddressUpdate(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	d.Partial(true)

	if d.HasChange("address_num") {

		args := protocol.GlobalAddressVAddIPAddressNumChange{
			GisServiceCode: gis,
			IgaServiceCode: d.Id(),
			AddressNum: d.Get("address_num").(string),
		}
		var res = protocol.GlobalAddressVAddIPAddressNumChangeResponse{}

		if err := p2pubapi.Call(*api, args, &res); err != nil {
			return err
		}
		
		d.SetPartial("address_num")
	}

	d.Partial(false)

	return nil
}

func resourceGlobalIPAddressDelete(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.GlobalAddressVCancel{
		GisServiceCode: gis,
		IgaServiceCode: d.Id(),
	}
	var res = protocol.GlobalAddressVCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	return nil
}
