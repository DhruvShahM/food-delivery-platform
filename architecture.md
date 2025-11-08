[13 tools called]

Based on my exploration of your food delivery platform codebase, I can provide you with comprehensive architecture diagrams. This is a well-structured microservices application with event-driven architecture. Let me create several architectural views:

## **System Architecture Overview**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Food Delivery Platform                       │
│                    ┌─────────────────────────────────────┐      │
│                    │         API Gateway (Kong)         │      │
│                    │         Port: 8000                 │      │
│                    └─────────────────┬───────────────────┘      │
│                                      │                          │
│         ┌─────────────────────┬──────┴─────┬────────────────────┐
│         │                     │            │                    │
│ ┌───────▼──────┐    ┌─────────▼────────┐   │   ┌─────────▼────────┐ │
│ │ Auth Service │    │ Order Service    │   │   │ Payment Service  │ │
│ │ Port: 8081   │    │ Port: 8082       │   │   │ Port: 8083       │ │
│ │ gRPC: 50051  │    │ gRPC: 50052      │   │   │ gRPC: 50053      │ │
│ └──────────────┘    └─────────┬────────┘   │   └─────────────────┘ │
│                               │            │                        │
│         ┌─────────────────────▼────────────┼────────────────────┐ │
│         │                     │            │                    │ │
│ ┌───────▼──────┐    ┌─────────▼────────┐   │   ┌─────────▼────────┐ │
│ │Restaurant Svc│    │ Delivery Service│   │   │ Tracking Service│ │
│ │ Port: 8085   │    │ Port: 8084       │   │   │ Port: 8086       │ │
│ │ gRPC: 50055  │    │ gRPC: 50054      │   │   │ gRPC: 50056      │ │
│ └──────────────┘    └─────────────────┘   │   └─────────────────┘ │
│                                           │                        │
└───────────────────────────────────────────┼────────────────────────┘
                                            │
                       ┌─────────▼────────┐
                       │   Kafka Broker  │
                       │   Port: 9092     │
                       └─────────────────┘
```

## **Data Flow Architecture**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │────│ API Gateway │────│Auth Service │
│             │    │   (Kong)    │    │  JWT Gen    │
└─────────────┘    └─────────────┘    └──────┬──────┘
                                               │
                                               ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│Restaurant   │    │ Order       │────│  Kafka      │────│ Delivery    │
│Service      │    │ Service     │    │  Events     │    │ Service     │
│(Menu Data)  │    │ (Create)    │    │             │    │ (Assign)    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       ▲                     │                           │
       │                     ▼                           ▼
       │           ┌─────────────┐             ┌─────────────┐
       │           │ Payment     │             │ Tracking    │
       │           │ Service     │             │ Service     │
       │           │ (Process)   │             │ (Location)  │
       │           └─────────────┘             └─────────────┘
       │                     │                           │
       └─────────────────────┼───────────────────────────┘
                             ▼
                   ┌─────────────┐
                   │ PostgreSQL  │
                   │   Database  │
                   └─────────────┘
```

## **Technology Stack Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                 Application Layer                           │
├─────────────────────────────────────────────────────────────┤
│  • Go 1.21 + Gin Framework                                  │
│  • gRPC for Service Communication                           │
│  • REST API via Kong Gateway                                │
│  • JWT Authentication                                       │
├─────────────────────────────────────────────────────────────┤
│                 Data Layer                                  │
├─────────────────────────────────────────────────────────────┤
│  • PostgreSQL (Primary Database)                            │
│  • Redis (Caching & Sessions)                               │
│  • Kafka (Event Streaming)                                  │
├─────────────────────────────────────────────────────────────┤
│                 Infrastructure Layer                        │
├─────────────────────────────────────────────────────────────┤
│  • Kong API Gateway (8000)                                  │
│  • Docker Compose (Container Orchestration)                 │
│  • Prometheus + Grafana (Monitoring)                        │
│  • Jaeger (Distributed Tracing)                             │
├─────────────────────────────────────────────────────────────┤
│                 Cross-Cutting Concerns                      │
├─────────────────────────────────────────────────────────────┤
│  • Circuit Breakers (Resilience4j)                          │
│  • Rate Limiting                                            │
│  • CORS Handling                                            │
│  • Structured Logging (Zap)                                 │
│  • OpenTelemetry Tracing                                    │
└─────────────────────────────────────────────────────────────┘
```

## **Event-Driven Architecture**

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Order Service │────▶│   Kafka Topic   │────▶│ Delivery Service│
│                 │     │ "order-events"  │     │                 │
│  1. Place Order │     │                 │     │ 4. Assign Agent │
│  2. Publish     │     │  OrderEvent{    │     │                 │
│     Event       │     │   order_id,     │     └─────────────────┘
└─────────────────┘     │   status,       │             │
                        │   amount        │             ▼
                        │ }               │     ┌─────────────────┐
                        └─────────────────┘     │ Tracking Service│
                                                │                 │
                        ┌─────────────────┐     │ 5. Track        │
                        │   Kafka Topic   │     │    Location     │
                        │ "location-updates"    │                 │
                        └─────────────────┘     └─────────────────┘
```

