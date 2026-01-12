# GitHub Copilot Instructions

These instructions define **how GitHub Copilot (IDE + CLI)** should assist during development of this project.

The goal is **senior-level, production-grade code** with strong reasoning, fast feedback, and minimal complexity. Copilot is an **assistant**, not an author.

---

## Core Operating Principles

Copilot must:

* Be **concise by default**
* Prefer **clarity over cleverness**
* Optimize for **maintainability and testability**
* Preserve **explicit domain intent** at all times
* Encourage **small, reversible changes**

If a request is ambiguous, Copilot should:

* Ask **one clarifying question**, or
* Present **2–3 options with brief trade-offs**

---

## Mandatory Conciseness Rules

Every Copilot response must follow this structure:

1. **One-sentence summary** of what it is about to do
2. **Short reasoning** (3–5 bullet points max)
3. **Code / command output** (only what is necessary)

No long explanations. No tutorials. No repetition.

---

## Technology Constraints (Non-Negotiable)

### Backend

* Language: **Go 1.21+**
* Style: idiomatic Go, explicit error handling
* Logging: **`log/slog`** with JSON output
* Tracing: **OpenTelemetry** (free, vendor-neutral)
* No hidden magic, no heavy frameworks

### Frontend

* **HTMX + server-rendered HTML**
* Minimal JavaScript (only where unavoidable)
* Fast first paint, simple pages

### Architecture

* **Domain-Driven Design (DDD)**
* **Hexagonal (Ports & Adapters) Architecture**
* Domain must not depend on infrastructure

---

## Architectural Rules

Copilot must enforce:

* Clear separation of: Domain, Application, Infrastructure, Interfaces
* Domain contains: Aggregates, Entities, Value Objects, Domain Events
* **No business logic in handlers, controllers, or templates**
* **No implementation leakage** — consumers depend on interfaces, not concrete types

If a rule is violated, Copilot must **flag it explicitly**.

---

## Code Design Principles

### Encapsulation

Hide implementation details. Expose behaviour, not data.

```go
// ✅ DO: Encapsulate state, expose behaviour
type Property struct {
    id       PropertyID
    address  Address
    status   ListingStatus
}

func (p *Property) Publish() error {
    if p.status != ListingStatusDraft {
        return ErrCannotPublish
    }
    p.status = ListingStatusPublished
    return nil
}

func (p *Property) Status() ListingStatus {
    return p.status
}
```

```go
// ❌ DON'T: Expose internal state directly
type Property struct {
    ID      string
    Address string
    Status  string // Anyone can mutate this
}
```

### Dependency Injection

Depend on interfaces. Inject dependencies at construction.

```go
// ✅ DO: Accept interface, inject at construction
type PropertyService struct {
    repo   PropertyRepository
    events EventPublisher
}

func NewPropertyService(repo PropertyRepository, events EventPublisher) *PropertyService {
    return &PropertyService{repo: repo, events: events}
}
```

```go
// ❌ DON'T: Create dependencies internally
type PropertyService struct{}

func (s *PropertyService) Create(p Property) error {
    repo := NewPostgresRepository() // Hardcoded dependency
    return repo.Save(p)
}
```

### Functional Patterns

Prefer pure functions where possible. Make side effects explicit.

```go
// ✅ DO: Pure function, no side effects
func CalculateCommission(price Money, rate Percentage) Money {
    return price.Multiply(rate.AsDecimal())
}
```

```go
// ❌ DON'T: Hidden side effects
func CalculateCommission(price Money, rate Percentage) Money {
    log.Printf("Calculating commission...") // Hidden side effect
    db.IncrementCalculationCount()          // Hidden side effect
    return price.Multiply(rate.AsDecimal())
}
```

### Decorator Pattern

Extend behaviour without modifying existing code.

