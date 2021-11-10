package aci

import (
	"os"
	"fmt"
	"testing"
)

var cookie string
var err error

func TestACILogin(t *testing.T) {

	// APIC Authentication
	fmt.Println("\n\nAuthentication Test")
	fmt.Println("=============================================================================================================================================")
	apic := os.Getenv("ACI_APIC");
	user := os.Getenv("ACI_APIC_USERNAME");
	pass := os.Getenv("ACI_APIC_PASSWORD");
	cookie, err := Aci_login(apic, user, pass)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		if len(cookie) > 0 {
			fmt.Println("\nAPIC-cookie: ", cookie)
		} else {
			fmt.Println("Invalid Cookie")
			t.Fail()
		}
	}
}

func TestACIGetTenant(t *testing.T) {

	// simple get of a tenant with query-target-filter
	fmt.Println("\n\nGET Tenant Test")
	fmt.Println("=============================================================================================================================================")	
	var info = new(ApicGetInfo)
	info.Path = "class/fvTenant"
	info.Cookie = cookie
	info.Filter.Query_target_filter = `wcard(fvTenant.name, "TEN_.*")`
	data, err := Get(info)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(data))
	}
}

func TestACICreateTenant(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create Tenant Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{  
			"fvTenant" : { 
				"attributes" : {  
					"name" : "TEN_TF_TEST", 
					"descr": "Terraform Managed"    
				}, 
				"children": [ 
					{ 
						"tagInst": {  
							"attributes": { 
								"name": "terraform"  
							} 
						} 
					} 
				]  
			}  
		}`)

	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateVrf(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create VRF Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
			"fvCtx": {
				"attributes": {
					"name": "VRF_TF_TEST",
					"descr": "Terraform Managed",
					"bdEnforcedEnable": "no",
                    "knwMcastAct": "permit",
                    "pcEnfDir": "ingress",
                    "pcEnfPref": "enforced"
				},
				"children": [
					{
						"tagInst": {
							"attributes": {
								"name": "terraform"
							}
						}
					}
				]
			}
		}`)

	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateBridgeDomain(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create Bridge Domain Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
			"fvBD": {
				"attributes": {
					"name": "BD_TF_TEST_01",
					"descr": "Terraform Managed"
					"OptimizeWanBandwidth": "no",
                    "arpFlood": "no",
                    "epMoveDetectMode": "",
                    "intersiteBumTrafficAllow": "no",
                    "intersiteL2Stretch": "no",
                    "ipLearning": "yes",
                    "limitIpLearnToSubnets": "yes",
                    "mcastAllow": "no",
                    "multiDstPktAct": "bd-flood",
                    "unicastRoute": "yes",
                    "unkMacUcastAct": "proxy",
                    "unkMcastAct": "flood"
				},
				"children": [
					{
						"fvRsCtx": {
							"attributes": {
								"tnFvCtxName": "VRF_TF_TEST"
							}
						}
					},
					{
						"tagInst": {
							"attributes": {
								"name": "terraform"
							}
						}
					}
				]
			}
		}`)

	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateSubnet(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create Subnet in Bridge Domain Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST/BD-BD_TF_TEST_01.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
			"fvSubnet": {
				"attributes": {
                    "ip": "192.168.45.254/24",
					"descr": "Terraform Managed",
					"scope": "public,shared",
					"ctrl": "nd",
					"preferred": "no",
                    "virtual": "no"
				},
				"children": [
					{
						"tagInst": {
							"attributes": {
								"name": "terraform"
							}
						}
					}
				]
			}
		}`)
		
	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateAppProfile(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create App Profile Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
			"fvAp": {
				"attributes": {
					"name": "APP_TF_01",
					"descr": "Terraform Managed"
				},
				"children": [
					{
						"tagInst": {
							"attributes": {
								"name": "terraform"
							}
						}
					}
				]
			}
		}`)
	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateContractFilters(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create Contract Filters Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
            "vzFilter": {
                "attributes": {
                    "descr": "",
                    "name": "TCP_6574"
                },
                "children": [
                    {
                        "vzEntry": {
                            "attributes": {
                                "applyToFrag": "no",
                                "arpOpc": "unspecified",
                                "dFromPort": "6574",
                                "dToPort": "6574",
                                "descr": "",
                                "etherT": "ip",
                                "icmpv4T": "unspecified",
                                "icmpv6T": "unspecified",
                                "matchDscp": "unspecified",
                                "name": "6574",
                                "prot": "tcp",
                                "sFromPort": "unspecified",
                                "sToPort": "unspecified",
                                "stateful": "no",
                                "tcpRules": ""
                            }
                        }
					},
					{ 	"tagInst": {  
							"attributes": { 
									"name": "terraform"  
								} 
							} 
					}
                ]
            }
        }`)


	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

