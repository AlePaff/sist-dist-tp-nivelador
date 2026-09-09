Redactar un breve informe en donde se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado y los mecanismos para sincronizar la ejecución concurrente.




make up funciona solo en la consola de linux, si estas en windows poner la consola de "bash" o "git bash"

se cambió el makefile y dockercompose de python3 a python. En todo caso cuando se suba al campus va a andar bien, es solo mi entorno que no encuentra python3 pero si python

### Ejercicio 1
Se creó el script scripts/1_generar_compose_clientes.py para generar la cantidad de clientes según indique el enunciado. De acuerdo al punto 1 opcional
Ejemplo de ejecución

```bash
$ python 1_generar_compose_clientes.py 2
> Se generó docker-compose.yaml con 2 clientes.
```



### Ejercicio 2
Ejemplo de ejecución

```bash
$ ./scripts/2_verificar_servidor.sh 
Hello World
```

### Ejercicio 3
Ejemplo de ejecución. luego de "make up"

Configurar INPUTFILE y OUTPUTFILE en docker-compose.yaml

```bash
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress config.input-file=/input/input-0.csv config.output-file=/output/output-0.csv
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO action=process-input-file-from-server result=success agency-id=0
```



### Ejercicio 4
el problema es que por TCP se puede enviar 50 bytes y recibir solo 25, o primero 25 y luego los otros 25 (short write). Entonces para solucionarlo se deberia enviar todo, mediante un loop por ejemplo 



### Ejercicio 5
Una apuesta tiene 6 campos, el agency_id se obtiene del docker compose, mientras que los otros 5 datos de un .csv por ejemplo
Ver bet.py


El protocolo es muy sencillo, cada paquete tiene
┌────────┬──────────┬─────────────────┐
│ TYPE   │ LENGTH   │ PAYLOAD         │
│ 1 byte │ 4 bytes  │ N bytes         │
└────────┴──────────┴─────────────────┘

con 
1 = BET
2 = END
3 = WINNERS


El cliente envia el archivo de apuestas dividiendo cada linea por una apuesta enviando un paquete diferente de acuerdo al protocolo definido anteriormente
y recibe tambien un bet pero de los ganadores??

El servidor recibe estas apuestas, procesa todas y luego le envia a cada cliente los ganadores

Se envian los ganadores separados por \n
Ejemplo
fields="[0,Santiago Lionel,Lorca,30904465,1999-03-17,7574\n0,Camila Rocio,Varela,37130775,1995-05-09,7574]"

bet="{AgencyID:0 FirstName:Santiago Lionel LastName:Lorca Document:30904465 Birthdate:1999-03-17 Number:7574}" err=<nil>
bet="{AgencyID:0 FirstName:Camila Rocio LastName:Varela Document:37130775 Birthdate:1995-05-09 Number:7574}" err=<nil>


el formato es
tipo mensaje = apuesta
apuesta_1_en_binario \n apuesta_2_en_binario \n apuesta_3_en_binario

### Ejercicio 6
Antes se enviaba 1 apuesta por 1 mensaje, ahora se intenta N=BATCH_SIZE apuestas por 1 mensaje

Tras cada batch el server envia un ACK al cliente, indicando que salió todo bien. 
Por ser una implementación simple el cliente espera un ack, si no lo recibe no sigue enviando y se queda tildado o quieto ahí
el servidor procesa el batch pero nunca le informa al cliente que terminó correctamente.

Ejemplo: un posible error es que el cliente envíe una apuesta como string, entonces el servidor lo procesa mal y tira error pero el cliente no tiene manera de saber que ocurrió y sigue mandando paquetes. Esto se soluciona gracias al ack

me estaba fallando el test porque no habia bajado al servicio. Asegurarse con "make down" y luego "make test"


### Ejercicio 7
Hasta ahora
client_socket, _ = server_socket.accept()
acepta una conexión y queda bloqueado, por lo tanto no permite con multiples clientes realmente. solo atiende uno a la vez
en los logs se ve que dice primero client_0 hace todo el procesamiento, luego hace client_1 (aunque es mas dificil de rastrear porque los logs pueden imprimirse en otro orden)

