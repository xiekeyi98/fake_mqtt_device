package device

type DeviceInterface interface {
	Connect() error
	Disconnect()
}
