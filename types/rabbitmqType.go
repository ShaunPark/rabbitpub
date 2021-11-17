package types

import "encoding/json"

type BeeRequest struct {
	MetaData *MetaData       `json:"metadata"`
	PayLoad  *RequestPayLoad `json:"payload"`
}

type MetaData struct {
	Type          string `json:"type"`
	SubType       string `json:"subType,omitempty"`
	From          string `json:"from"`
	To            string `json:"to"`
	Queue         string `json:"queue"`
	CorrelationId string `json:"correlationId"`
}

type RequestPayLoad struct {
	RequestName string         `json:"requestName,omitempty"`
	Data        *RequestData   `json:"data,omitempty"`
	BatchData   *[]RequestData `json:"batchData,omitempty"`
}

type RequestData struct {
	WorkerId   string    `json:"workerId,omitempty"`
	ClusterIps *[]string `json:"clusterIps,omitempty"`
	NodeIp     string    `json:"nodeIp,omitempty"`
	NodePort   int       `json:"nodePort,omitempty"`
	ProxyPort  int       `json:"proxyPort,omitempty"`
}

type BeeResponse struct {
	MetaData *MetaData       `json:"metadata"`
	PayLoad  ResponsePayLoad `json:"payload"`
}

type ResponsePayLoad struct {
	Status           string             `json:"status"`
	StatusType       string             `json:"statusType,omitempty"`
	Data             string             `json:"data,omitempty"`
	BatchData        *ResponseBatchData `json:"batchData,omitempty"`
	Code             string             `json:"code,omitempty"`
	Message          string             `json:"message,omitempty"`
	ErrorDetail      string             `json:"errorDetail,omitempty"`
	ErrorOrigin      string             `json:"errorOrigin,omitempty"`
	ErrorUserMessage string             `json:"errorUserMessage,omitempty"`
	ErrorUserTitle   string             `json:"errorUserTitle,omitempty"`
}

type ResponseBatchData struct {
	Success []string `json:"success,omitempty"`
	Fail    []string `json:"fail,omitempty"`
}

func (r BeeResponse) String() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}