segun entiendo el threading no escala bien en CPython (interprete de python) debido al GIL. CPython si tiene multi threading, es GIL quien lo frena
pero sirve mucho para operaciones I/O-bound (como esperar red, archivos, sockets, etc). Ahí si es util, pero de lo contrario para cosas pesadas es lo mismo que tener un hilo


El servidor tiene que esperar a que hayan terminado como mínimo AGENCY_QUORUM_MIN agencias.

hay un thread por cliente, en miles o millones de clientes esto no escala bien, pero para este tp sirve esta simplificación



El quorum sigue siendo global; el almacenamiento y el cálculo pueden ser locales al hilo. Eso evita el problema actual de que todos carguen bets.csv y puedan recibir ganadores de otras agencias.

faltaría un mecanismo de protección para acceder a server->lottery

Otra forma era guardar un diccionario de Lottery de forma que cada agencia tengo un archivo de apuestas separado, pero se evita esto ya que el objetivo es intentar sincronizar el acceso y simular concurrencia



### Ejercicio 8
-t indica el tiempo de espera (timeout) antes de forzar la eliminación de los contenedores.

SIGTERM primero le avisa que va a cerrarlo y luego SIGKILL lo mata a la fuerza si es que sigue abierto

Se puede usar mutex para diseñar un estado compartido y preguntar si el programa sigue abierto o no, o bien usar la funcionalidad de context que es una especie de propagación entre todo esto


Se puede simular pasandole un archivo muy grande al cliente (por ejemplo input-1.csv) y luego hace "make up" y en otra consola muentras se muestra el envio de apuestas hace "make down". Se podrá ver una salida similar a la siguiente, registrandose correctamente que se hizo una salida _grateful_


```
client_0  | 2026/09/08 21:14:45 INFO action=sigterm-received result=in-progress
client_1  | 2026/09/08 21:14:45 INFO action=sigterm-received result=in-progress
client_1  | 2026/09/08 21:14:45 INFO action=receive-message result=in-progress AAAAAAAAAAAAAAAA=4 payload-size=0
client_1  | 2026/09/08 21:14:45 INFO action=receive-ack result=success !BADKEY="ack recibido del servidor"
client_1  | 2026/09/08 21:14:45 INFO action=send-bets result=in-progress info="shutdown solicitado, se corta el envío"
server    | Exception in thread Thread-2 (_handle_client):
client_1  | 2026/09/08 21:14:45 INFO action=process-input-file-from-server result=success info="cierre graceful por SIGTERM"
server    | Traceback (most recent call last):
server    |   File "/usr/local/lib/python3.14/threading.py", line 1082, in _bootstrap_inner
server    |     self._context.run(self.run)
server    |     ~~~~~~~~~~~~~~~~~^^^^^^^^^^
server    |   File "/usr/local/lib/python3.14/threading.py", line 1024, in run
server    |     self._target(*self._args, **self._kwargs)
server    |     ~~~~~~~~~~~~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
server    |   File "/src/server/server.py", line 68, in _handle_client
server    |     raise e
server    |   File "/src/server/server.py", line 59, in _handle_client
server    |     agency_id = self._receive_bets(client_socket)
server    |   File "/src/server/server.py", line 77, in _receive_bets
server    |     message_type, payload = receive_message(client_socket)
server    |                             ~~~~~~~~~~~~~~~^^^^^^^^^^^^^^^
server    |   File "/src/protocol/protocol.py", line 18, in receive_message
server    |     header = safe_socket.recv_all(socket, 5)
server    |   File "/src/safe_socket/safe_socket.py", line 12, in recv_all
server    |     raise RuntimeError("socket connection broken")
server    | RuntimeError: socket connection broken
client_1 exited with code 0
client_0  | 2026/09/08 21:14:49 INFO action=process-input-file-from-server result=success info="cierre graceful por SIGTERM"
client_0 exited with code 0
server    | 2026/09/08 21:14:49 INFO action=sigterm-received result=in-progress 
server    | 2026/09/08 21:14:49 INFO action=sigterm-received result=in-progress 
server exited with code 0
```









