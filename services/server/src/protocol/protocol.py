import struct

import safe_socket


MESSAGE_TYPE_BET = 1
MESSAGE_TYPE_END = 2


def send_message(socket, message_type, payload):
    header = struct.pack("!BI", message_type, len(payload))     # !BI significa: ! = big-endian, B = unsigned char (1 byte), I = unsigned int (4 bytes)

    safe_socket.send_all(socket, header)
    safe_socket.send_all(socket, payload)


def receive_message(socket):
    header = safe_socket.recv_all(socket, 5)

    message_type, payload_size = struct.unpack("!BI", header)

    payload = safe_socket.recv_all(socket, payload_size)

    return message_type, payload