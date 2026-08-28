package client

import (
	// imports de la libreria estandar
	"net"
	"time"

	// imports de terceros
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_BUFFER_SIZE = 512
const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	// si la conexión falló devuelve un puntero nulo y el error
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
	// devuelve un puntero a cliente y nil si no hay error
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port) // devuelve un net.Conn que permite enviar y recibir datos a traves de la conexion TCP
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		// si se conectó exitosamente, loguea el éxito y rompe el bucle
		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close() // ejecuta este comando al final de la función, sin importar si hubo error o no (similar a un 'finally' en otros lenguajes)

	// envia N mensajes al servidor y espera la respuesta de cada uno
	for messageId := range ECHO_CLIENT_MESSAGE_AMOUNT {
		// slice, para pasar argumentos al logger
		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)

		clientMessage := client.config.AgencyId

		// convierte el string en bytes
		// manda esos bytes al servidor
		if err := safe_socket.SendAll(client.conn, []byte(clientMessage)); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}

		// espera la respuesta del servidor
		responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		// convierte los bytes en string y compara con el mensaje original (verifica el echo)
		if string(responseBuffer) != clientMessage {
			logger.Error("check-response", logger.Fail, messageArgs...)
			return err
		}

		time.Sleep(ECHO_CLIENT_MESSAGE_DELAY_MS * time.Millisecond)
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
