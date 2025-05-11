package messager

// ProducerConfig определяет конфигурацию продюсера
type ProducerConfig struct {
	BootstrapServers string
	Retries          int
}

func NewProducerConfig(bootstrapServers string, retries int) *ProducerConfig {
	return &ProducerConfig{
		BootstrapServers: bootstrapServers,
		Retries:          retries,
	}
}
