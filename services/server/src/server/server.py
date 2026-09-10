import os
import signal
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
        with open("/data/bets.csv", "w") as f:
            f.write("")
        self.lottery = Lottery("/data/bets.csv")
        self.finished_agencies = set()          # set para evitar contar dos veces a una agencia
        self.quorum_condition = threading.Condition()
        self.lottery_lock = threading.Lock()

        # manejo de SIGTERM
        self._shutdown_event = threading.Event()        # "evento" Shutdown
        self._server_socket = None
        self._client_sockets = []
        self._client_threads = []
        signal.signal(signal.SIGTERM, self._handle_sigterm)     # cuando se detecta SIGTERM ejecutar la funcion _handle_sigterm
        
    def _handle_sigterm(self, signum, frame):
        logger.info("sigterm-received", logger.LogResult.in_progress)
        self._shutdown_event.set()      # se setea a True el evento "shutdown solicitado"

        # despierta threads que puedan estar esperando el quorum para que hagan shutdown
        with self.quorum_condition:
            self.quorum_condition.notify_all()

        # cierra el socket de escucha para desbloquear accept()
        if self._server_socket is not None:
            try:
                self._server_socket.close()
            except OSError:
                pass

        # cierra los sockets de clientes activos para desbloquear recv()
        for s in self._client_sockets:
            try:
                s.close()
            except OSError:
                pass

    def _handle_client(self, client_socket):
        self._client_sockets.append(client_socket)

        try:
            logger.info("handle-client", logger.LogResult.in_progress)

            if self._shutdown_event.is_set():
                return

            agency_id = self._receive_bets(client_socket)
            if self._shutdown_event.is_set() or agency_id is None:
                return
            print(f"debug: termino recibir apuestas de {agency_id}, calculando ganadores...")
            winners = self._calculate_winners(agency_id)
            self._send_winners(winners, client_socket)
    
        except Exception as e:
            if not self._shutdown_event.is_set():
                raise e

        finally:
            client_socket.close()

    def _receive_bets(self, client_socket):
        agency_id = None

        while True:
            message_type, payload = receive_message(client_socket)

            if message_type == MESSAGE_TYPE_BET:
                bets_batch = deserialize_bets_batch(payload)

                if(bets_batch is None or len(bets_batch) == 0):
                    logger.error("receive-bets", logger.LogResult.fail, "err", "Se recibio un lote de apuestas vacio, se continua")
                    continue

                print("Batch recibido:", bets_batch)

                # guarda el agency_id
                if agency_id is None:
                    agency_id = bets_batch[0].agency_id

                with self.lottery_lock:
                    self.lottery.store_bets(bets_batch)

                # enviar mensaje ack de que se recibio el lote correctamente
                send_message(client_socket, MESSAGE_TYPE_ACK, b"")

            elif message_type == MESSAGE_TYPE_END:
                print(f"El cliente terminó de enviar apuestas. Se guardan las apuestas")
                with self.quorum_condition:         # toma el lock, al salir de aqui se libera
                    self.finished_agencies.add(agency_id)       # es para que solo uno a la vez pueda agregarse en finished_agencies

                    self.quorum_condition.notify_all()      # despierta a los threads que estan esperando el quorum

                    # si aún no se llegó al quorum, pone a dormir al hilo actual. se asegura tambien de que no este el evento de shutdown activado
                    while len(self.finished_agencies) < self.agency_quorum_min and not self._shutdown_event.is_set():         # solo saldra del ciclo dormir->ser despertado->comprobar condicion->dormir, cuando se cumpla la condición
                        self.quorum_condition.wait()        #aqui tambien se libera el lock del "with"
                if self._shutdown_event.is_set():
                    return None
                break
        return agency_id

    def _calculate_winners(self, agency_id):
        with self.lottery_lock:
            bets = list(self.lottery.load_bets())       # aca se consume el iterador
            
        print("Cant apuestas recibidas:", len(bets))
        winners = []

        for bet in bets:
            print(f"Evaluando apuesta: {bet}")
            if bet.agency_id == agency_id and self.lottery.has_won(bet):
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
            self._server_socket = server_socket
            # escucha conexiones entrantes en la dirección y puerto especificados
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while not self._shutdown_event.is_set():
                # acepta una conexión entrante y obtiene el socket del cliente
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()

                    thread = threading.Thread(
                        target=self._handle_client,
                        args=(client_socket,)
                    )
                    # los va guardando para que en caso que llegue el evento SIGTERM entonces itere a cada uno y los vaya cerrando
                    self._client_threads.append(thread)
                    thread.start()
                    
                except Exception as e:
                    if self._shutdown_event.is_set():
                        break
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

        self._handle_sigterm(None, None)
        for thread in self._client_threads:
            thread.join()

