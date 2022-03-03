package types

import "encoding/json"

type BeeRequest struct {
	MetaData *MetaData `json:"metadata"`
}

type BeeRequestWorker struct {
	MetaData *MetaData             `json:"metadata"`
	PayLoad  *WorkerRequestPayLoad `json:"payload"`
}

type BeeRequestBatch struct {
	MetaData *MetaData            `json:"metadata"`
	PayLoad  *BatchRequestPayLoad `json:"payload"`
}

type BeeRequestCluster struct {
	MetaData *MetaData              `json:"metadata"`
	PayLoad  *ClusterRequestPayLoad `json:"payload"`
}

type MetaData struct {
	Type          string `json:"type"`
	SubType       string `json:"subType,omitempty"`
	From          string `json:"from"`
	To            string `json:"to"`
	Queue         string `json:"queue"`
	CorrelationId string `json:"correlationId"`
}

type WorkerRequestPayLoad struct {
	// RequestName string `json:"requestName,omitempty"`
	// Data        string `json:"data"`
	Data *WorkerRequestData `json:"data,omitempty"`
	// BatchData   *[]RequestData `json:"batchData,omitempty"`
}

type BatchRequestPayLoad struct {
	// RequestName string `json:"requestName,omitempty"`
	// Data        string `json:"data"`
	// Data        *WorkerRequestData   `json:"data,omitempty"`
	Data *[]WorkerRequestData `json:"batchData,omitempty"`
}

type ClusterRequestPayLoad struct {
	// RequestName string `json:"requestName,omitempty"`
	// Data        string `json:"data"`
	Data *ClusterRequestData `json:"data,omitempty"`
	// BatchData   *[]RequestData `json:"batchData,omitempty"`
}
type WorkerRequestData struct {
	WorkerId  string `json:"workerId,omitempty"`
	ClusterId string `json:"clusterId,omitempty"`
	// ClusterIps *[]string `json:"clusterIps,omitempty"`
	NodeIp    string `json:"nodeIp,omitempty"`
	NodePort  int    `json:"nodeSSHPort,omitempty"`
	ProxyPort int    `json:"proxyPort,omitempty"`
}

type ClusterRequestData struct {
	ClusterId                 string    `json:"clusterId,omitempty"`
	RouteNodeIps              *RouteIps `json:"routeNodeIps,omitempty"`
	IngresControllerNodeports []int     `json:"ingresControllerNodeports,omitempty"`
}

type RouteIps struct {
	SSH  *[]string `json:"ssh,omitempty"`
	Http *[]string `json:"http,omitempty"`
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
