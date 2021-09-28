package shadow

type ShadowUploadReq struct {
	Type        string      `json:"type"`
	State       shadowState `json:"state,omitempty"`
	Version     int         `json:"version"`
	ClientToken string      `json:"clientToken"`
}

//https://cloud.tencent.com/document/product/634/11918
type ShadowInfo struct {
	State     shadowState   `json:"state"`
	Metadata  metadataState `json:"metadata"`
	Version   int           `json:"version"`
	Timestamp int           `json:"timestamp"`
}

func (s *ShadowInfo) GetVersion() int {
	return s.Version
}

type shadowState struct {
	Reported map[string]interface{} `json:"reported,omitempty"`
	Desired  map[string]interface{} `json:"desired,omitempty"`
}

type metadataState struct {
	Reported map[string]interface{} `json:"reported,omitempty"`
	Desited  map[string]interface{} `json:"desired,omitempty"`
}
