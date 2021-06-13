package kvserver

import "fmt"

type Creative struct {
	ID   int    `json:"id,omitempty"`
	ADM  string `json:"adm,omitempty"`
	Type string `json:"type,omitempty"`
}

func NewCreative(id int, adm string, crtype string) (*Creative, error) {
	obj := &Creative{
		ID:   id,
		ADM:  adm,
		Type: crtype,
	}

	if id <= 0 {
		return nil, fmt.Errorf("invalid creative id")
	}

	if (crtype == "banner" || crtype == "video" || crtype == "native") == false {
		return nil, fmt.Errorf("invalid creative type:%v", crtype)
	}

	return obj, nil
}
