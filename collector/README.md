# collector

Samlar MeshCore-noder i Värmland från MQTT (`mqtt.meshat.se`) och publicerar dem som `nodes.json` på grenen `data` i det här repot.
Kartan på varmlandmesh.se läser filen därifrån.

## Så fungerar det

1. Lyssnar på `meshcore/+/+/packets`, `neighbors` och `status`.
2. Avkodar adverts (namn, typ, position) och kontrollerar signaturen.
3. Sparar alla noder och grannlänkar i `/data/state.json`.
4. Var 15:e minut skrivs noderna inom `BBOX` till `nodes.json`.
   Grenen `data` skrivs över med en enda commit, så historiken växer inte och sajten byggs inte om.

Noder syns först när de har skickat en advert med position.
Det kan ta flera timmar efter första start.

## Kör

```sh
cp .env.example .env   # fyll i GITHUB_TOKEN
docker compose up -d
docker compose logs -f
```

Imagen `ghcr.io/stone/varmlandmesh-collector` byggs för amd64 och arm64 (t.ex. Raspberry Pi) av workflowet `Collector` vid push till `main`.
Taggar: `latest` och `sha-<commit>`.
Bygg lokalt i stället med `docker compose up -d --build`.

Token: en fine-grained token med åtkomst till bara `stone/varmlandmesh_web` och behörigheten **Contents: Read and write**.

## Testa lokalt utan GitHub

```sh
STATE_FILE=./state.json OUTPUT_FILE=./nodes.json PUBLISH_INTERVAL=1m go run .
```

## Inställningar

| Variabel | Standard | |
|---|---|---|
| `GITHUB_TOKEN` | | Utan token publiceras inget till GitHub. |
| `GITHUB_REPO` | `stone/varmlandmesh_web` | |
| `GITHUB_BRANCH` | `data` | |
| `PUBLISH_INTERVAL` | `15m` | |
| `BBOX` | `58.8,61.1,11.6,14.5` | Syd, norr, väst, öst. |
| `MAX_AGE` | `2160h` | Noder som inte hörts på 90 dagar tas bort. |
| `OUTPUT_FILE` | | Skriver även `nodes.json` lokalt. |
| `STATE_FILE` | `/data/state.json` | |
| `MQTT_BROKER` | `tcp://mqtt.meshat.se:1883` | Användare och lösenord: `msh`/`msh`. |

## Format

```json
{
  "schema": 1,
  "generated": "2026-11-17T17:00:00Z",
  "bbox": {"s": 58.8, "n": 61.1, "w": 11.6, "e": 14.5},
  "count": 1,
  "nodes": [{"id": "ea3350b136d9869a", "key": "…", "name": "…", "type": "repeater", "lat": 59.38, "lon": 13.5, "last": "…", "advert": "…"}],
  "links": [{"a": "<id>", "b": "<id>", "snr": 7.5, "heard": "…", "count": 3}]
}
```

`generated` uppdateras bara när det har kommit MQTT-data sedan förra publiceringen.
