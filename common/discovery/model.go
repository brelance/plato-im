package discovery

import "encoding/json"

// this model define the value stored in ETCD
type EndpointInfo struct {
	IP       string                 `json:"ip"`
	Port     string                 `json:"port"`
	MetaData map[string]interface{} `json:"meta"`
}

func (edi *EndpointInfo) Marshal() string {
	data, err := json.Marshal(edi)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func UnMarshal(data []byte) (*EndpointInfo, error) {
	edi := &EndpointInfo{}
	err := json.Unmarshal(data, edi)
	if err != nil {
		return nil, err
	}
	return edi, nil
}
