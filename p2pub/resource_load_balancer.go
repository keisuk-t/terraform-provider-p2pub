package p2pub

import (
	"fmt"
	"time"

	"github.com/iij/p2pubapi"
	"github.com/iij/p2pubapi/protocol"

	"github.com/hashicorp/terraform/helper/schema"
)

const (
	pollInterval = time.Duration(10 * time.Second)
)

func resourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerCreate,
		Read:   resourceLoadBalancerRead,
		Update: resourceLoadBalancerUpdate,
		Delete: resourceLoadBalancerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(10 * time.Minute),
			Update:  schema.DefaultTimeout(10 * time.Minute),
			Delete:  schema.DefaultTimeout(10 * time.Minute),
			Default: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// D10M, D100M, D150M, D1000M
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"redundant": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			// Global, PrivateStandard, Private
			"external_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"external_servicecode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			// PrivateStandard, Private
			"internal_type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"internal_trafficip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"internal_servicecode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"administration_server_allow_network_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"trafficip_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv4_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"ipv4_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ipv4_domainname": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ipv6_name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ipv6_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"ipv6_domainname": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Required: true,
			},
			"host_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"master": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"external_ipv4_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"external_ipv6_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"internal_ipv4_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"filter_in_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						// IPAddr/mask or ANY
						"source_network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// IPAddr/mask or ANY
						"destination_network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// number or ANY
						"destination_port": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// TCP or UDP
						"protocol": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// ACCEPT or DROP or REJECT
						"action": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"label": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Optional: true,
			},
			"filter_out_list": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						// IPAddr/mask or ANY
						"source_network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// IPAddr/mask or ANY
						"destination_network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// number or ANY
						"destination_port": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// TCP or UDP
						"protocol": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						// ACCEPT or DROP or REJECT
						"action": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"label": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Optional: true,
			},
		},
	}
}

/*
  Utility
*/

func waitLoadBalancerContract(api *p2pubapi.API, gis, ifl string, cstatus p2pubapi.Status, maxwait time.Duration) error {
	start := time.Now()
	for {
		arg := protocol.FwLbContractGet{
			GisServiceCode: gis,
			IflServiceCode: ifl,
			Item:           "ContractStatus",
		}
		var res = protocol.FwLbContractGetResponse{}
		if err := p2pubapi.Call(*api, arg, &res); err != nil {
			return err
		}
		if cstatus == p2pubapi.None || cstatus.String() == res.ContractStatus {
			break
		}
		if time.Since(start) > maxwait {
			return fmt.Errorf("timeout")
		}
		time.Sleep(pollInterval)
	}

	return nil
}

// WaitLoadBalancer wait LoadBalancer status (contract status, resource status)
// Contract Status(cstatus): InPreparation/InService
// Resource Status(rstatus): Initialized/Starting/Running/Configuring/Configured/Locked/Updating
func waitLoadBalancer(api *p2pubapi.API, gis, ifl string, cstatus, rstatus p2pubapi.Status, maxwait time.Duration) error {
	if err := waitLoadBalancerContract(api, gis, ifl, cstatus, maxwait); err != nil {
		return err
	}

	start := time.Now()
	for {
		arg := protocol.FwLbGet{
			GisServiceCode: gis,
			IflServiceCode: ifl,
		}
		var res = protocol.FwLbGetResponse{}
		if err := p2pubapi.Call(*api, arg, &res); err != nil {
			return err
		}
		if (cstatus == p2pubapi.None || cstatus.String() == res.ContractStatus) &&
			(rstatus == p2pubapi.None || rstatus.String() == res.ResourceStatus) {
			break
		}
		if time.Since(start) > maxwait {
			return fmt.Errorf("timeout")
		}
		time.Sleep(pollInterval)
	}

	return nil
}

func setLoadBalancerLabel(api *p2pubapi.API, gis, ifl, name string) error {
	args := protocol.FwLbLabelSet{
		GisServiceCode: gis,
		IflServiceCode: ifl,
		Name:           name,
	}
	res := protocol.FwLbLabelSetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	return nil
}

func getFilter(api *p2pubapi.API, gisServiceCode, iflServiceCode, direction string) *[]map[string]string {
	args := protocol.FwLbFilterGet{
		GisServiceCode: gisServiceCode,
		IflServiceCode: iflServiceCode,
		IpVersion:      "v4",
		Direction:      direction,
	}
	res := protocol.FwLbFilterGetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return nil
	}

	filters := make([]map[string]string, 0)
	for _, rule := range res.FilterRuleList {
		filters = append(filters, map[string]string{
			"filter_id":           rule.FilterId,
			"source_network":      rule.SourceNetwork,
			"destination_network": rule.DestinationNetwork,
			"destination_port":    rule.DestinationPort,
			"protocol":            rule.Protocol,
			"action":              rule.Action,
			"label":               rule.Label,
		})
	}

	return &filters
}

