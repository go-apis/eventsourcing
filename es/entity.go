package es

type Entity struct {
	ServiceName   string      `json:"service_name"`
	Namespace     string      `json:"namespace"`
	AggregateId   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Data          interface{} `json:"data"`
}
