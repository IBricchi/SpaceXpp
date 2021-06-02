package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type MQTTClient struct {
	client mqtt.Client
	logger *zap.Logger
}

func InitMQTT(ctx context.Context, logger *zap.Logger, mqttBrokerURL string) (*MQTTClient, error) {
	tlsConfig, err := NewTlsConfig()
	if err != nil {
		return &MQTTClient{}, fmt.Errorf("server: mqtt: failed to get TLS config: %w", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBrokerURL)
	opts.SetTLSConfig(tlsConfig)
	opts.SetClientID("SpaceXpp_server")
	opts.SetCleanSession(true)
	opts.SetConnectRetry(true)

	opts.OnConnect = mqttConnectHandler
	opts.OnConnectionLost = mqttConnectLostHandler

	return &MQTTClient{
		client: mqtt.NewClient(opts),
		logger: logger,
	}, nil
}

func (m *MQTTClient) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("server: mqtt: failed to connect to broker: %w", token.Error())
	}
	return nil
}

func (m *MQTTClient) Disconnect() {
	m.client.Disconnect(100)

	m.logger.Info("Disconnected from MQTT broker successfully")
}

var testStatusMessagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var mqttConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT broker successfully")

	// Subscribe to topics
	if token := client.Subscribe("/test/status", 0, testStatusMessagePubHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("server: mqtt: failed to subscribe to /test/status: %v", token.Error())
	}
	fmt.Println("Subscribed to topic: /test/status")

	// Subscribe to topics
	if token := client.Subscribe("/feedback/instruction", 2, instructionFeedPubHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("server: mqtt: failed to subscribe to /feedback/instruction: %v", token.Error())
	}
	fmt.Println("Subscribed to topic: /feedback/instruction")

}

var mqttConnectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect to MQTT broker lost: %v", err)
}

func NewTlsConfig() (*tls.Config, error) {
	certpool := x509.NewCertPool()

	ca, err := ioutil.ReadFile("cert/mqtt_ca_cert.pem")
	if err != nil {
		return nil, fmt.Errorf("server: mqtt: failed to read CA certificate: %w", err)
	}

	certpool.AppendCertsFromPEM(ca)

	return &tls.Config{
		RootCAs: certpool,
	}, nil
}

func (m *MQTTClient) publish(topic string, data string, qos byte) {
	token := m.client.Publish(topic, qos, false, data)
	go func() {
		token.Wait()
		if err := token.Error(); err != nil {
			m.logger.Error("server: mqttGeneral: failed to publish mqtt message", zap.Error(err))
		}
	}()
}

// Subscribing to instruction feed
var stopData string
var instructionFeedPubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	s := strings.Split(string(msg.Payload()), ":")
	value := s[1]
	v, _ := strconv.Atoi(value) // No error checking as this is supposed to fail for stop instructions

	var instruction driveInstruction
	if s[0] == "F" {
		instruction.instruction = "forward"
		instruction.value = v
		updateMap(instruction)
	} else if s[0] == "R" {
		instruction.instruction = "turnRight"
		instruction.value = v
		updateMap(instruction)
	} else if s[0] == "L" {
		instruction.instruction = "turnLeft"
		instruction.value = v
		updateMap(instruction)
	} else if s[0] == "X" {
		instruction.instruction = "nil"
		instruction.value = 0
		updateMap(instruction)
	} else if s[0] == "S" {
		if stashedDriveInstruction.instruction == "forward" { // wait for second part of stop instruction to update map and stop
			stopData = value
		} else { // turning => update map without stopping
			updateMapWithObstructionWhileTurning(value)
		}
	} else if s[0] == "SD" {
		stop(v, stopData)
		stopData = ""
	} else if s[0] == "B" {
		// Ignore backwards instruction that are used for distance correction (drive only)
	} else {
		fmt.Println("server: mqttGeneral: unknown drive instruction")
	}
}
