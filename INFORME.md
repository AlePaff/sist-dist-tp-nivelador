## TP Nivelador: Docker, Comunicaciones y Concurrencia
### Consideraciones generales
Para ejecutar `make up`, `make down` y demás comandos desde Windows, se debe utilizar una consola compatible con Bash, como Git Bash.

Se modificaron el Makefile y docker-compose.yaml para utilizar python en lugar de python3. Esto se debe únicamente a la configuración del entorno local de Windows, donde python está disponible pero python3 no. En el entorno de entrega esto no debería representar un inconveniente.

Se dejaron comentarios a lo largo del codigo tal vez algo redundantes pero muy utiles para mi como estudiante ya que me ayudaron a entender y aprender a lo largo de la realización del TP. Decidí dejarlos en la entrega para poder consultarlos mas adelante en futuros TPs de manera sencilla.

### Ejercicio 1
Se creó el script `scripts/1_generar_compose_clientes.py` para generar automáticamente la cantidad de clientes indicada por parámetro, de acuerdo con el punto opcional del ejercicio

Ejemplo de ejecución

```bash
$ python 1_generar_compose_clientes.py 2
> Se generó docker-compose.yaml con 2 clientes.
```


### Ejercicio 2
Se expusieron los puertos del servidor para poder conectarse desde el equipo anfitrión mediante netcat. El script creado se encuentra en `scripts/2_verificar_servidor.sh`

Ejemplo de ejecución

```bash
$ ./scripts/2_verificar_servidor.sh 
Hello World
```

### Ejercicio 3
Se modificó el cliente para leer `INPUT_FILE` línea por línea y enviar cada apuesta al servidor. Las respuestas se persisten en `OUTPUT_FILE` (ver docker-compose.yaml), utilizando un volumen de Docker para que el archivo sea accesible desde el equipo anfitrión y no quede unicamente en el contenedor

Ejemplo de ejecución. luego de `make up`

```bash
client_0  | 2026/08/28 06:55:34 INFO result=in-progress config.input-file=/input/input-0.csv config.output-file=/output/output-0.csv
client_0  | 2026/08/28 06:55:34 INFO result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO result=in-progress agency-id=0
client_0  | 2026/08/28 06:55:34 INFO result=success agency-id=0
```


Para solucionarlo, send_all y recv_all realizan operaciones repetidas hasta completar la cantidad de bytes esperada o detectar un error de comunicación. Una escritura que devuelve cero bytes se considera un short write y evita un loop infinito.


### Ejercicio 4
El problema es que TCP no garantiza que una operación de lectura o escritura procese todos los bytes solicitados. Por ejemplo, se pueden enviar 50 bytes y recibirlos en dos lecturas de 25 bytes cada una (short read), o escribir solamente una parte de los datos (short write)

Para solucionarlo, send_all y recv_all realizan operaciones repetidas hasta completar la cantidad de bytes esperada o detectar un error de comunicación.

Tanto en safe_socket.go como en safe_socket.py se dejó explicitamente fuera la validación en caso que se envien 0 bytes para poder correr los tests de short read y write ya que uno de ellos esta esperando que al enviarse 0 bytes el programa no falle, lo normal sería arrojar un error. Por eso se dejó comentada esta parte del codigo

### Ejercicio 5
Se implementó un protocolo de comunicación entre cliente y servidor que permite enviar las apuestas y los resultados del sorteo

El protocolo es muy sencillo, cada paquete tiene
┌────────┬──────────┬─────────────────┐
│ TYPE   │ LENGTH   │ PAYLOAD         │
│ 1 byte │ 4 bytes  │ N bytes         │
└────────┴──────────┴─────────────────┘

con 
1 = BET
2 = END
3 = WINNERS
4 = ACK

Una apuesta contiene seis campos: agency_id, first_name, last_name, document, birthdate y number. El agency_id se obtiene de la configuración del cliente, mientras que los demás datos se obtienen del archivo .csv

El cliente serializa las apuestas y las envía al servidor. Cuando termina de enviar las apuestas, envía un mensaje END. El servidor almacena las apuestas mediante `Lottery.store_bets` y calcula los ganadores de la ronda utilizando `has_won`. Finalmente envía al cliente el listado de ganadores y el cliente persiste estos datos en `OUTPUT_FILE`.

Los ganadores se serializan utilizando \n como separador entre apuestas.

Ejemplo
```
fields="[0,Santiago Lionel,Lorca,30904465,1999-03-17,7574\n0,Camila Rocio,Varela,37130775,1995-05-09,7574]"

bet="{AgencyID:0 FirstName:Santiago Lionel LastName:Lorca Document:30904465 Birthdate:1999-03-17 Number:7574}" err=<nil>
bet="{AgencyID:0 FirstName:Camila Rocio LastName:Varela Document:37130775 Birthdate:1995-05-09 Number:7574}" err=<nil>
```

El formato es
```
tipo mensaje = apuesta
apuesta_1_en_binario \n apuesta_2_en_binario \n apuesta_3_en_binario
```

### Ejercicio 6
Se modificó el protocolo para permitir enviar varias apuestas dentro de un mismo mensaje. 

La cantidad de apuestas por mensaje se configura mediante BATCH_SIZE.

Luego de cada batch, el servidor envía un ACK al cliente indicando que todas las apuestas del lote fueron procesadas correctamente. El cliente espera este ACK antes de continuar enviando el siguiente lote. Se hace un manejo simple por tratarse de un protocolo sencillo.

