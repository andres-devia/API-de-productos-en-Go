# Item Comparison API

Una API backend lista para producción construida con Go que proporciona información de productos para una funcionalidad de comparación de ítems. El proyecto sigue los principios de la Clean Architecture (Arquitectura Hexagonal) y los patrones de diseño SOLID.

## Arquitectura

Este proyecto implementa Clean Architecture / Hexagonal Architecture con una clara separación de responsabilidades:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer (Chi)                     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │   Handlers   │   │  Middleware  │   │   Router     │ │
│  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘ │
└─────────┼──────────────────┼──────────────────┼─────────┘
          │                  │                  │
┌─────────┼──────────────────┼──────────────────┼───────┐
│         │                  │                  │       │
│  ┌──────▼───────┐          │          ┌───────▼─────┐ │
│  │   Services   │◄─────────┼──────────┤  Domain     │ │
│  │  (Business   │          │          │  Models     │ │
│  │   Logic)     │          │          │             │ │
│  └──────┬───────┘          │          └─────────────┘ │
│         │                  │                          │
│  ┌──────▼───────┐          │          ┌─────────────┐ │
│  │ Repositories │          │          │   Errors    │ │
│  │  (Data Access│          │          │             │ │
│  │   Layer)     │          │          └─────────────┘ │
│  └──────┬───────┘          │                          │
└─────────┼──────────────────┼──────────────────────────┘
          │                  │
