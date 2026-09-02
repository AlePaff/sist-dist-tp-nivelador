package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0

	for totalSent < len(bytes) {
		n, err := socket.Write(bytes[totalSent:])

		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}

		totalSent += n
	}

	return nil

}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	// se podria arreglar mas facil con io.ReadFull, pero el enunciado pide WriteAll y ReadAll

	buff := make([]byte, size)
	totalReceived := 0

	for totalReceived < size {
		// recibe la cantidad de bytes que faltan para completar el tamaño solicitado y si hay un error lo devuelve
		n, err := socket.Read(buff[totalReceived:])

		// si se recibieron todos los bytes solicitados, devuelve el buffer (ignorando el error, si lo hubiera)
		if totalReceived == size {
			return buff, nil
		}

		// si hubo un error, se descartan los bytes recibidos y se devuelve el error
		if err != nil {
			return nil, err
		}

		// si se cerró la conexión
		if n == 0 {
			return nil, io.ErrUnexpectedEOF
		}

		totalReceived += n
	}

	return buff, nil

}
