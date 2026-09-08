import os
import socket
import threading
import logger
from protocol.protocol import MESSAGE_TYPE_BET, MESSAGE_TYPE_END, MESSAGE_TYPE_WINNERS, MESSAGE_TYPE_ACK, deserialize_bets_batch, receive_message, send_message, serialize_winners
from lottery.lottery import Lottery


class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.agency_quorum_min = agency_quorum_min
        # crea el directorio donde se guardaran los archivos de apuestas de cada cliente
        os.makedirs("/data", exist_ok=True)
        self.lotteries = {}
        self.finished_agencies = set()          # set para evitar contar dos veces a una agencia
        self.quorum_condition = threading.Condition()
        

    def _handle_client(self, client_socket):
        try:
            logger.info("handle-client", logger.LogResult.in_progress)

            agency_id = self._receive_bets(client_socket)
            print(f"debug: termino recibir apuestas de {agency_id}, calculando ganadores...")
            winners = self._calculate_winners(agency_id)
            self._send_winners(winners, client_socket)
    
        except Exception as e:
            raise e

    def _receive_bets(self, client_socket):
        agency_id = None

        while True:
            message_type, payload = receive_message(client_socket)

            if message_type == MESSAGE_TYPE_BET:
                betsBatch = deserialize_bets_batch(payload)

                if(betsBatch is None or len(betsBatch) == 0):
                    logger.error("receive-bets", logger.LogResult.fail, "err", "Se recibio un lote de apuestas vacio, se continua")
                    continue

                print("Batch recibido:", betsBatch)

                # guarda el agency_id
                if agency_id is None:
                    agency_id = betsBatch[0].agency_id

                if agency_id not in self.lotteries:
                    # durante la inicialización deja el archivo vacio
                    with open(f"/data/bets_{agency_id}.csv", "w") as f:
                        f.write("")
                    self.lotteries[agency_id] = Lottery(f"/data/bets_{agency_id}.csv")

                self.lotteries[agency_id].store_bets(betsBatch)

                # enviar mensaje ack de que se recibio el lote correctamente
                send_message(client_socket, MESSAGE_TYPE_ACK, b"")

            elif message_type == MESSAGE_TYPE_END:
                print(f"El cliente terminó de enviar apuestas. Se guardan las apuestas")
                with self.quorum_condition:         # toma el lock, al salir de aqui se libera
                    self.finished_agencies.add(agency_id)       # es para que solo uno a la vez pueda agregarse en finished_agencies

                    self.quorum_condition.notify_all()      # despierta a los threads que estan esperando el quorum

                    # si aún no se llegó al quorum, pone a dormir al hilo actual
                    while len(self.finished_agencies) < self.agency_quorum_min:         # solo saldra del ciclo dormir->ser despertado->comprobar condicion->dormir, cuando se cumpla la condición
                        self.quorum_condition.wait()        #aqui tambien se libera el lock del "with"
                break
        return agency_id

    def _calculate_winners(self, agency_id):
        bets = list(self.lotteries[agency_id].load_bets())       # aca se consume el iterador
        print("Cant apuestas recibidas:", len(bets))
        winners = []

        for bet in bets:
            print(f"Evaluando apuesta: {bet}")
            if self.lotteries[agency_id].has_won(bet):
                winners.append(bet)

        print(f"Ganadores: {winners} de la agencia {agency_id}")
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

                    thread = threading.Thread(
                        target=self._handle_client,
                        args=(client_socket,)
                    )
                    thread.start()
                    
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