```go
// ✅ DO: Decorate to add cross-cutting concerns
type PropertyRepository interface {
    Save(ctx context.Context, p Property) error
    FindByID(ctx context.Context, id PropertyID) (Property, error)
}

// Logging decorator
type LoggingPropertyRepository struct {
    next   PropertyRepository
    logger *slog.Logger
}

func (r *LoggingPropertyRepository) Save(ctx context.Context, p Property) error {
    r.logger.InfoContext(ctx, "saving property", slog.String("property_id", p.id.String()))
    err := r.next.Save(ctx, p)
    if err != nil {
        r.logger.ErrorContext(ctx, "failed to save property", slog.String("property_id", p.id.String()), slog.Any("error", err))
    }
    return err
}

// Tracing decorator
type TracingPropertyRepository struct {
    next   PropertyRepository
    tracer trace.Tracer
}

func (r *TracingPropertyRepository) Save(ctx context.Context, p Property) error {
    ctx, span := r.tracer.Start(ctx, "PropertyRepository.Save")
    defer span.End()
    span.SetAttributes(attribute.String("property.id", p.id.String()))
    return r.next.Save(ctx, p)
}

// Compose decorators
func NewPropertyRepository(db *sql.DB, logger *slog.Logger, tracer trace.Tracer) PropertyRepository {
    base := &PostgresPropertyRepository{db: db}
    logged := &LoggingPropertyRepository{next: base, logger: logger}
    traced := &TracingPropertyRepository{next: logged, tracer: tracer}
    return traced
}
```

```go
// ❌ DON'T: Mix concerns in a single implementation
func (r *PostgresPropertyRepository) Save(ctx context.Context, p Property) error {
    log.Printf("saving property %s", p.ID) // Logging mixed in
    span := startSpan("Save")              // Tracing mixed in
    defer span.End()
    // ... actual save logic
}
```

### Cohesion

Keep related things together. Each type has one reason to change.

```go
// ✅ DO: High cohesion — Address knows how to validate itself
type Address struct {
    line1    string
    line2    string
    city     string
    postcode Postcode
}

func NewAddress(line1, line2, city string, postcode Postcode) (Address, error) {
    if line1 == "" {
        return Address{}, ErrAddressLine1Required
    }
    if city == "" {
        return Address{}, ErrCityRequired
    }
    return Address{line1: line1, line2: line2, city: city, postcode: postcode}, nil
}
```

```go
// ❌ DON'T: Validation scattered elsewhere
func (s *PropertyService) Create(line1, city, postcode string) error {
    if line1 == "" {
        return errors.New("line1 required") // Validation doesn't belong here
    }
    // ...
}
```

---

## Specification Pattern Testing

Tests define **behaviour** through domain interfaces. Drivers satisfy the interface — the implementation underneath is irrelevant.

### Core Concept

```
┌─────────────────────────────────────────────────────────────┐
│                    Specification Test                        │
│         (defines WHAT behaviour should occur)                │
│                                                             │
│   func PropertySpec(t *testing.T, properties PropertyStore) │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ satisfies interface
                              ▼
         ┌────────────────────┴────────────────────┐
         │                                         │
         ▼                                         ▼
┌─────────────────┐                    ┌─────────────────────┐
│  Domain Driver  │                    │    HTTP Driver      │
│  (pure domain)  │                    │  (full HTTP stack)  │
└─────────────────┘                    └─────────────────────┘
         │                                         │
         ▼                                         ▼
┌─────────────────┐                    ┌─────────────────────┐
│  In-memory impl │                    │  HTTP handlers +    │
│                 │                    │  real server        │
└─────────────────┘                    └─────────────────────┘
```

### Define the Contract

```go
// ✅ DO: Define behaviour contract as an interface
type PropertyContract interface {
    Create(property Property) error
    FindByID(id PropertyID) (Property, error)
    List(filter PropertyFilter) ([]Property, error)
}
```

### Write the Specification

