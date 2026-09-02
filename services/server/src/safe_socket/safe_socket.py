import socket


def recv_all(socket: socket.socket, size):
    received = bytearray()      # para ir acumulando los bytes recibidos (simil received += chunk)

    while len(received) < size:
        chunk = socket.recv(size - len(received))

        # si se cerró la conexión
        if not chunk:
            raise RuntimeError("socket connection broken")

        received.extend(chunk)

    return bytes(received)
    return socket.recv(size)


def send_all(socket: socket.socket, bytes):
    total_sent = 0

    while total_sent < len(bytes):
        # me devuelve un entero indicando la cantidad de bytes enviados
        # que puede ser menor que la longitud de los bytes que quiero enviar
        sent = socket.send(bytes[total_sent:])

        # trato a sent == 0 como un error, ya que significa que la conexión se rompió
        if sent == 0:
            raise RuntimeError("socket connection broken")

        total_sent += sent
    return total_sent
