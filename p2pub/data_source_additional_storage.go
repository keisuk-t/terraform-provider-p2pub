package p2pub

import (
	"time"
	"log"
	"regexp"
	"errors"

	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAdditionalStorage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAdditionalStorageRead,

		Timeouts: &schema.ResourceTimeout{
    		        Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
			"most_recent": &schema.Schema{
				Type: schema.TypeBool,
				Optional: true,
				Default: false,
			},
			"service_code": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},

			//
			//

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"created_at": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func getAdditionalStorageList(api *p2pubapi.API, gis string) (*protocol.StorageListGetResponse, error) {
	args := protocol.StorageListGet{
		GisServiceCode: gis,
	}
	var res = protocol.StorageListGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func dataSourceAdditionalStorageRead(d *schema.ResourceData, m interface{}) error {

	if d.Get("filter") == nil && d.Get("service_code") == "" {
		return errors.New("filter or service_code is required")
	}

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	storages, err := getAdditionalStorageList(api, gis)
	if err != nil {
		return err
	}

	var matches []int
	for idx, storage := range storages.AdditionalStorageList {
		if d.Get("service_code") == storage.ServiceCode {
			matches = []int{ idx }
			break
		}
		match := true
		for _, f := range d.Get("filter").([]interface{}) {
			filter := f.(map[string]interface{})
			switch filter["name"] {
			case "os_type":
				match = match && filter["value"] == storage.OSType
			case "label":
				matched , _ := regexp.MatchString(filter["value"].(string), storage.Label)
				match = match && matched
			case "type":
				match = match && filter["value"] == storage.Type
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
		return errors.New("no volumes matched")
	}

	if len(matches) >= 2 {
		if !d.Get("most_recent").(bool) {
			return errors.New("two or more volumes matched. please narrow down")
		}
		// TODO:
		picked := 0
		last_modified := "ZZZ"
		for _, idx := range matches {
			if storages.AdditionalStorageList[idx].StartDate < last_modified {
				picked = idx
				last_modified = storages.AdditionalStorageList[idx].StartDate
			}
		}
		matches = []int{ picked }
	}

	ans := storages.AdditionalStorageList[matches[0]]
	d.SetId(ans.ServiceCode)
	d.Set("os_type", ans.OSType)
	d.Set("created_at", ans.StartDate)
	d.Set("label", ans.Label)
	d.Set("type", ans.Type)
	d.Set("storage_size", ans.StorageSize)
	d.Set("storage_group", ans.StorageGroup)
	d.Set("service_code", ans.ServiceCode)
	
	return nil
}
