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

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
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

	if err := enviarApuestas(client); err != nil {
		logger.Error(mainAction, logger.Fail)
		return err
	}

	// recibir mensaje de ganadores
	if err := recibirGanadores(client); err != nil {
		logger.Error(mainAction, logger.Fail)
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}

func enviarApuestas(client *Client) error {
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

	batch := make([]protocol.Bet, 0, client.config.BatchSize)
	// lee cada linea del archivo de entrada, la manda al servidor y escribe la respuesta en el archivo de salida
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")

		logger.Info("send-bets", logger.InProgress, "agency-id", client.config.AgencyId, "fields", fields)

		bet := protocol.Bet{
			AgencyID:  client.config.AgencyId,
			FirstName: fields[0],
			LastName:  fields[1],
			Document:  fields[2],
			Birthdate: fields[3],
			Number:    fields[4],
		}

		batch = append(batch, bet)

		// enviar batch, sigue el mismo comportamiento que el ejercicio5 si es batch_siz = 1
		if len(batch) == client.config.BatchSize {
			if err := enviarBatch(client.conn, batch); err != nil {
				return err
			}

			batch = batch[:0] // vaciar el batch para la siguiente iteración
		}
	}
	// si hubo un error al leer el archivo scanner.Scan() devuelve false y el error se guarda en scanner.Err()
	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail)
		return err
	}

	// si no se llenó el batch, enviar lo que quedó
	if len(batch) > 0 {
		if err := enviarBatch(client.conn, batch); err != nil {
			return err
		}
	}

	// enviar mensaje de que termino de enviar batch
	logger.Info("send-end", logger.InProgress, "envia mensaje end al servidor", client.config.AgencyId)
	if err := protocol.SendMessage(client.conn, protocol.MessageTypeEnd, []byte{}); err != nil {
		return err
	}

	return nil
}

func enviarBatch(client_conn net.Conn, batch []protocol.Bet) error {
	payload, err := protocol.SerializeBetsBatch(batch)
	if err != nil {
		return err
	}

	// enviar mensaje
	err = protocol.SendMessage(client_conn, protocol.MessageTypeBet, payload)
	if err != nil {
		return err
	}
	logger.Info("send-bets", logger.InProgress, "mensajito", "cliente manda batch -------")

	// espera recibir el mensaje ack
	logger.Info("receive-ack", logger.InProgress, "ack", "espera recibir ack del servidor")
	message, err := protocol.ReceiveMessage(client_conn)
	if err != nil {
		return err
	}

	if message.Type != protocol.MessageTypeAck {
		logger.Error("receive-ack-error", logger.Fail, "unexpected-message-type", message.Type)
		return err
	}
	logger.Info("receive-ack", logger.Success, "ack recibido del servidor")

	return nil
}

func recibirGanadores(client *Client) error {
	BETS_SEPARATOR := "\n"
	// recibe mensaje ganadores
	message, err := protocol.ReceiveMessage(client.conn)
	if err != nil {
		return err
	}

	if message.Type != protocol.MessageTypeWinners {
		logger.Error("receive-winners", logger.Fail, "unexpected-message-type", message.Type)
		return err
	}

	// imprimir payload y tipo de mensaje
	logger.Info("receive-winners", logger.InProgress, "TIPO DE MSG", message.Type, "payload", string(message.Payload))

	winners, err := protocol.DeserializeWinners(message.Payload)
	if err != nil {
		return err
	}

	logger.Info("receive-winners", logger.Success, "ganadores recibidos: ", winners)

	// guardar en archivo de salida
	outputFile, err := os.Create(client.config.OutputFile)

	for _, winner := range winners {
		_, err := outputFile.WriteString(
			winner.FirstName + "," +
				winner.LastName + "," +
				winner.Document + "," +
				winner.Birthdate + "," +
				winner.Number + BETS_SEPARATOR,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
