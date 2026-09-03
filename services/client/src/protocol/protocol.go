package protocol

import (
	"encoding/binary"
	"io"

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

func SerializeBet(bet Bet) ([]byte, error)

func DeserializeBet(data []byte) (Bet, error)
