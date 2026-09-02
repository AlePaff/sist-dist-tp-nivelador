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


