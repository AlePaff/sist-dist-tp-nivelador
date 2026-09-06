import struct

import safe_socket
from src_frozen.lottery.bet import Bet


MESSAGE_TYPE_BET = 1
MESSAGE_TYPE_END = 2


def send_message(socket, message_type, payload):
    packet = struct.pack("!BI", message_type, len(payload)) + payload     # !BI significa: ! = big-endian, B = unsigned char (1 byte), I = unsigned int (4 bytes)

    safe_socket.send_all(socket, packet)


def receive_message(socket):
    header = safe_socket.recv_all(socket, 5)

    message_type, payload_size = struct.unpack("!BI", header)

    payload = safe_socket.recv_all(socket, payload_size)

    return message_type, payload


def serialize_bet(bet):
    fields = [
        str(bet.agency_id),
        str(bet.first_name),
        str(bet.last_name),
        str(bet.document),
        str(bet.birthdate),
        str(bet.number),
    ]

    return ",".join(fields).encode("utf-8")



def deserialize_bet(data):
    fields = data.decode("utf-8").split(",")

    if len(fields) != 6:
        raise ValueError(
            f"invalid bet: expected 6 fields, got {len(fields)}"
        )

    return Bet(
        agency_id=int(fields[0]),
        first_name=fields[1],
        last_name=fields[2],
        document=int(fields[3]),
        birthdate=fields[4],
        number=int(fields[5]),
    )