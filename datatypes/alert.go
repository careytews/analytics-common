package datatypes

type Alert struct {
	Id          string `json:"id,omitempty"`
	Time        string `json:"time,omitempty"`
	Host        string `json:"host,omitempty"`
	Action      string `json:"action,omitempty"`
	Device      string `json:"device,omitempty"`
	Description string `json:"description,omitempty"`
}
