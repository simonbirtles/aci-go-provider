package aci

type ApicPostInfo struct {
	Path   		string
	Filter 		ApicQueryFilter
	Payload 	[]byte
	ApicClient 	ApicClientInfo
	Delay		int
}

type ApicGetInfo struct {
	Path   		string
	Filter 		ApicQueryFilter
	ApicClient 	ApicClientInfo
	Delay		int
}

type ApicDeleteInfo struct {
	Path 			string
	ApicClient 	ApicClientInfo
	Delay		int
}

type ApicClientInfo struct {
	ApicHosts 	[]string
	Cookie		string
}

type ApicQueryFilter struct {
	Query_target         		string `json:"query-target"`
	Target_subtree_class 		string `json:"target-subtree-class"`
	Query_target_filter  		string `json:"query-target-filter"`
	Rsp_subtree          		string `json:"rsp-subtree"`
	Rsp_subtree_class    		string `json:"rsp-subtree-class"`
	Rsp_subtree_filter   		string `json:"rsp-subtree-filter"`
	Rsp_subtree_include  		string `json:"rsp-subtree-include"`
	Rsp_prop_include	  		string `json:"rsp-prop-include"`
	Order_by             		string `json:"order-by"`
}
