# Warehouse API
A (WIP) RESTful API for managing warehouse inventory. Supports tracking items, stock levels across multiple storage locations, and stock movement history. Built for internal use by warehouse staff and managers, using role-based access to control what each user can do.

## Core Features
- User registration and login with JWT auth
- Create and manage items (SKU, name, description, quantity)
- Create and manage locations (bays, shelves, etc)
- Move stock between locations
- View current stock levels per item and per location
- Role-based access: `worker` and `manager` roles with different permissions

## Tech Stack
- **Go:** HTTP server (`net/http`)
- **PostgreSQL:** persistent storage
- **goose:** database migrations
- **sqlc:** type-safe SQL query generation
- **Docker:** containerize the app and database together (not yet implemented)

## API Endpoints


| METHOD | ENDPOINT               | DESCRIPTION                                           |
| ------ | ---------------------- | ----------------------------------------------------- |
| POST   | `/register`            | Create user account                                   |
| POST   | `/login`               | Returns JWT                                           |
|        |                        |                                                       |
| GET    | `/items`               | List all items                                        |
| POST   | `/items`               | Create item (manager only)                            |
| GET    | `/items/{id}`          | Get single item                                       |
| PATCH  | `/items/{id}`          | Update item (manager only)                            |
| DELETE | `/items/{id}`          | Delete item (manager only)                            |
|        |                        |                                                       |
| GET    | `/locations`           | List all locations                                    |
| POST   | `/locations`           | Create location (manager only)                        |
| GET    | `/locations/{id}`      | Get a single location                                 |
| PATCH  | `/locations/{id}`      | Update location (manager only)                        |
| DELETE | `/locations/{id}`      | Delete location (manager only)                        |
|        |                        |                                                       |
| GET    | `/stock`               | View all current stock levels                         |
| GET    | `/stock/alerts`        | View low-stock items - quantity < low stock threshold |
| GET    | `/stock/item/{id}`     | Stock levels for one item across locations            |
| GET    | `/stock/location/{id}` | All items from one location                           |
| POST   | `/stock/move`          | Move quantity from one location to another            |
| POST   | `/stock/receive`       | Add incoming stock to a location (manager only)       |
|        |                        |                                                       |
| GET    | `/movements`           | Item movement history (manager only)                  |
| GET    | `/movements/item/{id}` | Movement history for a single item                    |


## Project Structure
```
warehouse-api/
├── main.go               # Server entry point
├── server.go             # Server struct and helpers
├── handlers.go           # HTTP handlers
├── helpers.go            # Helper functions
├── middleware.go         # requireAuth, requireRole
├── types.go              # Custom type definitions
├── internal/
│   └── auth/             # user authentication stuff
│   └── database/         # sqlc generated code
├── sql/
│   ├── queries/          # SQL queries
│   └── schema/           # goose migrations
└── sqlc.yaml
```

## Stretch Goals
- [ ] Low stock alerts - flag items that fall below a configurable threshold
- [ ] Audit log - append-only record of every stock movement (who did it, when, how much)
- [ ] Soft deletes - instead of hard deleting items or locations, mark them inactive
- [ ] Search and filtering on item listings (by name, SKU, low stock status)
- [ ] Receiving endpoint - bulk-add stock from a shipment to a location
- [ ] Admin role - can manage users, reset passwords, assign roles
