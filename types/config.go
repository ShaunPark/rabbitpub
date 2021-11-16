package types

type Config struct {
	TestMode bool     `yaml:"testMode"`
	RabbitMQ RabbitMQ `yaml:"rabbitmq"`
	HaProxy  HaProxy  `yaml:"haproxy"`
	Batch    Batch    `yaml:"batchProcess"`
}

type Batch struct {
	EventCount  int    `yaml:"eventCount"`
	MaxDuration int    `yaml:"maxDuration"`
	Durable     bool   `yaml:"durable"`
	WalDir      string `yaml:"walDirectory"`
}
type Host struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Id       string `yaml:"id"`
	Password string
}

type RabbitMQ struct {
	Server                   Host   `yaml:"server"`
	Queue                    string `yaml:"queueName"`
	UseReplyQueueFromMessage bool   `yaml:"useReplyQueueFromMessage"`
	ReplyQueue               string `yaml:"replyQueueName"`
}
type HaProxy struct {
	Master           Host `yaml:"master"`
	Second           Host `yaml:"second"`
	ReloadSecondNode bool `yaml:"reloadSecondNode"`
}
