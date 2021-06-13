package kvserver

type KeyValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
type CSIGMap struct {
	CS        KeyValue `json:"cs,omitempty"`
	IG        string   `json:"ig,omitempty"`
	LineItems []int    `json:"lineitems,omitempty"`
}
type CSIGResult struct {
	Mappings  []*CSIGMap        `json:"mappings,omitempty"`
	LineItems map[int]*LineItem `json:"lineitems,omitempty"`
	Creatives map[int]*Creative `json:"creatives,omitempty"`
}

func AppendResult(csKey, csValue, ig string, result *CSIGResult) {
	csigmap := &CSIGMap{
		CS: KeyValue{Key: csKey, Value: csValue},
		IG: ig,
	}
	result.Mappings = append(result.Mappings, csigmap)

	key := CSIGKey(csKey, csValue, ig)
	if lineitems, ok := csigLineItemMap[key]; ok {
		for _, li := range lineitems {
			if liObj, ok := lineItemMap[li]; ok {
				found := false
				//append creatives
				if creatives, ok := lineitemCreativeMap[li]; ok {
					for _, cr := range creatives {
						if crObj, ok := creativeMap[cr]; ok {
							found = true
							result.Creatives[crObj.ID] = crObj
						}
					}
				}

				if found {
					result.LineItems[liObj.ID] = liObj
					csigmap.LineItems = append(csigmap.LineItems, li)
				}
			}
		}
	}
}