## **Service Communication Patterns**

```
HTTP/REST (External APIs)
    ↓
┌─────────────────────────────────────┐
│         Kong API Gateway            │
│  • JWT Authentication               │
│  • Rate Limiting                    │
│  • Request Routing                  │
│  • CORS Handling                    │
└─────────────────────────────────────┘
    ↓
gRPC (Internal Service-to-Service)
    ↓
┌─────────────────────────────────────┐
│         Microservices               │
│  • Auth Service (Port 50051)        │
│  • Order Service (Port 50052)       │
│  • Payment Service (Port 50053)     │
│  • Delivery Service (Port 50054)    │
│  • Restaurant Service (Port 50055)  │
│  • Tracking Service (Port 50056)    │
└─────────────────────────────────────┘
    ↓
Kafka (Event-Driven Communication)
    ↓
┌─────────────────────────────────────┐
│         Event Topics                │
│  • order-events                     │
│  • order-created                    │
│  • location-updates                 │
└─────────────────────────────────────┘
```

## **Deployment Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose Network                   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │  Services   │ │  Databases  │ │ Monitoring  │           │
│  │             │ │             │ │             │           │
│  │ • Auth      │ │ • Postgres  │ │ • Prometheus│           │
│  │ • Order     │ │ • Redis     │ │ • Grafana   │           │
│  │ • Payment   │ │ • Kafka     │ │ • Jaeger    │           │
│  │ • Delivery  │ │ • Zookeeper │ │             │           │
│  │ • Restaurant│ │             │ └─────────────┘           │
│  │ • Tracking  │ └─────────────┘                           │
│  └─────────────┘                                           │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure                           │
├─────────────────────────────────────────────────────────────┤
│  • Kong Gateway (8000)                                     │
│  • Swagger UI (8081)                                       │
│  • API Documentation                                       │
└─────────────────────────────────────────────────────────────┘
```

## **Security Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Layers                          │
├─────────────────────────────────────────────────────────────┤
│  Client Request                                            │
│  ↓                                                         │
├─────────────────────────────────────────────────────────────┤
│  Kong API Gateway                                          │
│  • JWT Token Validation                                    │
│  • Rate Limiting                                           │
│  • Request Filtering                                       │
├─────────────────────────────────────────────────────────────┤
│  Service Layer                                             │
│  • Authentication Middleware                               │
│  • Authorization Checks                                    │
│  • Input Validation                                        │
├─────────────────────────────────────────────────────────────┤
│  Data Layer                                                │
│  • SQL Injection Prevention                                │
│  • Secure Database Connections                             │
│  • Encrypted JWT Tokens                                    │
└─────────────────────────────────────────────────────────────┘
```

This architecture demonstrates a production-ready microservices system with:
- **Scalability** through independent services
- **Resilience** with circuit breakers and monitoring
- **Observability** via Prometheus/Grafana/Jaeger
- **Event-driven** communication with Kafka
- **API Gateway** pattern with Kong
- **Security** through JWT and middleware

The system supports the complete food delivery workflow from user registration through order placement, payment processing, delivery assignment, and real-time tracking.