De esta forma, si el servidor detecta un error al procesar un batch, el cliente no continúa enviando nuevos lotes sin conocer el resultado del anterior.

Ejemplo: Un posible error es que el cliente envíe una apuesta como string, entonces el servidor lo procesa mal y arroja un error pero el cliente no tiene manera de saber que ocurrió y continúa mandando paquetes. Esto se soluciona gracias al ACK.


##### Observaciones
Durante el desarrollo me estaba fallando el test porque no se había bajado el servicio. Asegurarse con "make down" y luego "make test"



### Ejercicio 7
Se modificó el servidor para procesar los clientes concurrentemente mediante un thread por conexión. De esta forma, mientras un cliente espera operaciones de red, otros clientes pueden ser atendidos.

El servidor utiliza una Condition para implementar el mecanismo de quorum. Cada cliente notifica cuando terminó de enviar sus apuestas y queda esperando hasta que se haya alcanzado como mínimo la cantidad de agencias indicada por AGENCY_QUORUM_MIN. Los clientes se agrupan en rondas independientes: al alcanzar el quorum se calcula un resultado por agencia y se libera ese grupo, sin reutilizar el quorum ni las apuestas de rondas anteriores.

La Condition permite que los threads que todavía no alcanzaron el quorum queden bloqueados sin consumir CPU y sean despertados cuando otra agencia finaliza.

Además, como todas las agencias utilizan el mismo archivo /data/bets.csv, se utiliza un Lock para sincronizar el acceso al archivo y evitar escrituras concurrentes inconsistentes. El resultado se calcula sobre las apuestas de la ronda reunidas en memoria, evitando mezclar sorteos anteriores. Cada cliente recibe únicamente los ganadores de su propia agencia; no se realiza broadcast.



##### Observaciones
- Hasta ahora se aceptaba una conexión y el server quedaba bloqueado, no permitiendo multiples clientes realmente, solo atendiendo uno por vez. En los logs se podía ver que primero estan los logs de client_0, se hace todo el procesamiento, se envian los ganadores y luego se hacía para client_1. Aunque puede ser dificil de rastrear ya que los logs pueden imprimirse en otro orden
- El threading no escala bien en CPython (interprete de python) debido al GIL. CPython si tiene multi threading, es GIL quien lo frena. Es muy útil para operaciones I/O-bound (como esperar red, archivos, sockets, etc). De lo contrario para operaciones pesadas es lo mismo que tener un hilo y no existe paralelismo real. 
- Otra forma de implementar la solución era guardar un diccionario de Lottery de forma que cada agencia tenga un archivo de apuestas separado, pero se evita esto ya que el objetivo es intentar sincronizar el acceso y simular problemas concurrencia, que mediante esta alternativa no ocurrirían o sería mas dificil que ocurran. 



### Ejercicio 8
Se implementó el cierre graceful del cliente y servidor ante la recepción de `SIGTERM`. El objetivo es que los recursos utilizados, como sockets, archivos y threads, sean liberados correctamente antes de finalizar.

En el cliente se utiliza context para propagar la señal de cancelación. Al recibir `SIGTERM`, el contexto queda cancelado y el cliente deja de enviar nuevas apuestas. Una goroutine cierra inmediatamente el socket si hay una operación de red bloqueada, evitando que la finalización dependa de un sleep o timeout artificial.

En el servidor se utiliza `threading.Event` para representar el estado de shutdown. Al recibir `SIGTERM`, se activa el evento, se despiertan los threads que estén esperando el quorum y se cierran tanto el socket de escucha como los sockets de los clientes para desbloquear posibles operaciones de accept() y recv().

Finalmente, el servidor espera mediante a que terminen los threads de los clientes antes de finalizar el proceso.

El tiempo -t utilizado por docker compose down establece cuánto tiempo Docker espera a que los procesos terminen correctamente antes de forzar su finalización mediante SIGKILL.

#### Observaciones
- Se puede usar mutex para diseñar un estado compartido y preguntar si el programa sigue ejecutandose o no, pero se decidió ir por la implementación de context
- SIGTERM sirve para avisar que se intentará cerrar el programa y la idea es hacerlo lo mas rapido posible, mientras que SIGKILL directamente no espera y hace el cierre a la fuerza
- Se puede simular una señal de SIGTERM (más allá de los tests) mediante el pasaje de un archivo muy grande al cliente (por ejemplo `input-1.csv`) ejecutando `make up` seguido de `make logs` y luego en otra consola ejecutar `make down`. Se podrá ver una salida similar a la siguiente, registrandose correctamente que se hizo una salida _grateful_

```
...
client_0  | 2026/09/08 21:14:45 INFO action=sigterm-received result=in-progress
client_1  | 2026/09/08 21:14:45 INFO action=sigterm-received result=in-progress
client_1  | 2026/09/08 21:14:45 INFO action=send-bets result=in-progress info="shutdown solicitado, se corta el envío"
client_1  | 2026/09/08 21:14:45 INFO action=process-input-file-from-server result=success info="cierre graceful por SIGTERM"
...
client_1 exited with code 0
client_0  | 2026/09/08 21:14:49 INFO action=process-input-file-from-server result=success info="cierre graceful por SIGTERM"
client_0 exited with code 0
server    | 2026/09/08 21:14:49 INFO action=sigterm-received result=in-progress 
server    | 2026/09/08 21:14:49 INFO action=sigterm-received result=in-progress 
server exited with code 0
```









