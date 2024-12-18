# Dashboard Backend

This project is a backend application for a dashboard app built in Go. It consists of multiple microservices that handle different functionalities, including user management, order processing, notifications, and middleware for third-party API calls.

## Project Structure

```
dashboard-backend/
├── api-gateway/                # API Gateway service
│   ├── config/                 # Configurations for gateway routes
│   ├── handlers/               # Route handlers for requests
│   ├── middlewares/            # Middleware logic for authentication, logging, etc.
│   ├── main.go                 # Entry point for API Gateway
│   └── go.mod                  # Module dependencies
│
├── load-balancer/              # Load balancer logic (optional: NGINX config or custom logic)
│   ├── nginx.conf              # Example: NGINX config
│   └── README.md               # Documentation
│
├── services/                   # Directory for all microservices
│   ├── user-service/           # Microservice for user management
│   │   ├── config/             # Configuration files (MongoDB connection, env vars)
│   │   ├── controllers/        # REST API controllers
│   │   ├── routes/             # Route definitions
│   │   ├── models/             # Data models (MongoDB schemas)
│   │   ├── repositories/       # DB interactions
│   │   ├── services/           # Business logic layer
│   │   ├── events/             # Event listeners for messages (e.g., Kafka, NATS)
│   │   ├── middlewares/        # Authentication, validation middleware
│   │   ├── utils/              # Utility functions (logger, error handling, etc.)
│   │   ├── main.go             # Entry point for User Service
│   │   └── go.mod              # Module dependencies
│   │
│   ├── order-service/          # Microservice for order handling
│   │   ├── (same structure as user-service)
│   │   └── ...
│   │
│   ├── notification-service/   # Microservice for notifications
│   │   ├── (same structure as user-service)
│   │   └── ...
│   │
│   └── sentinel-service/       # Middleware service (fancy name: Sentinel Service)
│       ├── config/             # Configurations for third-party API keys
│       ├── controllers/        # HTTP or gRPC endpoints for communication
│       ├── routes/             # Route definitions
│       ├── services/           # Business logic for third-party integrations
│       ├── validators/         # Custom validators for API payloads
│       ├── middlewares/        # Logging, authentication for external calls
│       ├── utils/              # Utility functions
│       ├── events/             # Event listening logic
│       ├── main.go             # Entry point for Sentinel Service
│       └── go.mod              # Module dependencies
│
├── configs/                    # Global configurations for all services
│   ├── dev.env                 # Development environment variables
│   ├── prod.env                # Production environment variables
│   └── common.yaml             # Shared configurations (e.g., MongoDB URI, Kafka brokers)
│
├── libs/                       # Shared libraries (if applicable)
│   ├── logger/                 # Custom logger implementation
│   ├── event-bus/              # Kafka/NATS wrapper library for event streaming
│   ├── utils/                  # Shared utility functions (e.g., JSON validation)
│   └── auth/                   # Shared authentication logic (JWT or OAuth2)
│
├── deployments/                # Deployment-related files
│   ├── docker-compose.yml      # Docker setup for local development
│   ├── kubernetes/             # K8s manifests for microservices
│   └── README.md               # Deployment documentation
│
└── README.md                   # Project documentation
```

## Microservices

- **User Microservice**: Manages user data with functionalities to create, retrieve, and update users.
- **Orders Microservice**: Handles order data with functionalities to create, retrieve, and update orders.
- **Notifications Microservice**: Manages notifications with functionalities to send and retrieve notifications.
- **Middleware**: Acts as a layer for handling requests and responses to external services.

## Setup Instructions

1. Clone the repository:
   ```
   git clone <repository-url>
   cd dashboard-backend
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Run the microservices:
   - For User Service:
     ```
     go run cmd/user/main.go
     ```
   - For Orders Service:
     ```
     go run cmd/orders/main.go
     ```
   - For Notifications Service:
     ```
     go run cmd/notifications/main.go
     ```
   - For Middleware:
     ```
     go run cmd/middleware/main.go
     ```

## Usage

Each microservice exposes its own set of APIs. Refer to the respective service documentation for details on available endpoints and their usage.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.