package client

import (
	// imports de la libreria estandar
	"bufio" // para leer y escribir en archivos linea por linea
	"net"
	"os" // acceder a variables de entorno, crear y leer archivo, etc.
	"strings"
	"time"

	// imports de terceros
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
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
	InputFile  string
	OutputFile string
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

// *Client ==> se usan punteros por rendimiento, ya que en Go todo se copia por valor (si tengo datos muy grande para evitar copiarlo todo, uso un puntero)
func (client *Client) Run() error {
	const mainAction = "process-input-file-from-server"
	logger.Info(mainAction, logger.InProgress, "config.input-file", client.config.InputFile, "config.output-file", client.config.OutputFile)

	defer client.conn.Close() // ejecuta este comando al final de la función, sin importar si hubo error o no (similar a un 'finally' en otros lenguajes)

	// abre el archivo de entrada para leerlo
	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail)
		return err
	}
	defer inputFile.Close()

	// crea el archivo de salida para escribir en él (si ya existe lo reemplaza)
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail)
		return err
	}
	defer outputFile.Close()

	// crea un scanner asociado al archivo. lee linea por linea por default
	scanner := bufio.NewScanner(inputFile)

	// lee cada linea del archivo de entrada, la manda al servidor y escribe la respuesta en el archivo de salida
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")

		logger.Info(mainAction, logger.InProgress, "agency-id", client.config.AgencyId, "fields", fields)

		bet := protocol.Bet{
			AgencyID:  client.config.AgencyId,
			FirstName: fields[0],
			LastName:  fields[1],
			Document:  fields[2],
			Birthdate: fields[3],
			Number:    fields[4],
		}

		payload, err := protocol.SerializeBet(bet)
		if err != nil {
			return err
		}

		err = protocol.SendMessage(client.conn, protocol.MessageTypeBet, payload)
		if err != nil {
			return err
		}
	}

	// si hubo un error al leer el archivo scanner.Scan() devuelve false y el error se guarda en scanner.Err()
	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail)
		return err
	}

	// enviar mensaje para finalizar (end)
	err = protocol.SendMessage(client.conn, protocol.MessageTypeEnd, []byte{})
	if err != nil {
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