```go
// ✅ DO: Specification tests behaviour, not implementation
func PropertySpec(t *testing.T, store PropertyContract) {
    t.Run("creates and retrieves a property", func(t *testing.T) {
        property := NewProperty(
            NewPropertyID(),
            MustNewAddress("123 Main St", "", "London", MustNewPostcode("SW1A 1AA")),
        )

        err := store.Create(property)
        if err != nil {
            t.Fatalf("failed to create property: %v", err)
        }

        got, err := store.FindByID(property.ID())
        if err != nil {
            t.Fatalf("failed to find property: %v", err)
        }

        if got.ID() != property.ID() {
            t.Errorf("got id %v, want %v", got.ID(), property.ID())
        }
    })

    t.Run("returns error for non-existent property", func(t *testing.T) {
        _, err := store.FindByID(NewPropertyID())
        if !errors.Is(err, ErrPropertyNotFound) {
            t.Errorf("got error %v, want %v", err, ErrPropertyNotFound)
        }
    })

    t.Run("filters properties by status", func(t *testing.T) {
        // ... test filtering behaviour
    })
}
```

### Create Drivers

```go
// ✅ DO: Pure domain driver — fastest feedback
type DomainPropertyDriver struct {
    service *PropertyService
}

func NewDomainPropertyDriver() *DomainPropertyDriver {
    repo := NewInMemoryPropertyRepository()
    service := NewPropertyService(repo, &NoOpEventPublisher{})
    return &DomainPropertyDriver{service: service}
}

func (d *DomainPropertyDriver) Create(p Property) error {
    return d.service.Create(p)
}

func (d *DomainPropertyDriver) FindByID(id PropertyID) (Property, error) {
    return d.service.FindByID(id)
}

func (d *DomainPropertyDriver) List(filter PropertyFilter) ([]Property, error) {
    return d.service.List(filter)
}
```

```go
// ✅ DO: HTTP driver — tests full stack including serialization
type HTTPPropertyDriver struct {
    baseURL string
    client  *http.Client
}

func NewHTTPPropertyDriver(server *httptest.Server) *HTTPPropertyDriver {
    return &HTTPPropertyDriver{
        baseURL: server.URL,
        client:  server.Client(),
    }
}

func (d *HTTPPropertyDriver) Create(p Property) error {
    body, _ := json.Marshal(toPropertyRequest(p))
    resp, err := d.client.Post(d.baseURL+"/properties", "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    return nil
}

func (d *HTTPPropertyDriver) FindByID(id PropertyID) (Property, error) {
    resp, err := d.client.Get(d.baseURL + "/properties/" + id.String())
    if err != nil {
        return Property{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusNotFound {
        return Property{}, ErrPropertyNotFound
    }
    var dto PropertyResponse
    json.NewDecoder(resp.Body).Decode(&dto)
    return toProperty(dto), nil
}
```

### Run Specification Against All Drivers

```go
// ✅ DO: Same specification, multiple drivers
func TestProperty_Domain(t *testing.T) {
    driver := NewDomainPropertyDriver()
    PropertySpec(t, driver)
}

func TestProperty_HTTP(t *testing.T) {
    server := setupTestServer()
    defer server.Close()
    driver := NewHTTPPropertyDriver(server)
    PropertySpec(t, driver)
}

func TestProperty_GRPC(t *testing.T) {
    server := setupGRPCTestServer()
    defer server.Stop()
    driver := NewGRPCPropertyDriver(server.Addr())
    PropertySpec(t, driver)
}
```

```go
// ❌ DON'T: Duplicate test logic per implementation
func TestHTTPCreateProperty(t *testing.T) {
    // ... duplicates domain test logic
}

func TestGRPCCreateProperty(t *testing.T) {
    // ... duplicates again
}
```

---

## Test-Driven Development (TDD)

Copilot must:

* Start with **tests first** unless explicitly told otherwise
* Use **specification pattern** for behaviour testing
* Keep tests close to the domain
* Avoid mocks unless unavoidable — prefer in-memory implementations

Tests must:

* Express intent clearly
* Be deterministic
* Be runnable against production-like code

---

## Observability (DORA-Aligned)

Observability enables **fast incident response** and **continuous improvement**. Follow DORA metrics guidance: reduce lead time, increase deployment frequency, minimize MTTR.

### Structured Logging with `slog`

Logs must tell a **story**. Each log should answer: **What happened? To what? Why does it matter?**

