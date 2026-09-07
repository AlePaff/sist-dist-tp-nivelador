package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

/*
tipos de mensajes
*/
const (
	MessageTypeBet     byte = 1
	MessageTypeEnd     byte = 2
	MessageTypeWinners byte = 3
	MessageTypeAck     byte = 4
	BetsSeparator           = "\n"
)

type Message struct {
	Type    byte
	Payload []byte
}

// envia un mensaje con el tipo y payload especificados a traves del socket
func SendMessage(socket io.Writer, messageType byte, payload []byte) error {
	payloadLength := len(payload)
	packet := make([]byte, 5+payloadLength) // 1 byte para el tipo de mensaje y 4 para el tamaño del payload y el resto para el payload

	packet[0] = messageType
	binary.BigEndian.PutUint32(packet[1:5], uint32(payloadLength))
	copy(packet[5:], payload) // se copia el payload en el paquete a partir del byte 5
	logger.Info(
		"send-message",
		logger.InProgress,
		"message-type", messageType,
		"payload-size", payloadLength,
		"packet-size", len(packet),
	)
	if err := safe_socket.SendAll(socket, packet); err != nil {
		return err
	}

	return nil
}

func ReceiveMessage(socket io.Reader) (Message, error) {
	header, err := safe_socket.RecvAll(socket, 5)
	if err != nil {
		return Message{}, err
	}

	messageType := header[0]
	payloadSize := binary.BigEndian.Uint32(header[1:])

	payload, err := safe_socket.RecvAll(socket, int(payloadSize))
	if err != nil {
		return Message{}, err
	}

	logger.Info("receive-message", logger.InProgress, "AAAAAAAAAAAAAAAA", messageType, "payload-size", payloadSize)

	return Message{
		Type:    messageType,
		Payload: payload,
	}, nil
}

// para evitar convertir de string a entero al leer, para luego serializar y enviarlo, luego convertirlo de nuevo a entero
//
//	al recibirlo y deserializarlo, se dejan todos como string
type Bet struct {
	AgencyID  string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

// recibe una apuesta y la serieliza en un slice de bytes
// el payload va separado por comas
func SerializeBet(bet Bet) ([]byte, error) {
	fields := []string{
		bet.AgencyID,
		bet.FirstName,
		bet.LastName,
		bet.Document,
		bet.Birthdate,
		bet.Number,
	}

	return []byte(strings.Join(fields, ",")), nil
}

// va serializando las apuestas por batch en binario separadas por \n
func SerializeBetsBatch(bets []Bet) ([]byte, error) {
	lines := make([]string, 0, len(bets))

	for _, bet := range bets {
		data, err := SerializeBet(bet)
		if err != nil {
			return nil, err
		}

		lines = append(lines, string(data))
	}

	return []byte(strings.Join(lines, BetsSeparator)), nil
}

// recibe un slice de bytes y lo deserializa en una apuesta Bet
func DeserializeBet(data []byte) (Bet, error) {
	fields := strings.Split(string(data), ",")

	// esperamos 6 campos: agency_id, first_name, last_name, document, birthdate, number
	if len(fields) != 6 {
		return Bet{}, fmt.Errorf("invalid bet: expected 6 fields, got %d", len(fields))
	}

	return Bet{
		AgencyID:  fields[0],
		FirstName: fields[1],
		LastName:  fields[2],
		Document:  fields[3],
		Birthdate: fields[4],
		Number:    fields[5],
	}, nil
}

func DeserializeWinners(data []byte) ([]Bet, error) {
	fields := strings.Split(string(data), BetsSeparator)

	winners := make([]Bet, 0, len(fields))

	logger.Info("deserialize-winners", logger.InProgress, "fields", fields)

	for _, field := range fields {
		if field == "" {
			continue
		}

		bet, err := DeserializeBet([]byte(field))
		logger.Info("deserialize-winners", logger.InProgress, "bet", bet, "err", err)
		if err != nil {
			return nil, err
		}

		winners = append(winners, bet)
	}

	// imrpimir los ganadores deserializados
	logger.Info("deserialize-winners", logger.InProgress, "winners", winners)

	return winners, nil
}
