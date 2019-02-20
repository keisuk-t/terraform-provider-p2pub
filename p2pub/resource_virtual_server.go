package p2pub

import (
	"time"
	"log"
	"strings"
	
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func resourceVirtualServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualServerCreate,
		Read:   resourceVirtualServerRead,
		Update: resourceVirtualServerUpdate,
		Delete: resourceVirtualServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(5 * time.Minute),
			Update:  schema.DefaultTimeout(5 * time.Minute),
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
			"os_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"server_group": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"label": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"category": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"storage_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"boot": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"pci_slot": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"service_code": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"os_type": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"type": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},
			"network_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac_address": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ip_address_list": &schema.Schema{
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ipv4_address": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ipv4_type": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ipv6_address": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ipv6_type": &schema.Schema{
										Type: schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						"label": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"service_code": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"network_type": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ipv6_enabled": &schema.Schema{
							Type: schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},
			"serverspec_memory": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"serverspec_cpu": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// define resource linkage attributes/flags below.
			// these attributes are used for 
			
			"system_storage": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
			},
			"data_storage": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: DATA_STORAGE_MAX_ATTACH_COUNT,
				Optional: true,
			},
			"private_network": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: PRIVATE_NETWORK_MAX_ATTACH_COUNT,
				Optional: true,
			},
			"enable_global_ip": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
		},
	}
}

func needUpdateAttributes(d *schema.ResourceData) bool {
	return d.Get("enable_global_ip") != nil ||
		d.Get("private_network") != nil ||
		d.Get("system_storage") != nil ||
		d.Get("data_storage") != nil
}

func bootable(d *schema.ResourceData) bool {
	// TODO: 
	return d.Get("system_storage") != nil && d.Get("system_storage") != ""
}

//
// api call shorthands
//

func getVMInfo(api *p2pubapi.API, gis, ivm string) (*protocol.VMGetResponse, error) {
	args := protocol.VMGet{
		GisServiceCode: gis,
 		IvmServiceCode: ivm,
	}
	var res = protocol.VMGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func power(api *p2pubapi.API, gis, ivm, onoff string, timeout time.Duration) error {
	if info, err := getVMInfo(api, gis, ivm); err != nil {
		return err
	} else {
		if onoff == "On" && info.ResourceStatus == p2pubapi.Running.String() ||
			onoff == "Off" && info.ResourceStatus == p2pubapi.Stopped.String() {
			return nil
		}
	}
	args := protocol.VMPower{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
		Power: onoff,
	}
	var res = protocol.VMPowerResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	state := p2pubapi.Running
	if onoff == "Off" {
		state = p2pubapi.Stopped
	}
	// should i use terrafrom helper function ?
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, state, timeout); err != nil {
		return err;
	}
	return nil
}

func allocateGlobalIP(api *p2pubapi.API, gis, ivm string, timeout time.Duration) (string, error) {
	args := protocol.GlobalAddressAllocate{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
	}
	var res = protocol.GlobalAddressAllocateResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return "", err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return "", err
	}
	return res.IPv4.IpAddress, nil
}

func releaseGlobalIP(api *p2pubapi.API, gis, ivm string, timeout time.Duration) error {
	args := protocol.GlobalAddressRelease{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
	}
	var res = protocol.GlobalAddressReleaseResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func attachBootDevice(api *p2pubapi.API, gis, ivm, system_storage string, timeout time.Duration) error {

	args := protocol.BootDeviceStorageConnect{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
	}

	if strings.HasPrefix(system_storage, "iba") {
		args.IbaServiceCode = system_storage
	} else if strings.HasPrefix(system_storage, "ica") {
		args.IcaServiceCode = system_storage
	}

	var res = protocol.BootDeviceStorageConnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err;
	}
	return nil
}

