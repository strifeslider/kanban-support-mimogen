# KanBan Support Service - MimoGen

Support ticket management microservice for KanBan SaaS platform.

## Features

- Ticket CRUD with status tracking
- Message system within tickets
- Agent management
- Status audit logging
- Statistics dashboard

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/tickets | Create ticket |
| GET | /api/v1/tickets | List tickets |
| GET | /api/v1/tickets/:id | Get ticket |
| POST | /api/v1/tickets/:id/messages | Add message |
| GET | /api/v1/agents | List agents |
| GET | /api/v1/stats | Dashboard stats |

## Tech Stack

- Go 1.22
- Chi Router
- PostgreSQL

---

<details>
<summary><strong>IMPORTANT NOTICE</strong></summary>

<br>

**This repository was entirely generated using [MiMoCode](https://github.com/xiaomi/mimocode) - an AI-powered coding assistant by Xiaomi.**

All code, tests, documentation, and infrastructure configuration in this repository were created through AI-assisted development with MiMoCode.

---

**Этот репозиторий был полностью сгенерирован с помощью [MiMoCode](https://github.com/xiaomi/mimocode) - AI-ассистента для программирования от Xiaomi.**

Весь код, тесты, документация и инфраструктурная конфигурация в этом репозитории были созданы с помощью AI-ассистированной разработки MiMoCode.

</details>
