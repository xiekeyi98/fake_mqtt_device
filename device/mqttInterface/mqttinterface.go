package mqttClient

type Interface interface {
	Publish(topic string, payload []byte)
}
