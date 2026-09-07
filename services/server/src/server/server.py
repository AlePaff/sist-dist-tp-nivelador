import os
import socket
import logger
from protocol.protocol import MESSAGE_TYPE_BET, MESSAGE_TYPE_END, MESSAGE_TYPE_WINNERS, MESSAGE_TYPE_ACK, deserialize_bets_batch, receive_message, send_message, serialize_winners
from lottery.lottery import Lottery


class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        # vaciar archivo bets.csv al iniciar el servidor. Si no existe crearlo
        os.makedirs("/data", exist_ok=True)
        with open("/data/bets.csv", "w") as f:
            f.write("")
        self.lottery = Lottery("/data/bets.csv")
        

    def _handle_client(self, client_socket):
        try:
            logger.info("handle-client", logger.LogResult.in_progress)

            self._receive_bets(client_socket)
            print("debug: termino recibir apuestas, calculando ganadores...")
            winners = self._calculate_winners()
            self._send_winners(winners, client_socket)
    
        except Exception as e:
            raise e

    def _receive_bets(self, client_socket):
        while True:
            message_type, payload = receive_message(client_socket)

            if message_type == MESSAGE_TYPE_BET:
                betsBatch = deserialize_bets_batch(payload)

                print("Batch recibido:", betsBatch)

                self.lottery.store_bets(betsBatch)

                # enviar mensaje ack de que se recibio el lote correctamente
                send_message(client_socket, MESSAGE_TYPE_ACK, b"")

            elif message_type == MESSAGE_TYPE_END:
                print(f"El cliente terminó de enviar apuestas. Se guardan las apuestas")
                break

    def _calculate_winners(self):
        # NOTE: por ahora solo un cliente, luego se hace un quorum para saber a cuantos clientes esperar
        bets = list(self.lottery.load_bets())       # aca se consume el iterador
        print("Cant apuestas recibidas:", len(bets))
        winners = []

        for bet in bets:
            print(f"Evaluando apuesta: {bet}")
            if self.lottery.has_won(bet):
                winners.append(bet)

        print(f"Ganadores: {winners}")
        return winners

    def _send_winners(self, winners, client_socket):
        # Serializamos y enviamos los ganadores.
        payload = serialize_winners(winners)

        send_message(
            client_socket,
            MESSAGE_TYPE_WINNERS,
            payload,
        )

        print("Ganadores enviados al cliente. Fin.")

    
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