┌─────────▼──────────────────▼──────────────────────────┐
│              SQLite Database                          │
└───────────────────────────────────────────────────────┘
```

### Responsabilidades de las capas

1. **Handlers** (`internal/handlers/`): Manejo de solicitudes/respuestas HTTP, validación de entrada y serialización JSON
2. **Services** (`internal/services/`): Lógica de negocio y orquestación entre capas
3. **Repositories** (`internal/repositories/`): Abstracción de acceso a datos con interfaces
4. **Models** (`internal/models/`): Entidades de dominio y objetos de valor (Item, CompareRequest, CompareResponse)
5. **Errors** (`internal/errors/`): Manejo de errores a nivel de dominio con códigos HTTP apropiados
6. **Middleware** (`internal/middleware/`): Cross-cutting concerns (CORS, security headers, rate limiting)
7. **Server** (`internal/server/`): Configuración del servidor HTTP, router y gestión del ciclo de vida

### Principios de diseño aplicados

- **SOLID Principles**: Todas las capas utilizan interfaces, inyección de dependencias y responsabilidad única
- **Dependency Inversion**: Los módulos de alto nivel dependen de abstracciones, no de implementaciones concretas
- **Separation of Concerns**: Cada capa tiene una responsabilidad única y bien definida
- **Clean Code**: Código completamente comentado, legible y con estilo de producción
- **Graceful Shutdown**: Manejo seguro del cierre del servidor con timeouts configurados

## Estructura del proyecto

```
.
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada de la aplicación
├── internal/
│   ├── handlers/                # HTTP handlers
│   │   ├── item_handler.go      # Handlers para endpoints de items
│   │   └── item_handler_test.go # Tests de handlers
│   ├── services/                # Capa de lógica de negocio
│   │   ├── item_service.go      # Interfaz del servicio
│   │   ├── item_service_impl.go # Implementación del servicio
│   │   └── item_service_test.go # Tests del servicio
│   ├── repositories/            # Capa de acceso a datos
│   │   ├── item_repository.go   # Interfaz del repositorio
│   │   ├── error.go             # Errores específicos del repositorio
│   │   └── sqlite/              # Implementación SQLite
│   │       ├── sqlite_repository.go    # Repositorio SQLite
│   │       ├── sqlite_item_queries.go # Consultas SQL
│   │       └── sqlite_item_seed.go    # Datos iniciales (seed)
│   ├── models/                  # Entidades de dominio
│   │   └── item.go              # Modelos Item, CompareRequest, CompareResponse
│   ├── errors/                  # Manejo de errores
│   │   └── errors.go            # Errores de dominio tipados
│   ├── middleware/              # Middleware HTTP
│   │   ├── cors.go              # Configuración CORS
│   │   ├── security.go          # Headers de seguridad
│   │   └── ratelimit.go        # Rate limiting por IP
│   └── server/                  # Configuración del servidor
│       ├── server.go            # Inicialización y ciclo de vida del servidor
│       ├── router.go            # Configuración de rutas y middlewares
│       ├── config_struct.go     # Estructura de configuración
│       └── server_struct.go     # Estructura del servidor
├── docs/
│   └── swagger.yaml             # Documentación OpenAPI/Swagger
├── go.mod                       # Dependencias del proyecto
├── go.sum                       # Checksums de dependencias
└── README.md                    # Este archivo
```

## Inicialización del proyecto

### Prerequisitos

- Go 1.24.0 o superior
- SQLite3

### Instalación

1. Clona el repositorio:
```bash
git clone <repository-url>
cd f876e99f-a0f2-4e4d-89d2-642de438fb76-v2
```

2. Instala las dependencias:
```bash
go mod download
```

### Ejecutar el proyecto

Inicia el servidor con la configuración por defecto (puerto 8080, archivo de base de datos `items.db`):

```bash
go run cmd/api/main.go
```

O con una configuración personalizada:

```bash
go run cmd/api/main.go -port 3000 -db custom.db
```

El servidor realizará automáticamente:
- Inicialización de la base de datos SQLite
- Creación de la tabla de items si no existe
- Carga de datos de ejemplo (seed) con 5 items
- Inicio del servidor HTTP en el puerto especificado
- Configuración de todos los middlewares (CORS, seguridad, rate limiting)

### Compilar

Compilar la aplicación:

```bash
go build -o item-comparison-api cmd/api/main.go
```

Ejecutar el binario:

```bash
./item-comparison-api
```

En Windows:

```bash
item-comparison-api.exe
```

## Base de datos

La base de datos SQLite se inicializa automáticamente con 5 ítems de ejemplo al iniciar el servidor. Los datos incluyen información completa de productos con especificaciones, precios y ratings.

Para reiniciar la base de datos: elimina el archivo `.db` y ejecuta el proyecto nuevamente.

## API Endpoints

### Base URL

```
http://localhost:8080/api/v1
```

### Endpoints disponibles

#### 1. Obtener todos los items

**GET** `/api/v1/items`

Obtiene una lista de todos los items disponibles en el sistema.

**Respuesta exitosa (200):**
```json
[
  {
    "id": 1,
    "name": "MacBook Pro 16\"",
    "image_url": "https://example.com/images/macbook-pro.jpg",
    "description": "Powerful laptop for professionals with M2 Pro chip",
    "price": 2499.99,
    "rating": 4.8,
    "specifications": {
      "processor": "Apple M2 Pro",
      "memory": "16GB",
      "storage": "512GB SSD",
      "display": "16.2-inch Liquid Retina XDR"
    }
  }
]
```

**Códigos de respuesta:**
- `200`: Éxito
- `429`: Rate limit excedido
- `500`: Error interno del servidor

#### 2. Obtener item por ID

**GET** `/api/v1/items/{id}`

Obtiene un item específico por su identificador único.

**Parámetros:**
- `id` (path, requerido): ID numérico del item

**Ejemplo:**
```bash
GET /api/v1/items/1
```

**Respuesta exitosa (200):**
```json
{
  "id": 1,
  "name": "MacBook Pro 16\"",
  "image_url": "https://example.com/images/macbook-pro.jpg",
  "description": "Powerful laptop for professionals with M2 Pro chip",
  "price": 2499.99,
  "rating": 4.8,
  "specifications": {
    "processor": "Apple M2 Pro",
    "memory": "16GB",
    "storage": "512GB SSD",
    "display": "16.2-inch Liquid Retina XDR"
  }
}
```

**Códigos de respuesta:**
- `200`: Éxito
- `400`: ID inválido (formato incorrecto)
- `404`: Item no encontrado
- `429`: Rate limit excedido
- `500`: Error interno del servidor

#### 3. Comparar items

**POST** `/api/v1/items/compare`

Compara múltiples items y devuelve información detallada de comparación incluyendo:
- Rango de precios (mínimo/máximo)
- Rango de ratings (mínimo/máximo)
- Especificaciones comunes a todos los items
- Especificaciones únicas por item

**Cuerpo de la petición:**
```json
{
  "item_ids": [1, 2, 3]
}
```

**Validaciones:**
- Mínimo 2 items requeridos
- Máximo 10 items permitidos
- Todos los IDs deben existir en la base de datos

**Ejemplo de petición:**
```bash
curl -X POST http://localhost:8080/api/v1/items/compare \
  -H "Content-Type: application/json" \
  -d '{"item_ids": [1, 2, 3]}'
