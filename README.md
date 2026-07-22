# Rate Limiter Middleware

> Token bucket rate limiter for HTTP services — portable Go module.

## Features
- Token bucket algorithm (allows bursting)
- Per-IP rate limiting with configurable capacity/rate
- Lightweight, zero external dependencies
- Works as Go middleware or standalone library

## Usage

```go
import limiter "github.com/mojojojobeat26-arch/rate-limiter-middleware"

l := limiter.NewLimiter()
handler := l.Middleware(100, 10)(myHandler) // 100 burst, 10/sec
```

## Why I built this

A simple, correct rate limiter using the token bucket algorithm. No external deps, clean API.

— Faraz

## License
MIT — © 2026 Faraz Mustafa Seyed