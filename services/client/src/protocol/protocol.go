package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

/*
tipos de mensajes
*/
const (
	MessageTypeBet byte = 1
	MessageTypeEnd byte = 2
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

	return Message{
		Type:    messageType,
		Payload: payload,
	}, nil
}

type Bet struct {
	AgencyID  uint32
	FirstName string
	LastName  string
	Document  uint32
	Birthdate string
	Number    uint32
}

// recibe una apuesta y la serieliza en un slice de bytes
// el payload va separado por comas
func SerializeBet(bet Bet) ([]byte, error) {
	fields := []string{
		strconv.FormatUint(uint64(bet.AgencyID), 10), // 10 es la base decimal
		bet.FirstName,
		bet.LastName,
		strconv.FormatUint(uint64(bet.Document), 10),
		bet.Birthdate,
		strconv.FormatUint(uint64(bet.Number), 10),
	}

	return []byte(strings.Join(fields, ",")), nil
}

// recibe un slice de bytes y lo deserializa en una apuesta Bet
func DeserializeBet(data []byte) (Bet, error) {
	fields := strings.Split(string(data), ",")

	// esperamos 6 campos: agency_id, first_name, last_name, document, birthdate, number
	if len(fields) != 6 {
		return Bet{}, fmt.Errorf("invalid bet: expected 6 fields, got %d", len(fields))
	}

	agencyID, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return Bet{}, fmt.Errorf("invalid agency id: %w", err)
	}

	document, err := strconv.ParseUint(fields[3], 10, 32)
	if err != nil {
		return Bet{}, fmt.Errorf("invalid document: %w", err)
	}

	number, err := strconv.ParseUint(fields[5], 10, 32)
	if err != nil {
		return Bet{}, fmt.Errorf("invalid number: %w", err)
	}

	return Bet{
		AgencyID:  uint32(agencyID),
		FirstName: fields[1],
		LastName:  fields[2],
		Document:  uint32(document),
		Birthdate: fields[4],
		Number:    uint32(number),
	}, nil
}
