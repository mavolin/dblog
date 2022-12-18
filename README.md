<div align="center">
<h1>dblog</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/mavolin/dblog)](https://goreportcard.com/report/github.com/mavolin/dblog)
[![License MIT](https://img.shields.io/github/license/mavolin/dblog)](https://github.com/mavolin/dblog/blob/develop/LICENSE)
</div>

---

## About

dblog is a code generator that can generate a wrapper for your repository
interface to log database operations.
Especially useful, if you do things like performance tracking.

## Features

* ✨ **Simple:** Only generate code for the logger you use.
* 📦 **Modular:** Easily add more generators by implementing the `logger.Logger`
  interface, the [sentry implementation](logger/sentry/sentry.go) is only 143 LOC long.
  (PRs welcome!)
* 🕵 **Context-Aware:** Loggers can log before a request, after a request,
  when an error occurs, when the request succeeds, or any combination of those.

## Installation

```shell
go install github.com/mavolin/dblog/cmd/dblog@latest
```

## Examples

First impressions matter, so here is a simple example:

```go
package my-project/repository

import "context"

//go:generate dblog -sentry Repository
type (
	Repository interface {
		// dblog uses param and return value names in the log messages, so it makes
		// sense to name them yourself, if you don't want generated names.
		Gopher(ctx context.Context, id GopherID) (g *Gopher, err error)
		Gophers(ctx context.Context, search GopherSearchData) (gs []Gopher, err error)
	}

	GopherSearchData struct {
		Name           string
		MinAge, MaxAge int
	}
)

type GopherID uint64

type Gopher struct {
	ID      GopherID
	Name    string
	Age     int
	Hobbies []string
}

```

> **Quick Transparency Note:** 
> Variables in the below code were renamed for readability,
> normally more cryptic names are used to prevent name collisions.
> Additionally, I changed naked returns to explicit ones.
> Otherwise, the code is identical to the actual output of `dblog`.

```go
package my-project/repository/dblog

import (
    "context"
    "github.com/getsentry/sentry-go"
    "my-project/repository"
)

// Code generated by github.com/mavolin/dblog. DO NOT EDIT.

type Wrapper struct {
	repo repository.Repository
}

var _ repository.Repository = (*Wrapper)(nil)

func NewWrapper(repo repository.Repository) *Wrapper {
	return &Wrapper{repo: repo}
}

func (w *Wrapper) Gopher(ctx context.Context, id repository.GopherID) (g *repository.Gopher, err error) {
	span := sentry.StartSpan(ctx, "db.query")
	span.Description = "Gopher"
	ctx = span.Context()

	defer span.Finish()

	g, err = w.repo.Gopher(ctx, id)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "db.query",
			Message:  "Gopher",
			Data: map[string]any{
				"args":    map[string]any{"id": id},
				"results": map[string]any{"g": g, "err": err},
			},
		}, nil)
	}

	return g, err
}

func (w *Wrapper) Gophers(
	ctx context.Context, search repository.GopherSearchData,
) (gs []repository.Gopher, err error) {
	span := sentry.StartSpan(ctx, "db.query")
	span.Description = "Gophers"
	ctx = span.Context()

	defer span.Finish()

	gs, err = w.repo.Gophers(ctx, search)

	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "db.query",
			Message:  "Gophers",
			Data: map[string]any{
				"args":    map[string]any{"search": search},
				"results": map[string]any{"gs": gs, "err": err},
			},
		}, nil)
	}
	
	return gs, err
}
```

## License

Built with ❤ by [Maximilian von Lindern](https://github.com/mavolin).
Available under the [MIT License](./LICENSE).