func TestACICreateContracts(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create Contracts Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{
            "vzBrCP": {
                "attributes": {
                    "descr": "Terraform Managed",
                    "name": "CNT_TF_HTTP_PROXY",
                    "prio": "unspecified",
                    "scope": "context",
                    "targetDscp": "unspecified"
                },
                "children": [
                    {
                        "vzSubj": {
                            "attributes": {
                                "consMatchT": "AtleastOne",
                                "descr": "Terraform Managed",
                                "name": "SUBJ_PROXY",
                                "prio": "unspecified",
                                "provMatchT": "AtleastOne",
                                "revFltPorts": "yes",
                                "targetDscp": "unspecified"
                            },
                            "children": [
                                {
                                    "vzRsSubjFiltAtt": {
                                        "attributes": {
                                            "directives": "",
                                            "tnVzFilterName": "TCP_6574"
                                        }
                                    }
                                }
                            ]
                        }
					},
					{ 	"tagInst": {  
							"attributes": { 
								"name": "terraform"  
							} 
						} 
					}
                ]
            }
        }`)


	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}

/* TODO:
* Add EPG Subnet ?
* Add Contracts - Subject Label ?
*/
func TestACICreateEpg(t *testing.T) {

	// simple test of POST 
	fmt.Println("\n\nPOST Create EPG with Contracts Test")
	fmt.Println("=============================================================================================================================================")
	var postinfo = new(ApicPostInfo)
	postinfo.Path = "mo/uni/tn-TEN_TF_TEST/ap-APP_TF_01.json"
	postinfo.Cookie = cookie
	postinfo.Filter.Rsp_subtree = "modified"
	postinfo.Payload = []byte(`
		{  "fvAEPg" : { 
				"attributes" : {  
					"name" : "EPG_TF_TEST_01", 
					"descr": "Terraform Managed"    
				}, 
				"children": [ 
						{
						"fvRsBd" : { 
							"attributes": { 
								"tnFvBDName": "BD_TF_TEST_01" 
								} 
							} 
						},
						{
						"fvRsDomAtt": {
							"attributes": {
								"classPref": "encap",
								"encap": "unknown",
								"encapMode": "auto",
								"epgCos": "Cos0",
								"epgCosPref": "disabled",
								"instrImedcy": "immediate",
								"resImedcy": "immediate",
								"netflowDir": "both",
								"netflowPref": "disabled",
								"primaryEncap": "unknown",
								"primaryEncapInner": "unknown",
								"secondaryEncapInner": "unknown",
								"switchingMode": "native",
								"tDn": "uni/vmmp-VMware/dom-VMM_VMW_DVS_01"
							},
							"children": [
								{
									"vmmSecP": {
										"attributes": {
											"allowPromiscuous": "reject",
											"forgedTransmits": "reject",
											"macChanges": "reject"
											}
										}
									}
								]
							}
						}, 
						{
						"fvRsProv": {
							"attributes": {
								"matchT": "AtleastOne",
								"tnVzBrCPName": "CNT_TF_HTTP_PROXY"
								}
							}
						},
						{
						"fvRsCons": {
							"attributes": {
								"tnVzBrCPName": "CNT_TF_HTTP_PROXY"
								}
							}
						},
						{ 
						"tagInst": {  
							"attributes": { 
							"name": "terraform"  
								} 
							} 
						} 								
					]  
				}  
			}`)


	payload, err := Post(postinfo)
	if err != nil {
		fmt.Println("\nError: ", err)
		t.Fail()
	} else {
		fmt.Println(string(payload))
	}
}
