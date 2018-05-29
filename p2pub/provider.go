package p2pub

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iij/p2pubapi"
)

type Context struct {
	API *p2pubapi.API
	GisServiceCode string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key_id": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				Description: "",
				DefaultFunc: schema.EnvDefaultFunc("IIJAPI_ACCESS_KEY", ""),
			},
			"secret_access_key": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				Description: "",
				DefaultFunc: schema.EnvDefaultFunc("IIJAPI_SECRET_KEY", ""),
			},
			"gis_service_code": &schema.Schema{
				Type: schema.TypeString,
				Required: true,
				Description: "",
				DefaultFunc: schema.EnvDefaultFunc("GISSERVICECODE", ""),
			},
			"endpoint": &schema.Schema{
				Type: schema.TypeString,
				Optional: true,
				Description: "",
				Default: "p2pub.api.iij.jp",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"p2pub_virtual_server": resourceVirtualServer(),
			"p2pub_system_storage": resourceSystemStorage(),
			"p2pub_additional_storage": resourceAdditionalStorage(),
			"p2pub_storage_archive": resourceStorageArchive(),
			"p2pub_global_ip_address": resourceGlobalIPAddress(),
			"p2pub_private_network": resourcePrivateNetwork(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"p2pub_custom_os_image": dataSourceCustomOSImage(),
			"p2pub_virtual_server": dataSourceVirtualServer(),
			"p2pub_system_storage": dataSourceSystemStorage(),
			"p2pub_additional_storage": dataSourceAdditionalStorage(),
		},
		ConfigureFunc: func (d *schema.ResourceData) (interface{}, error) {
			api := p2pubapi.NewAPI(d.Get("access_key_id").(string), d.Get("secret_access_key").(string))
			context := &Context{
				API: api,
				GisServiceCode: d.Get("gis_service_code").(string),
			}
			return context, nil
		},
	}
}
