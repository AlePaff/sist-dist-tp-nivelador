import struct
import safe_socket
from lottery.bet import Bet

MESSAGE_TYPE_BET = 1
MESSAGE_TYPE_END = 2
MESSAGE_TYPE_WINNERS = 3
BETS_SEPARATOR = "\n"

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


# separa todas las apuestas por \n y deserializa cada una de ellas
def deserialize_bets_batch(data):
    lines = data.decode("utf-8").split(BETS_SEPARATOR)

    bets = []

    for line in lines:
        if not line:
            continue

        bets.append(deserialize_bet(line.encode("utf-8")))

    return bets

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

# separados por \n a cada ganador
def serialize_winners(winners):
    lines = []

    for bet in winners:
        line = ",".join([
            str(bet.agency_id),
            bet.first_name,
            bet.last_name,
            str(bet.document),
            bet.birthdate,
            str(bet.number),
        ])

        lines.append(line)

    print("ganadores serializados: ", lines)
    return BETS_SEPARATOR.join(lines).encode()