```

**Respuesta exitosa (200):**
```json
{
  "items": [
    {
      "id": 1,
      "name": "MacBook Pro 16\"",
      "image_url": "https://example.com/images/macbook-pro.jpg",
      "description": "Powerful laptop for professionals with M2 Pro chip",
      "price": 2499.99,
      "rating": 4.8,
      "specifications": {
        "processor": "Apple M2 Pro",
        "memory": "16GB",
        "storage": "512GB SSD",
        "display": "16.2-inch Liquid Retina XDR"
      }
    }
  ],
  "comparison": {
    "price_range": {
      "min": 1499.99,
      "max": 2499.99
    },
    "rating_range": {
      "min": 4.4,
      "max": 4.8
    },
    "common_specs": ["processor", "memory", "storage"],
    "unique_specs": {
      "1": ["touchscreen"],
      "2": ["durability"]
    }
  }
}
```

**Códigos de respuesta:**
- `200`: Comparación exitosa
- `400`: Cuerpo de petición inválido
- `404`: Uno o más items no encontrados
- `422`: Error de validación de negocio (menos de 2 IDs, más de 10 IDs)
- `429`: Rate limit excedido
- `500`: Error interno del servidor

## Testing

Ejecutar todos los tests:

```bash
go test ./...
```

Ejecutar las pruebas con cobertura:

```bash
go test -cover ./...
```

Ejecutar las pruebas con salida detallada:

```bash
go test -v ./...
```

Ejecutar tests de un paquete específico:

```bash
go test ./internal/services/...
go test ./internal/handlers/...
```

## Funcionalidades de seguridad

### Middleware implementados

El proyecto incluye los siguientes middlewares aplicados globalmente:

1. **Security Headers** (`internal/middleware/security.go`):
   - `X-Frame-Options: DENY` - Previene clickjacking
   - `X-Content-Type-Options: nosniff` - Previene MIME type sniffing
   - `X-XSS-Protection: 1; mode=block` - Protección XSS
   - `Content-Security-Policy` - Política de seguridad de contenido
   - `Referrer-Policy` - Control de información de referrer

2. **CORS** (`internal/middleware/cors.go`):
   - Habilita solicitudes cross-origin con headers configurables
   - Permite métodos GET, POST, OPTIONS
   - Headers permitidos: Content-Type, Authorization

3. **Rate Limiting** (`internal/middleware/ratelimit.go`):
   - 100 solicitudes por minuto por dirección IP (configurable)
   - Protección contra abuso y ataques de denegación de servicio (DoS)
   - Respuesta `429 Too Many Requests` cuando se excede el límite

4. **Request ID** (Chi middleware):
   - Asigna un ID único por petición para trazabilidad y debugging

5. **Logger** (Chi middleware):
   - Registra cada petición HTTP con detalles de ruta, método, latencia y código de respuesta

6. **Recoverer** (Chi middleware):
   - Captura panics y previene que el servidor colapse
   - Devuelve respuestas de error apropiadas

### Buenas prácticas de seguridad

- **Validación de inputs**: Todos los endpoints validan los datos de entrada
- **Consultas SQL parametrizadas**: Prevención de inyección SQL
- **Límite de tamaño de body**: Máximo 1MB para peticiones POST
- **Manejo estandarizado de errores**: Los errores no exponen detalles internos de implementación
- **Códigos HTTP apropiados**: Uso correcto de códigos de estado HTTP
- **Graceful Shutdown**: Cierre seguro del servidor con timeouts configurados

## Documentación de la API (Swagger)

Toda la documentación completa de endpoints, parámetros, ejemplos y esquemas se encuentra en formato OpenAPI 3.0:

```
docs/swagger.yaml
```

Esta documentación incluye:
- Descripción detallada de cada endpoint
- Esquemas de request y response
- Ejemplos de uso
- Códigos de error y sus significados
- Validaciones y restricciones

Puedes visualizar la documentación usando herramientas como:
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [ReDoc](https://github.com/Redocly/redoc)
- [Postman](https://www.postman.com/) (importar el archivo YAML)

## Configuración

La aplicación soporta parámetros de línea de comandos:

- `-port`: Puerto del servidor (por defecto: `8080`)
- `-db`: Ruta del archivo de base de datos SQLite (por defecto: `items.db`)

**Ejemplo:**
```bash
go run cmd/api/main.go -port 3000 -db /ruta/a/database.db
```

### Timeouts del servidor

El servidor HTTP está configurado con los siguientes timeouts:
- **ReadTimeout**: 15 segundos
- **WriteTimeout**: 15 segundos
- **IdleTimeout**: 60 segundos
- **ShutdownTimeout**: 10 segundos (para graceful shutdown)

## Consideraciones para producción

Para el despliegue en producción, se recomienda considerar:

1. **Base de datos**: Reemplazar SQLite con PostgreSQL/MySQL para mejor concurrencia y escalabilidad
2. **Rate Limiting**: Usar rate limiting basado en Redis para sistemas distribuidos
3. **Logging**: Integrar logging estructurado (ej: zap, logrus) con niveles y rotación
4. **Monitoreo**: Agregar métricas (Prometheus) y endpoints de health check
5. **Configuración**: Usar variables de entorno o archivos de configuración (viper)
6. **HTTPS**: Habilitar certificados TLS/SSL
7. **CORS**: Restringir orígenes permitidos en producción
8. **Migraciones de base de datos**: Usar una herramienta de migraciones (ej: golang-migrate)
9. **Autenticación/Autorización**: Implementar JWT o OAuth2 si es necesario
10. **Containerización**: Dockerizar la aplicación para despliegue consistente
11. **CI/CD**: Configurar pipelines de integración y despliegue continuo

## Características técnicas

### Gestión del ciclo de vida del servidor

- **Graceful Shutdown**: El servidor maneja señales SIGINT y SIGTERM para un cierre seguro
- **Cierre de recursos**: Cierra correctamente las conexiones de base de datos al detenerse
- **Timeouts configurados**: Previene conexiones colgadas

### Manejo de errores

- **Errores tipados**: Sistema de errores de dominio con códigos HTTP apropiados
- **Respuestas consistentes**: Formato estándar de error en todas las respuestas
- **Logging de errores**: Registro apropiado sin exponer información sensible

### Arquitectura y diseño

- **Inyección de dependencias**: Todas las dependencias se inyectan mediante constructores
- **Interfaces**: Abstracciones claras entre capas para facilitar testing y mantenimiento
- **Separación de responsabilidades**: Cada componente tiene una responsabilidad única
- **Testabilidad**: Arquitectura diseñada para facilitar pruebas unitarias e integración

## Licencia

Este proyecto es parte de la prueba técnica.

## Autor

Jorge Andres Cardeño Devia

---
