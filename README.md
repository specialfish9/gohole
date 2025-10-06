# Gohole

DNS blocker

<img src="frontend/public/gohole.png" alt="screenshot" width="300"/>

## Screenshot
<img src="assets/screen1.png" alt="screenshot" width="600"/>

## Quick start

Database:

```bash
docker compose -f compose.dev.yaml up -d
```

Backend

```bash
cd backend
make setup
make run
```

Frontend

```bash
cd frontend
npm i
npm run dev
```

## Docker run

```bash
docker compose up
```
