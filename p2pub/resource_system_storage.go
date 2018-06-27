package p2pub

import (
	"time"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourceSystemStorage() *schema.Resource {
	return &schema.Resource{
		Create: resourceSystemStorageCreate,
		Read:   resourceSystemStorageRead,
		Update: resourceSystemStorageUpdate,
		Delete: resourceSystemStorageDelete,

		Timeouts: &schema.ResourceTimeout{
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

			//
			
			"root_ssh_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"root_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"userdata": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

//
// api call
//

func getSystemStorageInfo(api *p2pubapi.API, gis, iba string) (*protocol.SystemStorageGetResponse, error) {
	args := protocol.SystemStorageGet{
		GisServiceCode: gis,
		IbaServiceCode: iba,
	}
	var res = protocol.SystemStorageGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func setSystemStorageLabel(api *p2pubapi.API, gis, iba, label string) error {
	args := protocol.SystemStorageLabelSet{
		GisServiceCode: gis,
		IbaServiceCode: iba,
		Name: label,
	}
	var res = protocol.SystemStorageLabelSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	return nil
}

func setSSHKey(api *p2pubapi.API, gis, iba, key string) error {
	info, err := getSystemStorageInfo(api, gis, iba)
	if err != nil {
		return err
	}
	attachStatus := p2pubapi.NotAttached
	if info.ResourceStatus == p2pubapi.Attached.String() {
		attachStatus = p2pubapi.Attached
	}
	args := protocol.PublicKeyAdd{
		GisServiceCode: gis,
		IbaServiceCode: iba,
		PublicKey: key,
	}
	var res = protocol.PublicKeyAddResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitSystemStorage(api, gis, iba,
		p2pubapi.InService, attachStatus, TIMEOUT); err != nil {
		return err
	}
	return nil
}

func setPassword(api *p2pubapi.API, gis, iba, password string) error {
	info, err := getSystemStorageInfo(api, gis, iba)
	if err != nil {
		return err
	}
	attachStatus := p2pubapi.NotAttached
	if info.ResourceStatus == p2pubapi.Attached.String() {
		attachStatus = p2pubapi.Attached
	}
	args := protocol.PasswordSet{
		GisServiceCode: gis,
		IbaServiceCode: iba,
		Password: password,
	}
	var res = protocol.PasswordSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitSystemStorage(api, gis, iba,
		p2pubapi.InService, attachStatus, TIMEOUT); err != nil {
		return err
	}
	return nil
}

func setUserData(api *p2pubapi.API, gis, iba, userdata string) error {
	info, err := getSystemStorageInfo(api, gis, iba)
	if err != nil {
		return err
	}
	attachStatus := p2pubapi.NotAttached
	if info.ResourceStatus == p2pubapi.Attached.String() {
		attachStatus = p2pubapi.Attached
	}
	args := protocol.UserDataSet{
		GisServiceCode: gis,
		IbaServiceCode: iba,
		UserData: userdata,
	}
	var res = protocol.UserDataSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitSystemStorage(api, gis, iba,
		p2pubapi.InService, attachStatus, TIMEOUT); err != nil {
			return err
		}
	return nil
}

//
// reosurce operations
//

func resourceSystemStorageCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.SystemStorageAdd{
		GisServiceCode: gis,
		Type: d.Get("type").(string),
		StorageGroup: d.Get("storage_group").(string),
	}
	var res = protocol.SystemStorageAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	iba := res.ServiceCode

	if err := p2pubapi.WaitSystemStorage(api, gis, iba,
		p2pubapi.InService, p2pubapi.NotAttached, TIMEOUT); err != nil {
		return err
	}

	if d.Get("label") != nil && d.Get("label").(string) != "" {
		if err := setSystemStorageLabel(api, gis, iba, d.Get("label").(string)); err != nil {
			return err
		}
	}

	if d.Get("root_ssh_key") != nil && d.Get("root_ssh_key").(string) != "" {
		if err := setSSHKey(api, gis, iba, d.Get("root_ssh_key").(string)); err != nil {
			return err
		}
	}

	if d.Get("root_password") != nil && d.Get("root_password").(string) != "" {
		if err := setPassword(api, gis, iba, d.Get("root_password").(string)); err != nil {
			return err
		}
	}

	if d.Get("userdata") != nil {
		if err := setUserData(api, gis, iba, d.Get("userdata").(string)); err != nil {
			return err
		}
	}  	

	d.SetId(iba)

	return resourceSystemStorageRead(d, m)
}

func resourceSystemStorageRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	res, err := getSystemStorageInfo(api, gis, d.Id())
	if err != nil {
		return err
	}

	d.Set("type", res.Type)
	d.Set("storage_group", res.StorageGroup)
	d.Set("os_type", res.OSType)
	d.Set("storage_size", res.StorageSize)
	d.Set("label", res.Label)

	return nil
}

func resourceSystemStorageUpdate(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	if !d.HasChange("label") && !d.HasChange("root_ssh_key") && !d.HasChange("root_password") && !d.HasChange("userdata") {
		return nil
	}

	d.Partial(true)

	info, err := getSystemStorageInfo(api, gis, d.Id())
	// TODO: 
	if err != nil {
		return err
	}
	vm_stopped := false

	if d.HasChange("label") {
		if err := setSystemStorageLabel(api, gis, d.Id(), d.Get("label").(string)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	if d.HasChange("root_ssh_key") {
		if info.ResourceStatus == p2pubapi.Attached.String() && !vm_stopped {
			if err := power(api, gis, info.AttachedVirtualServer.ServiceCode, "Off"); err != nil {
				return err
			}
			vm_stopped = true
		}
		if err := setSSHKey(api, gis, d.Id(), d.Get("root_ssh_key").(string)); err != nil {
			return err
		}
		d.SetPartial("root_ssh_key")
	}

	if d.HasChange("root_password") {
		if info.ResourceStatus == p2pubapi.Attached.String() && !vm_stopped {
			if err := power(api, gis, info.AttachedVirtualServer.ServiceCode, "Off"); err != nil {
				return err
			}
			vm_stopped = true
		}
		if err := setPassword(api, gis, d.Id(), d.Get("root_password").(string)); err != nil {
			return err
		}
		d.SetPartial("root_password")
	}

	if d.HasChange("userdata") {
		if info.ResourceStatus == p2pubapi.Attached.String() && !vm_stopped {
			if err := power(api, gis, info.AttachedVirtualServer.ServiceCode, "Off"); err != nil {
				return err
			}
			vm_stopped = true
		}
		if err := setUserData(api, gis, d.Id(), d.Get("userdata").(string)); err != nil {
			return err
		}
		d.SetPartial("userdata")
	}	

	d.Partial(false)

	if vm_stopped {
		if err := power(api, gis, info.AttachedVirtualServer.ServiceCode, "On"); err != nil {
			return err
		}
	}
	
	return nil
}

func resourceSystemStorageDelete(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	if err := p2pubapi.WaitSystemStorage(api, gis, d.Id(),
		p2pubapi.InService, p2pubapi.NotAttached, TIMEOUT); err != nil {
		return err
	}

	args := protocol.SystemStorageCancel{
		GisServiceCode: gis,
		IbaServiceCode: d.Id(),
	}
	var res = protocol.SystemStorageCancel{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
