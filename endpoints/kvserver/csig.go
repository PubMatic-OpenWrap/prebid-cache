package kvserver

import "strings"

type Result struct {
	LineItems []*LineItem       `json:"lineitems,omitempty"`
	Creatives map[int]*Creative `json:"creatives,omitempty"`
}

func GetResult(key string) *Result {
	result := &Result{
		LineItems: []*LineItem{},
		Creatives: make(map[int]*Creative),
	}
	subKeys := strings.Split(key, ":")
	for _, li := range lineItemMap {
		found := true
		if len(key) > 0 {
			if len(li.RegExpression) != len(subKeys) {
				continue
			}
			for index, subKey := range subKeys {
				found = found && li.RegExpression[index].Match([]byte(subKey))
				if found == false {
					break
				}
			}
		}
		if found {
			//append creatives
			for _, cr := range li.Creatives {
				if crObj, ok := creativeMap[cr]; ok {
					found = true
					result.Creatives[crObj.ID] = crObj
				}
			}

			if found {
				li.caluclatePacingRate()
				result.LineItems = append(result.LineItems, li)
			}
		}
	}
	return result
}
