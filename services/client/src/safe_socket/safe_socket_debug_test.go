package safe_socket

import (
	"fmt"
	"io"
	"testing"
	"time"
)

type FakeReader struct {
	data []byte
}

func (r *FakeReader) Read(p []byte) (int, error) {
	fmt.Println("\n--- FakeReader.Read() ---")
	fmt.Println("Me pidieron:", len(p), "bytes")
	fmt.Println("Tengo:", len(r.data), "bytes")

	if len(r.data) == 0 {
		return 0, io.EOF
	}

	// va a entregar de a 3 bytes por vez, para simular lecturas cortas
	n := 3
	if n > len(r.data) {
		n = len(r.data)
	}

	copy(p, r.data[:n])
	r.data = r.data[n:]

	fmt.Println("Entrego:", n, "bytes")
	fmt.Println("Datos:", string(p[:n]))
	fmt.Println("Quedan:", len(r.data), "bytes")

	time.Sleep(2 * time.Second)

	return n, nil
}

func TestPruebaRecvAll(t *testing.T) {
	fmt.Println("\n========================")
	fmt.Println("PRUEBA RECV_ALL")
	fmt.Println("========================")

	reader := &FakeReader{
		data: []byte("ABCDEFGHIJ"),
	}

	result, err := RecvAll(reader, 10)

	fmt.Println("\n=== RESULTADO ===")
	fmt.Println("Resultado:", string(result))
	fmt.Println("Error:", err)
}

/*
Con el safe_socket.go antes de ser modificado, la salida de la prueba TestPruebaRecvAll era:
========================
PRUEBA RECV_ALL
========================

--- FakeReader.Read() ---
Me pidieron: 10 bytes
Tengo: 10 bytes
Entrego: 3 bytes
Datos: ABC
Quedan: 7 bytes

=== RESULTADO ===
Resultado: ABC
Error: <nil>
--- PASS: TestPruebaRecvAll (2.00s)
=== RUN   TestPruebaSendAll

asumia que no habia error y que habia leido todo correctamente, finalizando asi y provocando un short read, ya que no se leyeron todos los bytes solicitados




Luego de la modificación fué
Ale@DESKTOP-F7PQ3UA MINGW64 ~/Ale/Carrera/Sistemas distribuidos I/Repositorios/sist-dist-tp-nivelador/services/client/src/safe_socket (ejercicio-4)
$ go test -v
=== RUN   TestPruebaRecvAll

========================
PRUEBA RECV_ALL
========================

--- FakeReader.Read() ---
Me pidieron: 10 bytes
Tengo: 10 bytes
Entrego: 3 bytes
Datos: ABC
Quedan: 7 bytes

--- FakeReader.Read() ---
Me pidieron: 7 bytes
Tengo: 7 bytes
Entrego: 3 bytes
Datos: DEF
Quedan: 4 bytes

--- FakeReader.Read() ---
Me pidieron: 4 bytes
Tengo: 4 bytes
Entrego: 3 bytes
Datos: GHI
Quedan: 1 bytes

--- FakeReader.Read() ---
Me pidieron: 1 bytes
Tengo: 1 bytes
Entrego: 1 bytes
Datos: J
Quedan: 0 bytes

=== RESULTADO ===
Resultado: ABCDEFGHIJ
Error: <nil>
--- PASS: TestPruebaRecvAll (8.00s)
=== RUN   TestPruebaSendAll







archivo safe_socket.go antes
package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	_, err := socket.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}
func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	n, err := socket.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff[:n], nil
}







// ejecutar la prueba como "go test" para ver los prints de debug o con "go test -v" si no se llegan a hacer los prints
desde aca
Ale@DESKTOP-F7PQ3UA MINGW64 ~/Ale/Carrera/Sistemas distribuidos I/Repositorios/sist-dist-tp-nivelador/services/client/src/safe_socket (ejercicio-4)

*/
