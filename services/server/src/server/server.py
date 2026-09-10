import os
import signal
import socket
import threading
import logger
from protocol.protocol import MESSAGE_TYPE_BET, MESSAGE_TYPE_END, MESSAGE_TYPE_WINNERS, MESSAGE_TYPE_ACK, deserialize_bets_batch, receive_message, send_message, serialize_winners
from lottery.lottery import Lottery


class RoundParticipant:
    def __init__(self, agency_id, bets):
        self.agency_id = agency_id
        self.bets = bets
        self.winners = None


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
        self.quorum_condition = threading.Condition()
        self.pending_round = []
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
            self.pending_round = []
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

            participant = self._receive_bets(client_socket)
            if self._shutdown_event.is_set() or participant is None:
                return
            with self.quorum_condition:
                while participant.winners is None and not self._shutdown_event.is_set():
                    self.quorum_condition.wait()
                if not self._shutdown_event.is_set():
                    winners = participant.winners
                else:
                    winners = None
            # calcula los ganadores
            if winners is not None:
                self._send_winners(winners, client_socket)
    
        except Exception as e:
            if not self._shutdown_event.is_set():
                logger.error("handle-client", logger.LogResult.fail, "err", e)

        finally:
            client_socket.close()
            if client_socket in self._client_sockets:
                self._client_sockets.remove(client_socket)

            
    def _receive_bets(self, client_socket):
        agency_id = None
        bets = []

        while True:
            message_type, payload = receive_message(client_socket)

            if message_type == MESSAGE_TYPE_BET:
                bets_batch = deserialize_bets_batch(payload)
                agency_id = self._handle_bet_batch(bets, bets_batch, client_socket, agency_id)

            elif message_type == MESSAGE_TYPE_END:
                print(f"El cliente terminó de enviar apuestas")
                if agency_id is None:
                    raise ValueError("Se envio primero END. La agencia tiene que mandar al menos una apuesta")
                participant = RoundParticipant(agency_id, bets)
                with self.quorum_condition:
                    self._register_participant(participant)
                return participant

            else:
                raise KeyError("Tipo de mensaje no identificado")

    def _handle_bet_batch(self, bets, bets_batch, client_socket, agency_id):
        if(bets_batch is None or len(bets_batch) == 0):
            logger.error("receive-bets", logger.LogResult.fail, "err", "Se recibio un lote de apuestas vacio, se continua")
            return agency_id

        print("Batch recibido:", bets_batch)

        # guarda el agency_id
        if agency_id is None:
            agency_id = bets_batch[0].agency_id
        if any(bet.agency_id != agency_id for bet in bets_batch):
            raise ValueError("Todas las apuestas deben pertenecer a la misma agencia")
        bets.extend(bets_batch)

        with self.lottery_lock:
            self.lottery.store_bets(bets_batch)

        # enviar mensaje ack de que se recibio el lote correctamente
        send_message(client_socket, MESSAGE_TYPE_ACK, b"")
        return agency_id

    def _register_participant(self, participant):
        self.pending_round.append(participant)      # lo agrega a la lista global de participantes de esta ronda
        if len(self.pending_round) >= self.agency_quorum_min:
            print(f"===Se alcanzó el QUORUM. en el cliente {participant.agency_id}, len(pending_round): {len(self.pending_round)}. Demas clientes en espera")
            current_participants = self.pending_round
            self.pending_round = []     # a partir de aca cualquier cliente nuevo utiliza esta lista

            # para cada participante calcula sus ganadores en la ronda actual
            self._compute_winners_per_round(current_participants)

            # se vacia cuando se alcanza el quorum
            # las apuestas estan guardadas en memoria (all_bets_from_all_agencies) por lo tanto no pasa nada si se borra en disco
            with self.lottery_lock:
                print(f"El cliente {participant.agency_id} vacia el bets.csv")
                with open(self.lottery.storage_path, "w") as f:
                    f.write("")    # truncar el archivo

        self.quorum_condition.notify_all()
        print(f"arranca nuevo quorum: self.pending_round.len() {len(self.pending_round)}")
                
    def _compute_winners_per_round(self, current_participants):
        all_bets_from_all_agencies = [bet for item in current_participants for bet in item.bets]
        for item in current_participants:
            item.winners = [
                bet for bet in all_bets_from_all_agencies
                if bet.agency_id == item.agency_id and self.lottery.has_won(bet)
            ]

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

