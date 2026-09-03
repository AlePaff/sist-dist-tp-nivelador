package protocol

import (
	"bytes"
	"fmt"
	"testing"
)

func TestBetSerialization(t *testing.T) {
	// Bet que simula la que leería el cliente del CSV
	originalBet := Bet{
		AgencyID:  0,
		FirstName: "Santiago Lionel",
		LastName:  "Lorca",
		Document:  30904465,
		Birthdate: "1999-03-17",
		Number:    7574,
	}

	// Simulamos la conexión entre cliente y servidor
	var connection bytes.Buffer

	// CLIENTE:
	// serializa la Bet
	// imprimir el payload
	fmt.Printf("CLIENTE: Payload antes de serializar: %+v\n", originalBet)
	payload, err := SerializeBet(originalBet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("CLIENTE: Payload serializado: %v\n", payload)

	// CLIENTE:
	// envía el mensaje
	err = SendMessage(&connection, MessageTypeBet, payload)
	if err != nil {
		t.Fatal(err)
	}

	// SERVIDOR:
	// recibe el mensaje. Simulando que el servidor usa go (en realidad esta en python)
	fmt.Printf("SERVIDOR: Recibiendo mensaje serializado...\n")
	message, err := ReceiveMessage(&connection)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("SERVIDOR: Mensaje recibido: Type=%d, Payload=%v\n", message.Type, message.Payload)
	// verificamos que recibió un mensaje BET
	if message.Type != MessageTypeBet {
		t.Fatalf("expected message type %d, got %d", MessageTypeBet, message.Type)
	}

	// SERVIDOR:
	// deserializa la Bet
	receivedBet, err := DeserializeBet(message.Payload)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("SERVIDOR: Bet deserializada: %+v\n", receivedBet)

	// verificamos que la Bet recibida sea igual a la original
	if receivedBet != originalBet {
		t.Fatalf("expected %+v, got %+v", originalBet, receivedBet)
	}
}