func detachBootDevice(api *p2pubapi.API, gis, ivm string, timeout time.Duration) error {
	args := protocol.BootDeviceStorageDisconnect{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
	}
	var res = protocol.BootDeviceStorageDisconnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func attachDataDevice(api *p2pubapi.API, gis, ivm, data_storage string, timeout time.Duration) error {
	args := protocol.DataDeviceStorageConnect{
			GisServiceCode: gis,
			IvmServiceCode: ivm,
	}
	
	if strings.HasPrefix(data_storage, "ibg") {
		args.IbgServiceCode = data_storage
	} else if strings.HasPrefix(data_storage, "ibb") {
		args.IbbServiceCode = data_storage
	} else if strings.HasPrefix(data_storage, "iba") {
		args.IbaServiceCode = data_storage
	} else if strings.HasPrefix(data_storage, "icg") {
		args.IcgServiceCode = data_storage
	} else if strings.HasPrefix(data_storage, "icb") {
		args.IcbServiceCode = data_storage
	} else if strings.HasPrefix(data_storage, "ica") {
		args.IcaServiceCode = data_storage
	}
	
	var res = protocol.DataDeviceStorageConnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func detachDataDevice(api *p2pubapi.API, gis, ivm, pci string, timeout time.Duration) error {
	args := protocol.DataDeviceStorageDisconnect{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
		PciSlot: pci,
	}
	var res = protocol.DataDeviceStorageDisconnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func attachPrivateNetowrk(api *p2pubapi.API, gis, ivm, ivl string, timeout time.Duration) error {
	args := protocol.PrivateNetworkConnect{
		GisServiceCode: gis,
		IvlServiceCode: ivl,
		IvmServiceCode: ivm,
	}
	var res = protocol.PrivateNetworkConnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func detachPrivateNetwork(api *p2pubapi.API, gis, ivm, mac string, timeout time.Duration) error {
	args := protocol.PrivateNetworkDisconnect{
		GisServiceCode: gis,
		MacAddress: mac,
		IvmServiceCode: ivm,
	}
	var res = protocol.PrivateNetworkDisconnectResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err
	}
	return nil
}

func setLabel(api *p2pubapi.API, gis, ivm, label string) error {
	args := protocol.VMLabelSet{
		GisServiceCode: gis,
		IvmServiceCode: ivm,
		Name: label,
	}
	var res = protocol.VMLabelSetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}
	return nil
}

//
// virtual server CRUD definition
//

func resourceVirtualServerCreate(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode
	timeout := d.Timeout(schema.TimeoutCreate)

	log.Printf("[DEBUG] p2pub: create virtual server resource on %s", gis)

	args := protocol.VMAdd{
		GisServiceCode: gis,
		Type: d.Get("type").(string),
		OSType: d.Get("os_type").(string),
		ServerGroup: d.Get("server_group").(string),
	}
	var res = protocol.VMAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	ivm := res.ServiceCode

	if err := p2pubapi.WaitVM(api, gis, ivm, p2pubapi.InService, p2pubapi.Stopped, timeout); err != nil {
		return err;
	}

	if d.Get("label") != nil && d.Get("label") != "" {
		if err := setLabel(api, gis, ivm, d.Get("label").(string)); err != nil {
			return err
		}
	}

	if d.Get("system_storage") != nil && d.Get("system_storage") != "" {
		if err := attachBootDevice(api, gis, ivm, d.Get("system_storage").(string), timeout); err != nil {
			return err
		}
	}

	if d.Get("data_storage") != nil {
		for _, ibg := range d.Get("data_storage").(*schema.Set).List() {
			if err := attachDataDevice(api, gis, ivm, ibg.(string), timeout); err != nil {
				return err
			}
		}
	}

	if d.Get("private_network") != nil {
		for _, ivl := range d.Get("private_network").(*schema.Set).List() {
			if err := attachPrivateNetowrk(api, gis, ivm, ivl.(string), timeout); err != nil {
				return err
			}
		}
	}

	if d.Get("enable_global_ip") != nil && d.Get("enable_global_ip").(bool) {
		global_ip, err := allocateGlobalIP(api, gis, ivm, timeout);
		if err != nil {
			return err
		}
		d.SetConnInfo(map[string]string {
			"type": "ssh",
			"user": "root",
			"host": global_ip,
		})
	} else {
		for _, net := range res.NetworkList {
			if net.NetworkType == "PrivateStandard" {
				d.SetConnInfo(map[string]string {
					"type": "ssh",
					"user": "root",
					"host": net.IpAddressList[0].IPv4.IpAddress,
				})
				break
			}
		}
	}

	if bootable(d) {
		if err := power(api, gis, ivm, "On", timeout); err != nil {
			return err;
		}
	}

	d.SetId(ivm)

	return resourceVirtualServerRead(d, m)
}

func resourceVirtualServerRead(d *schema.ResourceData, m interface{}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	res, err := getVMInfo(api, gis, d.Id())
	if err != nil {
		return err
	}

	d.Set("server_group", res.ServerGroup)
	d.Set("label", res.Label)
	d.Set("category", res.Category)
	d.Set("serverspec_cpu", res.ServerSpec.CPU)
	d.Set("serverspec_memory", res.ServerSpec.Memory)

	// set network_list
	network_list := make([]map[string]interface{}, 0)
	for _, elm := range res.NetworkList {
		n := make(map[string]interface{})
		n["mac_address"] = elm.MacAddress
		n["label"] = elm.Label
		n["service_code"] = elm.ServiceCode
		n["network_type"] = elm.NetworkType
		n["ipv6_enabled"] = elm.IPv6Enabled
		addrs := make([]map[string]string, 0)
		for _, addr := range elm.IpAddressList {
			addrs = append(addrs, map[string]string{
				"ipv4_address": addr.IPv4.IpAddress,
				"ipv4_type": addr.IPv4.Type,
				"ipv6_address": addr.IPv6.IpAddress,
				"ipv6_type": addr.IPv6.Type,
			})
		}
		n["ip_address_list"] = addrs
		network_list = append(network_list, n)
	}
	if err := d.Set("network_list", network_list); err != nil {
		return err
	}

	// set storage_list
	storage_list := make([]map[string]interface{}, 0)
	for _, elm := range res.StorageList {
		storage_list = append(storage_list, map[string]interface{}{
			"boot": elm.Boot,
			"pci_slot": elm.PciSlot,
			"service_code": elm.ServiceCode,
			"os_type": elm.OSType,
			"type": elm.Type,
		})
	}
	if err := d.Set("storage_list", storage_list); err != nil {
		return err
	}

	return nil
}

func resourceVirtualServerUpdate(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode
	timeout := d.Timeout(schema.TimeoutUpdate)

	d.Partial(true)

	stopped := false

	if d.HasChange("label") {
		if err := setLabel(api, gis, d.Id(), d.Get("label").(string)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	if d.HasChange("type") {
		log.Printf("[DEBUG] p2pub: %s - change VM type to %s", d.Id(), d.Get("type"))
		if !stopped {
			if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
				return err;
			}
			stopped = true
		}
		type_args := protocol.VMItemChange{
			GisServiceCode: gis,
			IvmServiceCode: d.Id(),
			Type: d.Get("type").(string),
		}
		var type_res = protocol.VMItemChangeResponse{}
		if err := p2pubapi.Call(*api, type_args, &type_res); err != nil {
			return err
		}
		if err := p2pubapi.WaitVM(api, gis, d.Id(), p2pubapi.InService, p2pubapi.Stopped, TIMEOUT); err != nil {
			return err
		}
		d.SetPartial("type")
	}

	if d.HasChange("system_storage") {
		log.Printf("[DEBUG] p2pub: %s - change boot device %s", d.Id(), d.Get("system_storage"))
		if !stopped {
			if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
				return err
			}
			stopped = true
		}
		if err := detachBootDevice(api, gis, d.Id(), timeout); err != nil {
			return err
		}
		if d.Get("system_storage") != nil && d.Get("system_storage") != "" {
			if err := attachBootDevice(api, gis, d.Id(), d.Get("system_storage").(string), timeout); err != nil {
				return err
			}
		}
		d.SetPartial("system_storage")
	}

	if d.HasChange("data_storage") {
		log.Printf("[DEBUG] p2pub: %s - change data device %s", d.Id(), d.Get("data_storage"))
		if !stopped {
			if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
				return err
			}
			stopped = true
		}
		if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
			return err
		}
		for _, elm := range d.Get("storage_list").([]interface{}) {
			boot := elm.(map[string]interface{})["boot"].(string)
			pci_slot := elm.(map[string]interface{})["pci_slot"].(string)
			if boot == "Yes" {
				continue;
			}
			if err := detachDataDevice(api, gis, d.Id(), pci_slot, timeout); err != nil {
				return err
			}
			log.Printf("[DEBUG] detach data device %v", pci_slot)
		}
		for _, ibg := range d.Get("data_storage").(*schema.Set).List() {
			if err := attachDataDevice(api, gis, d.Id(), ibg.(string), timeout); err != nil {
				return err
			}
		}
		d.SetPartial("data_storage")
	}

	if d.HasChange("private_network") {
		log.Printf("[DEBUG] p2pub: %s - change private network %s", d.Id(), d.Get("private_network"))
		if !stopped {
			if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
				return err
			}
			stopped = true
		}
		for _, elm := range d.Get("network_list").([]interface{}) {
			t := elm.(map[string]interface{})["network_type"].(string)
			mac_address := elm.(map[string]interface{})["mac_address"].(string)
			if t == "Private" {
				if err := detachPrivateNetwork(api, gis, d.Id(), mac_address, timeout); err != nil {
					return err
				}
			}
		}
		for _, ivl := range d.Get("private_network").(*schema.Set).List() {
			if err := attachPrivateNetowrk(api, gis, d.Id(), ivl.(string), timeout); err != nil {
				return err
			}
		}
		d.SetPartial("private_network")
	}

	if d.HasChange("enable_global_ip") {
		log.Printf("[DEBUG] p2pub: %s - change global ip address assignment %s", d.Id(), d.Get("enable_global_ip"))
		if !stopped {
			if err := power(api, gis, d.Id(), "Off", timeout); err != nil {
				return err
			}
			stopped = true
		}
		if d.Get("enable_global_ip") != nil && d.Get("enable_global_ip").(bool) {
			if _, err := allocateGlobalIP(api, gis, d.Id(), timeout); err != nil {
				return err
			}
		} else {
			if err := releaseGlobalIP(api, gis, d.Id(), timeout); err != nil {
				return err
			}
		}
		d.SetPartial("enable_global_ip")
	}

	d.Partial(false)

	if stopped && bootable(d) {
		if err := power(api, gis, d.Id(), "On", timeout); err != nil {
			return err;
		}
	}
	
	return nil
}

func resourceVirtualServerDelete(d *schema.ResourceData, m interface {}) error {

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.VMCancel{
		GisServiceCode: gis,
 		IvmServiceCode: d.Id(),
	}
	var res = protocol.VMCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")
	
	return nil
}