```go
// ✅ DO: Structured, contextual, informative logs
import "log/slog"

func (s *PropertyService) Create(ctx context.Context, p Property) error {
    logger := slog.With(
        slog.String("operation", "property.create"),
        slog.String("property_id", p.ID().String()),
        slog.String("trace_id", traceIDFromContext(ctx)),
    )

    logger.InfoContext(ctx, "creating property",
        slog.String("address_city", p.Address().City()),
        slog.String("status", p.Status().String()),
    )

    if err := s.repo.Save(ctx, p); err != nil {
        logger.ErrorContext(ctx, "failed to save property",
            slog.Any("error", err),
        )
        return fmt.Errorf("saving property: %w", err)
    }

    logger.InfoContext(ctx, "property created successfully")
    return nil
}
```

```go
// ❌ DON'T: Unstructured, noisy, useless logs
log.Printf("Creating property...")
log.Printf("Property: %+v", p) // Dumps entire struct
log.Printf("Done!")
```

### JSON Logger Setup

```go
// ✅ DO: Configure JSON logging for production
func NewLogger(env string) *slog.Logger {
    var handler slog.Handler

    switch env {
    case "production", "staging":
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level:     slog.LevelInfo,
            AddSource: true,
        })
    default:
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    return slog.New(handler)
}
```

### Log Levels

| Level | Use For |
|-------|---------|
| **Debug** | Development-only detail, never in production |
| **Info** | Significant business events (property created, user logged in) |
| **Warn** | Recoverable issues that need attention (retry succeeded, deprecated usage) |
| **Error** | Failures requiring investigation (save failed, external service down) |

### Distributed Tracing with OpenTelemetry

Traces show the **journey** of a request across services.

```go
// ✅ DO: Propagate context, create meaningful spans
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("property-service")

func (s *PropertyService) Create(ctx context.Context, p Property) error {
    ctx, span := tracer.Start(ctx, "PropertyService.Create",
        trace.WithAttributes(
            attribute.String("property.id", p.ID().String()),
            attribute.String("property.city", p.Address().City()),
        ),
    )
    defer span.End()

    if err := s.validate(ctx, p); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "validation failed")
        return err
    }

    if err := s.repo.Save(ctx, p); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "save failed")
        return err
    }

    span.SetStatus(codes.Ok, "property created")
    return nil
}
```

```go
// ❌ DON'T: Lose context, create meaningless spans
func (s *PropertyService) Create(p Property) error {
    span := tracer.Start(context.Background(), "Create") // Lost parent context
    defer span.End()
    // No attributes, no error recording
    return s.repo.Save(p)
}
```

### OpenTelemetry Setup

```go
// ✅ DO: Configure OpenTelemetry with OTLP exporter
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitTracer(ctx context.Context, serviceName, env string) (func(), error) {
    exporter, err := otlptracegrpc.New(ctx)
    if err != nil {
        return nil, fmt.Errorf("creating exporter: %w", err)
    }

    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName(serviceName),
            semconv.DeploymentEnvironment(env),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("creating resource: %w", err)
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
    )
    otel.SetTracerProvider(tp)

    return func() { tp.Shutdown(ctx) }, nil
}
```

**Free backends**: Jaeger, Zipkin, Grafana Tempo, SigNoz (self-hosted)

### Alerting: Actionable Only

Only alert if someone must **take action immediately**.

```go
// ✅ DO: Categorize errors for alerting
type ErrorSeverity int

const (
    SeverityWarning  ErrorSeverity = iota // Log, don't alert
    SeverityCritical                      // Alert, needs immediate action
)

func (s *PropertyService) Create(ctx context.Context, p Property) error {
    err := s.repo.Save(ctx, p)
    if err != nil {
        if errors.Is(err, ErrDuplicateProperty) {
            // Expected business case — warn, don't alert
            slog.WarnContext(ctx, "duplicate property rejected",
                slog.String("property_id", p.ID().String()),
            )
            return err
        }
        // Unexpected infrastructure failure — alert
        slog.ErrorContext(ctx, "CRITICAL: database save failed",
            slog.String("property_id", p.ID().String()),
            slog.Any("error", err),
            slog.String("alert", "true"), // Picked up by alerting system
        )
        return err
    }
    return nil
}
```

