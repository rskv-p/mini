package main

import (
	"fmt"
	"log"

	"github.com/rskv-p/mini/srv"
	"gorm.io/gorm"
)

func main() {
	// Create mock client instances
	dbClient := &MockDBClient{}
	logClient := &MockLogClient{}
	busClient := &MockBusClient{}
	cfgClient := &MockConfigClient{}

	// Create the service
	service := srv.New("TestService", dbClient, logClient, busClient, cfgClient)

	// Start the service
	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Stop the service (this would normally happen when shutting down)
	if err := service.Stop(); err != nil {
		log.Fatalf("Failed to stop service: %v", err)
	}
}

// Mock DBClient
type MockDBClient struct{}

func (m *MockDBClient) GetDB() (*gorm.DB, error) {
	fmt.Println("Mock DB Client: GetDB called")
	return nil, nil
}

func (m *MockDBClient) Migrate(models ...interface{}) error {
	fmt.Println("Mock DB Client: Migrate called")
	return nil
}

func (m *MockDBClient) First(dest interface{}, conds ...interface{}) *gorm.DB {
	fmt.Println("Mock DB Client: First called")
	return nil
}

func (m *MockDBClient) Create(model interface{}) error {
	fmt.Println("Mock DB Client: Create called")
	return nil
}

func (m *MockDBClient) Find(model interface{}, conditions ...interface{}) error {
	fmt.Println("Mock DB Client: Find called")
	return nil
}

// Mock LogClient
type MockLogClient struct{}

func (m *MockLogClient) Trace(message string, context map[string]interface{}) {
	fmt.Println("Mock Log Client: Trace -", message)
}

func (m *MockLogClient) Debug(message string, context map[string]interface{}) {
	fmt.Println("Mock Log Client: Debug -", message)
}

func (m *MockLogClient) Info(message string, context map[string]interface{}) {
	fmt.Println("Mock Log Client: Info -", message)
}

func (m *MockLogClient) Warn(message string, context map[string]interface{}) {
	fmt.Println("Mock Log Client: Warn -", message)
}

func (m *MockLogClient) Error(message string, context map[string]interface{}) {
	fmt.Println("Mock Log Client: Error -", message)
}

// Mock BusClient
type MockBusClient struct{}

func (m *MockBusClient) Subscribe(subject string) error {
	fmt.Println("Mock Bus Client: Subscribe called for subject:", subject)
	return nil
}

func (m *MockBusClient) Unsubscribe(subject string) {
	fmt.Println("Mock Bus Client: Unsubscribe called for subject:", subject)
}

func (m *MockBusClient) Deliver(subject string, data []byte) {
	fmt.Println("Mock Bus Client: Deliver called for subject:", subject, "data:", string(data))
}

func (m *MockBusClient) Publish(subject string, data []byte) error {
	fmt.Println("Mock Bus Client: Publish called for subject:", subject, "data:", string(data))
	return nil
}

func (m *MockBusClient) PublishWithReply(subject string, data []byte, reply string) {
	fmt.Println("Mock Bus Client: PublishWithReply called for subject:", subject, "data:", string(data), "reply:", reply)
}

func (m *MockBusClient) SubscribeWithQueue(subject string, queue string, handler func(subject string, msg []byte)) error {
	fmt.Println("Mock Bus Client: SubscribeWithQueue called for subject:", subject, "queue:", queue)
	return nil
}

func (m *MockBusClient) GetClientCount() int {
	return 1
}

func (m *MockBusClient) GetMsgHandlerCount() int {
	return 1
}

// Mock ConfigClient
type MockConfigClient struct{}

func (m *MockConfigClient) GetConfig(key string) any {
	fmt.Println("Mock Config Client: GetConfig called for key:", key)
	return nil
}

func (m *MockConfigClient) SetConfig(key string, value any) error {
	fmt.Println("Mock Config Client: SetConfig called for key:", key, "value:", value)
	return nil
}

func (m *MockConfigClient) DeleteConfig(key string) error {
	fmt.Println("Mock Config Client: DeleteConfig called for key:", key)
	return nil
}

func (m *MockConfigClient) PublishConfig(key string, value any) error {
	fmt.Println("Mock Config Client: PublishConfig called for key:", key, "value:", value)
	return nil
}

func (m *MockConfigClient) LoadConfig(file string) error {
	fmt.Println("Mock Config Client: LoadConfig called for file:", file)
	return nil
}

func (m *MockConfigClient) ReloadConfig() error {
	fmt.Println("Mock Config Client: ReloadConfig called")
	return nil
}
