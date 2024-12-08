package config

type RootConf struct {
	RabbitMQConf RabbitMQConf `mapstructure:"rabbitmqConf"`
}

type RabbitMQConf struct {
	Conection Connection `mapstructure:"connection"`
	Queue     Queue      `mapstructure:"queue"`
	Exchange  Exchange   `mapstructure:"exchange"`
	QoS       QoS        `mapstructure:"qos"`
}

type Connection struct {
	URL      string `mapstructure:"url"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
}

type Queue struct {
	Name        string `mapstructure:"name"`
	Durable     bool   `mapstructure:"durable"`
	AutoDelete  bool   `mapstructure:"autoDelete"`
	ConsumerTag string `mapstructure:"consumerTag"`
}

type Exchange struct {
	Name       string `mapstructure:"name"`
	Kind       string `mapstructure:"kind"`
	Durable    bool   `mapstructure:"durable"`
	AutoDelete bool   `mapstructure:"autoDelete"`
}

type QoS struct {
	Count  int  `mapstructure:"count"`
	Size   int  `mapstructure:"size"`
	Global bool `mapstructure:"global"`
}
