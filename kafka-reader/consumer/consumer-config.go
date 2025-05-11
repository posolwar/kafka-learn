package consumer

// определяет режим работы консьюмера
type ConsumerMode string

const (
	SingleMode ConsumerMode = "single" // Обработка по одному сообщению
	BatchMode  ConsumerMode = "batch"  // Обработка пачками
)

// ConsumerConfig конфигурация консьюмера
type ConsumerConfig struct {
	GroupID          int
	BootstrapServers string
	Mode             ConsumerMode

	FetchMinBytes  int
	FetchMaxWaitMs int
	MaxPollRecords int
	AutoCommit     bool
}

func NewConsumerConfig(bootstrapServer string, groupID int, mode ConsumerMode) *ConsumerConfig {
	var ConsumerConfig ConsumerConfig

	if mode == SingleMode {
		ConsumerConfig.FetchMinBytes = 4
		ConsumerConfig.FetchMaxWaitMs = 1000
		ConsumerConfig.MaxPollRecords = 1
		ConsumerConfig.AutoCommit = true
	}

	if mode == BatchMode {
		ConsumerConfig.FetchMinBytes = 1024 * 10
		ConsumerConfig.FetchMaxWaitMs = 10000
		ConsumerConfig.MaxPollRecords = 10
		ConsumerConfig.AutoCommit = false
	}

	ConsumerConfig.GroupID = groupID
	ConsumerConfig.BootstrapServers = bootstrapServer
	ConsumerConfig.Mode = mode

	return &ConsumerConfig
}
