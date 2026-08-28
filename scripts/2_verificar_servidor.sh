#!/bin/bash

NETWORK=$(docker inspect server \
  --format '{{range $key, $value := .NetworkSettings.Networks}}{{$key}}{{end}}')
# formatea la salida de docker inspect y toma el nombre de la red interna que tiene el servidor. Por ejemplo "sist-dist-tp-nivelador_default"
# ejemplo de salida de "docker inspect server"
        # "NetworkSettings": {
        #     "Ports": {}
        #     "Networks": {
        #         "sist-dist-tp-nivelador_default": {}          <--- ACA
        #     }
        # }
# echo "Network: $NETWORK"

# crea un contenedor temporal alpine, y se conecta a la red interna del servidor para enviarle un mensaje de prueba
# el nombre de la red interna se obtiene del comando anterior. El contenedor temporal se elimina al finalizar (--rm)
docker run --rm \
  --network "$NETWORK" \
  alpine:latest \
  sh -c 'echo "Hello World" | nc server 5678'

