package p2pub

import (
	"time"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourcePrivateNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateNetworkCreate,
		Read:   resourcePrivateNetworkRead,
		Update: resourcePrivateNetworkUpdate,
		Delete: resourcePrivateNetworkDelete,

		Timeouts: &schema.ResourceTimeout{
		        Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"network_address": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

//
// api call
//

func waitPrivateNetwork(api p2pubapi.API, gis string, ivl string) error {
	for {
		args := protocol.PrivateNetworkVGet{
			GisServiceCode: gis,
			IvlServiceCode: ivl,
		}
		var res = protocol.PrivateNetworkVGetResponse{}
		if err := p2pubapi.Call(api, args, &res); err != nil {
			return err
		}

		if res.ContractStatus == "InService" {
			break
		}

		time.Sleep(15 * time.Second)
	}
	return nil
}

func setPrivateNetworkLabel(api *p2pubapi.API, gis, ivl, label string) error {
	args := protocol.PrivateNetworkVLabelSet{
		GisServiceCode: gis,
		IvlServiceCode: ivl,
		Name: label,
	}
	var res = protocol.PrivateNetworkVLabelSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	return nil
}

//
// resource operations
//

func resourcePrivateNetworkCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode
	
	args := protocol.PrivateNetworkVAdd{
		GisServiceCode: gis,
	}
	var res = protocol.PrivateNetworkVAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	ivl := res.ServiceCode

	if err := waitPrivateNetwork(*api, gis, ivl); err != nil {
		return err
	}

	if d.Get("label") != nil {
		if err := setPrivateNetworkLabel(api, gis, ivl, d.Get("label").(string)); err != nil {
			return err
		}
	}

	d.SetId(ivl)

	return resourcePrivateNetworkRead(d, m)
}

func resourcePrivateNetworkRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.PrivateNetworkVGet{
		GisServiceCode: gis,
		IvlServiceCode: d.Id(),
	}
	var res = protocol.PrivateNetworkVGetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.Set("label", res.Label)
	d.Set("network_address", res.NetworkAddress)

	return nil
}

func resourcePrivateNetworkUpdate(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	d.Partial(true)

	if d.HasChange("label") {
		if err := setPrivateNetworkLabel(api, gis, d.Id(), d.Get("label").(string)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	d.Partial(false)

	return nil
}

func resourcePrivateNetworkDelete(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode
	
	args := protocol.PrivateNetworkVCancel{
		GisServiceCode: gis,
		IvlServiceCode: d.Id(),
	}
	var res = protocol.PrivateNetworkVCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