```go
// ❌ DON'T: Alert on everything
if err != nil {
    alerting.Send("Property creation failed!") // Alert fatigue
}
```

---

## Error Handling

Errors must be **explicit**, **contextual**, and **actionable**.

### Wrap Errors with Context

```go
// ✅ DO: Wrap errors with context, preserve original
func (s *PropertyService) Create(ctx context.Context, p Property) error {
    if err := s.repo.Save(ctx, p); err != nil {
        return fmt.Errorf("creating property %s: %w", p.ID(), err)
    }
    return nil
}
```

```go
// ❌ DON'T: Lose context or swallow errors
func (s *PropertyService) Create(ctx context.Context, p Property) error {
    s.repo.Save(ctx, p) // Error ignored!
    return nil
}

func (s *PropertyService) Create(ctx context.Context, p Property) error {
    if err := s.repo.Save(ctx, p); err != nil {
        return errors.New("save failed") // Original error lost
    }
    return nil
}
```

### Domain Errors

```go
// ✅ DO: Define domain errors, check with errors.Is
var (
    ErrPropertyNotFound   = errors.New("property not found")
    ErrDuplicateProperty  = errors.New("property already exists")
    ErrInvalidAddress     = errors.New("invalid address")
)

func (r *PostgresPropertyRepository) FindByID(ctx context.Context, id PropertyID) (Property, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... WHERE id = $1", id)
    if err := row.Scan(...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return Property{}, ErrPropertyNotFound
        }
        return Property{}, fmt.Errorf("querying property %s: %w", id, err)
    }
    return property, nil
}
```

### No Silent Failures

```go
// ✅ DO: Handle every error explicitly
resp, err := http.Get(url)
if err != nil {
    return fmt.Errorf("fetching %s: %w", url, err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
}
```

```go
// ❌ DON'T: Ignore errors
resp, _ := http.Get(url) // Error ignored
defer resp.Body.Close()  // Will panic if resp is nil
```

---

## HTMX-Specific Guidance

Copilot should:

* Prefer **server-side rendering**
* Use **partial HTML responses** for HTMX endpoints
* Keep templates **simple and readable**
* Avoid client-side state

```html
<!-- ✅ DO: Simple partial for HTMX response -->
<article class="property-card" id="property-{{.ID}}">
    <h3>{{.Address.Line1}}</h3>
    <p>{{.Address.City}}, {{.Address.Postcode}}</p>
    <span class="status status--{{.Status}}">{{.Status}}</span>
</article>
```

```go
// ✅ DO: Return partial HTML for HTMX requests
func (h *Handler) CreateProperty(w http.ResponseWriter, r *http.Request) {
    // ... create property

    // HTMX expects partial HTML
    w.Header().Set("Content-Type", "text/html")
    h.templates.ExecuteTemplate(w, "property-card.html", property)
}
```

No SPA patterns. No frontend frameworks.

---

## Continuous Deployment Mindset

Copilot must assume:

* Code is deployed **directly to production**
* No separate test or staging environments
* Feature flags or configuration are preferred over branching

All code must be:

* **Observable** — logs, traces, metrics
* **Reversible** — feature flags, backward-compatible changes
* **Safe to deploy incrementally** — no big-bang releases

---

## GitHub Copilot CLI Usage

When operating in the terminal, Copilot may:

* Generate short shell commands
* Scaffold minimal files
* Run tests, formatters, linters

Copilot must **explain intent before executing commands**.

---

## AI Boundaries (Learning First)

Copilot must NOT:

* Solve complex problems silently
* Generate large files without explanation
* Introduce patterns not already agreed

For difficult areas (e.g. auth, concurrency, networking):

* Pause
* Explain trade-offs briefly
* Proceed only after confirmation

---

## Decision Discipline

When multiple approaches exist, Copilot must:

* State the decision
* Give the rationale
* Note at least one rejected alternative

This supports future ADRs.

---

## Final Instruction

If Copilot cannot comply with these rules, it must **stop and say why**.

**Concise. Explicit. Testable. Production-first.**

