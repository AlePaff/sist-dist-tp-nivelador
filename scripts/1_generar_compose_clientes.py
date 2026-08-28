import sys

def generate_compose(num_clients):
    compose = """
services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678

"""

    for i in range(num_clients):
        compose += f"""  client_{i}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{i}
    depends_on:
      - server
    environment:
      - AGENCY_ID={i}
      - SERVER_HOST=server
      - SERVER_PORT=5678

"""

    return compose


# accede a los argumentos de la linea de comandos
if len(sys.argv) != 2:
    print("Uso: python3 generate_compose.py <cantidad_de_clientes>")
    sys.exit(1)

# algunas validaciones comunes
try:
    num_clients = int(sys.argv[1])
except ValueError:
    print("La cantidad de clientes debe ser un número entero.")
    sys.exit(1)

if num_clients < 1:
    print("La cantidad de clientes debe ser mayor o igual a 1.")
    sys.exit(1)

compose = generate_compose(num_clients)

# si el archivo ya existe, se sobrescribe
with open("docker-compose.yaml", "w") as file:
    file.write(compose)

print(f"Se generó docker-compose.yaml con {num_clients} clientes.")