func buildFilterList(d *schema.ResourceData, s string) []protocol.FilterRule {
	result := []protocol.FilterRule{}
	filters := d.Get(s).([]interface{})

	for _, filter := range filters {
		f := filter.(map[string]interface{})
		result = append(result, protocol.FilterRule{
			SourceNetwork:      f["source_network"].(string),
			DestinationNetwork: f["destination_network"].(string),
			DestinationPort:    f["destination_port"].(string),
			Protocol:           f["protocol"].(string),
			Action:             f["action"].(string),
			Label:              f["label"].(string),
		})
	}

	return result
}

func updateFilter(d *schema.ResourceData, m interface{}, direction string) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	filterRuleList := buildFilterList(d, "filter_"+direction+"_list")
	args := protocol.FwLbFilterSet{
		GisServiceCode: gis,
		IflServiceCode: d.Id(),
		IpVersion:      "v4",
		Direction:      direction,
		FilterRuleList: filterRuleList,
	}
	res := protocol.FwLbFilterSetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	return nil
}

func updateAdminAcl(d *schema.ResourceData, m interface{}) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode
	list := d.Get("administration_server_allow_network_list").([]interface{})

	acl := []string{}
	for _, a := range list {
		acl = append(acl, a.(string))
	}

	args := protocol.LBControlPanelACLSet{
		GisServiceCode:                       gis,
		IflServiceCode:                       d.Id(),
		AdministrationServerAllowNetworkList: acl,
	}
	res := protocol.LBControlPanelACLSetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	return nil
}

func createLoadBalancer(api *p2pubapi.API, gisServiceCode, lbType, redundant string) (string, error) {
	args := protocol.FwLbAdd{
		GisServiceCode: gisServiceCode,
		Type:           lbType,
		Redundant:      redundant,
	}
	res := protocol.FwLbAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return "", err
	}

	return res.ServiceCode, nil
}

func setupLoadBalancerSimple(api *p2pubapi.API, gisServiceCode, iflServiceCode, externalType, internalType, trafficIpName string) error {
	argsSetup := protocol.FwLbSetup{
		GisServiceCode: gisServiceCode,
		IflServiceCode: iflServiceCode,
		ActionType:     "Setup",
	}

	argsSetup.External.NetworkType = externalType
	argsSetup.Internal.NetworkType = internalType
	argsSetup.External.TrafficIpName = trafficIpName
	resSetup := protocol.FwLbSetupResponse{}

	if err := p2pubapi.Call(*api, argsSetup, &resSetup); err != nil {
		return err
	}

	return nil
}

/*
  API
*/

func resourceLoadBalancerRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.FwLbGet{
		GisServiceCode: gis,
		IflServiceCode: d.Id(),
	}
	res := protocol.FwLbGetResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.Set("type", res.Type)
	d.Set("redundant", res.Redundant)
	d.Set("label", res.Label)

	d.Set("internal_type", res.Internal.NetworkType)
	d.Set("internal_trafficip_address", res.Internal.TrafficIpAddress)
	d.Set("internal_servicecode", res.Internal.ServiceCode)

	d.Set("external_type", res.External.NetworkType)
	d.Set("external_servicecode", res.External.ServiceCode)

	adminacl := []string{}
	for _, acl := range res.Lb.AdministrationServerAllowNetworkList {
		adminacl = append(adminacl, acl)
	}
	d.Set("administration_server_allow_network_list", adminacl)

	trafficIPList := make([]map[string]string, 0)
	for _, trafficip := range res.Lb.TrafficIpList {
		trafficIPList = append(trafficIPList, map[string]string{
			"ipv4_name":       trafficip.IPv4.TrafficIpName,
			"ipv4_address":    trafficip.IPv4.TrafficIpAddress,
			"ipv4_domainname": trafficip.IPv4.DomainName,
			"ipv6_name":       trafficip.IPv6.TrafficIpName,
			"ipv6_address":    trafficip.IPv6.TrafficIpAddress,
			"ipv6_domainname": trafficip.IPv6.DomainName,
		})
	}
	d.Set("trafficip_list", trafficIPList)

	hostList := make([]map[string]string, 0)
	for _, host := range res.HostList {
		hostList = append(hostList, map[string]string{
			"url":                   host.LbAdministrationServerUrl,
			"version":               host.LbSoftwareVersion,
			"master":                host.Master,
			"external_ipv4_address": host.External.IPv4Address,
			"external_ipv6_address": host.External.IPv6Address,
			"internal_ipv4_address": host.Internal.IPv4Address,
		})
	}
	d.Set("host_list", hostList)

	// Snatは省略
	// StaticRouteは省略

	d.Set("filter_in_list", getFilter(api, gis, d.Id(), "in"))
	d.Set("filter_out_list", getFilter(api, gis, d.Id(), "out"))

	return nil
}

func addTrafficIp(api *p2pubapi.API, gisServiceCode, iflServiceCode, name, address string) error {
	args := protocol.TrafficIpAdd{
		GisServiceCode:   gisServiceCode,
		IflServiceCode:   iflServiceCode,
		TrafficIpName:    name,
		TrafficIpAddress: address,
	}

	res := protocol.TrafficIpAddResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	return nil
}

