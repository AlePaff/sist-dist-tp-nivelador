import socket
import logger
from protocol.protocol import MESSAGE_TYPE_BET, MESSAGE_TYPE_END, MESSAGE_TYPE_WINNERS, deserialize_bet, receive_message, send_message, serialize_winner
from lottery.lottery import Lottery


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        # vaciar archivo bets.csv al iniciar el servidor. Si no existe crearlo
        with open("/data/bets.csv", "w") as f:
            f.write("")
        self.lottery = Lottery("/data/bets.csv")
        

    def _handle_client(self, client_socket):
        action = "handle-client"
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                message_type, payload = receive_message(client_socket)

                if message_type == MESSAGE_TYPE_BET:
                    bet = deserialize_bet(payload)

                    print("Apuesta recibida:", bet)

                    self.lottery.store_bets([bet])

                elif message_type == MESSAGE_TYPE_END:
                    print(f"El cliente {bet.agency_id} terminó de enviar apuestas")
                    break
        except Exception as e:

            raise e


        # NOTE: por ahora solo un cliente, luego se hace un quorum para saber a cuantos clientes esperar
        bets = list(self.lottery.load_bets())       # aca se consume el iterador
        print("Cant apuestas recibidas:", len(bets))
        winners = []


        for bet in bets:
            print(f"Evaluando apuesta: {bet}")
            if self.lottery.has_won(bet):
                winners.append(bet)

        # envia mensajes individuales de ganadores a cada cliente
        for winner in winners:
            print(f"Ganador: {winner}")
            # Serializamos y enviamos los ganadores.
            payload = serialize_winner(winner)

            send_message(
                client_socket,
                MESSAGE_TYPE_WINNERS,
                payload,
            )

            # TODO: Para mas adelante, habria que enviar un mensaje "end" para que el cliente sepa que ya no hay mas ganadores

        print("Termino de informar ganadores")




    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            # escucha conexiones entrantes en la dirección y puerto especificados
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                # acepta una conexión entrante y obtiene el socket del cliente
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
