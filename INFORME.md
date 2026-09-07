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


