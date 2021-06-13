package kvserver

import "fmt"

var lineItemMap map[int]*LineItem
var creativeMap map[int]*Creative
var lineitemCreativeMap map[int][]int
var csigLineItemMap map[string][]int

func Initialise() {
	lineItemMap = make(map[int]*LineItem)
	creativeMap = make(map[int]*Creative)
	lineitemCreativeMap = make(map[int][]int)
	csigLineItemMap = make(map[string][]int)
}

//Add Functions
func AddNewLineItem(lineItem *LineItem) {
	lineItemMap[lineItem.ID] = lineItem
}

func AddNewCreative(creative *Creative) {
	creativeMap[creative.ID] = creative
}

func AddNewLineItemCreativeMapping(lineItemID, creativeID int) {
	values := lineitemCreativeMap[lineItemID]
	values = append(values, creativeID)
	lineitemCreativeMap[lineItemID] = values
}

func AddCSIGLIMapping(csigkey string, lineitemID int) {
	values := csigLineItemMap[csigkey]
	values = append(values, lineitemID)
	csigLineItemMap[csigkey] = values
}

func UnmapLineItemCreativeMapping(lineItemID, creativeID int) {
	if creatives, ok := lineitemCreativeMap[lineItemID]; ok {
		for index, id := range creatives {
			if id == creativeID {
				creatives = append(creatives[:index], creatives[index+1:]...)
				break
			}
		}

		if len(creatives) == 0 {
			delete(lineitemCreativeMap, lineItemID)
		}
	}
}

func UnmapCSIGLIMapping(csigkey string, lineItemID int) {
	if values, ok := csigLineItemMap[csigkey]; ok {
		for index, id := range values {
			if id == lineItemID {
				values = append(values[:index], values[index+1:]...)
				break
			}
		}

		if len(values) == 0 {
			delete(csigLineItemMap, csigkey)
		}
	}
}

func CSIGKey(csKey, csValue, ig string) string {
	return fmt.Sprintf("%s:%s:%s", csKey, csValue, ig)
}

func FlushAll() {
	Initialise()
}

func init() {
	Initialise()
}