/*
  FW+LBの契約、セットアップ、FWルーの設定まで一気に実行する
  external_typeはGlobalかPrivateStandard、internal_typeはPrivateStandardまで対応
  ToDo:
	external_type, internal_typeをPrivateに対応
	SNATに対応
	スタティックルーティングに対応
*/
func resourceLoadBalancerCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	var servicecode string
	var err error
	if servicecode, err = createLoadBalancer(api, gis, d.Get("type").(string), d.Get("redundant").(string)); err != nil {
		return err
	}

	if err := waitLoadBalancer(api, gis, servicecode, p2pubapi.InService, p2pubapi.Initialized, TIMEOUT); err != nil {
		return err
	}

	if d.Get("label") != nil && d.Get("label").(string) != "" {
		if err := setLoadBalancerLabel(api, gis, servicecode, d.Get("label").(string)); err != nil {
			return err
		}
	}

	first := true
	for _, t := range d.Get("trafficip_list").([]interface{}) {
		trafficip := t.(map[string]interface{})
		if first {
			if err := setupLoadBalancerSimple(api, gis, servicecode, d.Get("external_type").(string), d.Get("internal_type").(string), trafficip["ipv4_name"].(string)); err != nil {
				return err
			}
			first = false
		} else {
			if err := addTrafficIp(api, gis, servicecode, trafficip["ipv4_name"].(string), trafficip["ipv4_address"].(string)); err != nil {
				return err
			}
		}

		// セットアップが終わるのを待つ
		if err := waitLoadBalancer(api, gis, servicecode, p2pubapi.InService, p2pubapi.Configured, TIMEOUT); err != nil {
			return err
		}
	}

	d.SetId(servicecode)

	if d.Get("filter_out_list") != nil {
		if err := updateFilter(d, m, "out"); err != nil {
			return err
		}
	}

	if d.Get("filter_in_list") != nil {
		if err := updateFilter(d, m, "in"); err != nil {
			return err
		}
	}

	if d.Get("administration_server_allow_network_list") != nil {
		if err := updateAdminAcl(d, m); err != nil {
			return err
		}
	}

	/*
		argsSetup.External.ServiceCode = d.Get("external_servicecdoe").(string)			 // external_servicecode
		argsSetup.External.TrafficIpAddress = d.Get("external_ipv4_address").(string)	 // trafficip_list.ipv4_address
		argsSetup.External.Netmask = d.Get("external_ipv4_netmask").(string)             // ?
		argsSetup.External.MasterHostAddress = d.Get("external_master_address").(string) // host_list.external_ipv4_address when master = true
		argsSetup.External.SlaveHostAddress = d.Get("external_slave_address").(string)   // host_list.external_ipv4_address when master = false
		argsSetup.Internal.ServiceCode = d.Get("internal_servicecdoe").(string)          // internal_servicecode
		argsSetup.Internal.TrafficIpAddress = d.Get("internal_ipv4_address").(string)    // internal_trafficip_address
		argsSetup.Internal.Netmask = d.Get("internal_ipv4_netmask").(string)             // ?
		argsSetup.Internal.MasterHostAddress = d.Get("internal_master_address").(string) // host_list.interla_ipv4_address when master = true
		argsSetup.Internal.SlaveHostAddress = d.Get("internal_slave_address").(string)   // host_list.interla_ipv4_address when master = false
		argsSetup.Internal.TrafficIpName = d.Get("internal_trafficip_name").(string)     // ?
	*/

	return nil
}

func resourceLoadBalancerUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	d.Partial(true)

	if d.HasChange("type") {
		return fmt.Errorf("updating type is not supported")
	}

	if d.HasChange("redundant") {
		return fmt.Errorf("updating redundat is not supported")
	}

	if d.HasChange("label") {
		if err := setLoadBalancerLabel(api, gis, d.Id(), d.Get("label").(string)); err != nil {
			return err
		}
		d.SetPartial("label")
	}

	if d.HasChange("trafficip_list") {
		return fmt.Errorf("updating trafficip is not supported")
	}

	if d.HasChange("filter_out_list") {
		if err := updateFilter(d, m, "out"); err != nil {
			return err
		}
		d.SetPartial("filter_out_list")
	}

	if d.HasChange("filter_in_list") {
		if err := updateFilter(d, m, "in"); err != nil {
			return err
		}
		d.SetPartial("filter_in_list")
	}

	if d.HasChange("administration_server_allow_network_list") {
		if err := updateAdminAcl(d, m); err != nil {
			return err
		}
		d.SetPartial("administration_server_allow_network_list")
	}

	d.Partial(false)

	return nil
}

func resourceLoadBalancerDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*Context).API
	gis := m.(*Context).GisServiceCode

	args := protocol.FwLbCancel{
		GisServiceCode: gis,
		IflServiceCode: d.Id(),
	}
	res := protocol.FwLbCancelResponse{}

	if err := p2pubapi.Call(*api, args, &res); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
