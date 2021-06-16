package kvserver

var lineItemMap map[int]*LineItem
var creativeMap map[int]*Creative

//var csigLineItemMap map[string][]int

func Initialise() {
	lineItemMap = make(map[int]*LineItem)
	creativeMap = make(map[int]*Creative)
	//csigLineItemMap = make(map[string][]int)
}

//Add Functions
func AddNewLineItem(lineItem *LineItem) {
	lineItemMap[lineItem.ID] = lineItem
}

func AddNewCreative(creative *Creative) {
	creativeMap[creative.ID] = creative
}

func AddNewLineItemCreativeMapping(lineItemID, creativeID int) {
	if li, ok := lineItemMap[lineItemID]; ok {
		if _, ok := creativeMap[creativeID]; ok {
			li.Creatives = append(li.Creatives, creativeID)
		}
	}
}

func UnmapLineItemCreativeMapping(lineItemID, creativeID int) {
	if li, ok := lineItemMap[lineItemID]; ok {
		for index, id := range li.Creatives {
			if id == creativeID {
				li.Creatives = append(li.Creatives[:index], li.Creatives[index+1:]...)
				break
			}
		}
	}
}

/*
func AddCSIGLIMapping(key string, lineitemID int) {
	values := csigLineItemMap[key]
	values = append(values, lineitemID)
	csigLineItemMap[key] = values
}

func UnmapCSIGLIMapping(key string, lineItemID int) {
	if values, ok := csigLineItemMap[key]; ok {
		for index, id := range values {
			if id == lineItemID {
				values = append(values[:index], values[index+1:]...)
				break
			}
		}

		if len(values) == 0 {
			delete(csigLineItemMap, key)
		}
	}
}
*/

func FlushAll() {
	Initialise()
}

func init() {
	Initialise()
}
