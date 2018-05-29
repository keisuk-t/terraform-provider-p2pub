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

func dataSourceCustomOSImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomOSImageRead,

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
			
			"os_type": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_at": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"label": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"source": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_id": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_size": &schema.Schema{
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
	}
}

//
// api call
//

func getContract(api *p2pubapi.API, gis string) (*protocol.P2PUBContractGetForSAResponse, error) {
	args := protocol.P2PUBContractGetForSA{
		GisServiceCode: gis,
	}
	var res = protocol.P2PUBContractGetForSAResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func getCustomOSImageList(api *p2pubapi.API, gis, iar string) (*protocol.CustomOSImageListGetResponse, error) {
	args := protocol.CustomOSImageListGet{
		GisServiceCode: gis,
		IarServiceCode: iar,
	}
	var res = protocol.CustomOSImageListGetResponse{}
	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func dataSourceCustomOSImageRead(d *schema.ResourceData, m interface{}) error {

	if d.Get("filter") == nil && d.Get("image_id") == "" {
		return errors.New("filter or image_id is required")
	}

	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	contract, err := getContract(api, gis)
	if err != nil {
		return err
	}

	if contract.StorageArchive.ServiceCode == "" {
		return errors.New("cannot find storage archive contract")
	}

	iar := contract.StorageArchive.ServiceCode

	images, err := getCustomOSImageList(api, gis, iar)
	if err != nil {
		return err
	}

	var matches []int
	for idx, image := range images.ImageList {
		if d.Get("image_id") == image.ImageId {
			matches = []int{ idx }
			break
		}
		match := true
		for _, f := range d.Get("filter").([]interface{}) {
			filter := f.(map[string]interface{})
			switch filter["name"] {
			case "os_type":
				match = match && filter["value"] == image.OSType
			case "label":
				matched , _ := regexp.MatchString(filter["value"].(string), image.Label)
				match = match && matched
			case "image_id":
				match = match && filter["value"] == image.ImageId
			case "type":
				match = match && filter["value"] == image.Type
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
		return errors.New("no images matched")
	}

	if len(matches) >= 2 {
		if !d.Get("most_recent").(bool) {
			return errors.New("two or more images matched. please narrow down")
		}
		// TODO:
		picked := 0
		last_modified := "ZZZ"
		for _, idx := range matches {
			if images.ImageList[idx].ArchivedDateTime < last_modified {
				picked = idx
				last_modified = images.ImageList[idx].ArchivedDateTime
			}
		}
		matches = []int{ picked }
	}

	ans := images.ImageList[matches[0]]
	d.SetId(ans.ImageId)
	d.Set("os_type", ans.OSType)
	d.Set("created_at", ans.ArchivedDateTime)
	d.Set("label", ans.Label)
	d.Set("source", ans.SrcServiceCode)
	d.Set("image_id", ans.ImageId)
	d.Set("image_size", ans.ImageSize)
	d.Set("type", ans.Type)
	
	return nil
}
