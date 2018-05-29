package p2pub

import (
	"time"
	"errors"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
)

func dataSourceVirtualServer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVirtualServerRead,

		Timeouts: &schema.ResourceTimeout{
 		        Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"os_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			//
			//

			"filter": &schema.Schema{
				Type: schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
						"value": &schema.Schema{
							Type: schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"service_code": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getVMList(api *p2pubapi.API, gis string) (*protocol.VMListGetResponse, error) {
	args := protocol.VMListGet{
		GisServiceCode: gis,
	}
	var res = protocol.VMListGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func dataSourceVirtualServerRead(d *schema.ResourceData, m interface{}) error {

	if d.Get("filter") == nil && d.Get("service_code") == "" {
		return errors.New("filter or service code is required")
	}

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	vms, err := getVMList(api, gis)
	if err != nil {
		return err
	}

	var matches []int
	for idx, vm := range vms.VirtualServerList {
		if d.Get("service_code") == vm.ServiceCode {
			matches = []int{ idx }
		}
		match := true
		for _, f := range d.Get("filter").([]interface{}) {
			filter := f.(map[string]interface{})
			switch filter["name"] {
			case "os_type":
				match = match && filter["value"] == vm.OSType
			case "label":
				matched , _ := regexp.MatchString(filter["value"].(string), vm.Label)
				match = match && matched
			case "category":
				match = match && filter["value"] == vm.Category
			case "type":
				match = match && filter["value"] == vm.Type
			default:
				log.Printf("[ERROR] filter by '%s' not supported", filter["name"])
				return errors.New("invalid filter")
			}
		}
		if match {
			matches = append(matches, idx)
		}
	}

	if len(matches) == 0 {
		return errors.New("no virtual servers matched")
	}

	if len(matches) >= 2 {
		if !d.Get("most_recent").(bool) {
			return errors.New("two or more virtual servers matched. please narrow down")
		}
		picked := 0
		last_modified := "ZZZ"
		for _, idx := range matches {
			if vms.VirtualServerList[idx].StartDate < last_modified {
				picked = idx
				last_modified = vms.VirtualServerList[idx].StartDate
			}
		}
		matches = []int{ picked }
	}

	ans := vms.VirtualServerList[matches[0]]

	d.SetId(ans.ServiceCode)

	return resourceVirtualServerRead(d, m)
}
