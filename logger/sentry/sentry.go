package sentry

import (
	"errors"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/urfave/cli/v2"

	"github.com/mavolin/dblog/generator/file"
	"github.com/mavolin/dblog/logger"
)

type Logger struct {
	Category             string
	NoArgs               bool
	NoResults, OnlyError bool

	NoPerf bool
	SpanOp string
}

var (
	_ logger.Logger            = (*Logger)(nil)
	_ logger.PreRequestLogger  = (*Logger)(nil)
	_ logger.PostRequestLogger = (*Logger)(nil)
)

func NewLogger(app *cli.App) *Logger {
	var l Logger

	app.Flags = append(app.Flags,
		&cli.StringFlag{
			Name:        "sentry-category",
			Value:       "db.query",
			Usage:       "The category to use for the breadcrumb.",
			Destination: &l.Category,
		},
		&cli.BoolFlag{
			Name:        "sentry-no-args",
			Value:       false,
			Usage:       "Don't include the arguments in the breadcrumb.",
			Destination: &l.NoArgs,
		},
		&cli.BoolFlag{
			Name:        "sentry-no-results",
			Value:       false,
			Usage:       "Don't include the return values in the breadcrumb.",
			Destination: &l.NoResults,
		},
		&cli.BoolFlag{
			Name:        "sentry-only-error",
			Value:       false,
			Usage:       "Only include the error and no other return values in the breadcrumb.",
			Destination: &l.OnlyError,
		},
		&cli.BoolFlag{
			Name:        "sentry-no-perf",
			Value:       false,
			Usage:       "Disable performance monitoring.",
			Destination: &l.NoPerf,
		},
		&cli.StringFlag{
			Name:        "sentry-span-op",
			Value:       "db.query",
			Usage:       "The operation name to use for the span.",
			Destination: &l.SpanOp,
		},
	)

	return &l
}

func (l *Logger) Imports() []string {
	return []string{"github.com/getsentry/sentry-go"}
}

func (l *Logger) LogPreRequest(m file.Method) (string, error) {
	if l.NoPerf {
		return "", nil
	}

	var b strings.Builder

	if len(m.Params) == 0 || m.Params[0].Type.String() != "context.Context" {
		return "", errors.New("methods must accept a context.Context as their first argument")
	}

	ctxName := m.Params[0].Name

	b.WriteString("__sentry_span := sentry.StartSpan(")
	b.WriteString(ctxName)
	b.WriteString(", ")
	b.WriteString(strconv.Quote(l.SpanOp))
	b.WriteString(")\n")
	//
	b.WriteString("__sentry_span.Description = ")
	b.WriteString(strconv.Quote(m.Name))
	b.WriteString("\n")
	//
	b.WriteString(ctxName)
	b.WriteString(" = __sentry_span.Context()\n\n")

	b.WriteString("defer __sentry_span.Finish()")

	return b.String(), nil
}

func (l *Logger) LogPostRequest(m file.Method) (string, error) {
	var b strings.Builder

	if len(m.Params) == 0 || m.Params[0].Type.String() != "context.Context" {
		return "", errors.New("methods must accept a context.Context as their first argument")
	}

	ctxName := m.Params[0].Name

	b.WriteString("if __sentry_hub := sentry.GetHubFromContext(")
	b.WriteString(ctxName)
	b.WriteString("); __sentry_hub != nil {\n")
	{
		b.WriteString("__sentry_hub.AddBreadcrumb(&sentry.Breadcrumb{\n")
		b.WriteString("Category: ")
		b.WriteString(strconv.Quote(l.Category))
		b.WriteString(",\n")
		b.WriteString("Message: ")
		b.WriteString(strconv.Quote(m.Name))
		b.WriteString(",\n")

		b.WriteString("Data: map[string]any{\n")
		if !l.NoArgs {
			b.WriteString("\"args\": map[string]any{")
			for i, p := range m.Params[1:] {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(strconv.Quote(strcase.ToSnake(p.Name)))
				b.WriteString(": ")
				b.WriteString(p.Name)
			}
			b.WriteString("},\n")
		}
		if !l.NoResults {
			b.WriteString("\"results\": map[string]any{")
			if l.OnlyError {
				if len(m.Returns) > 0 && m.Returns[0].Type.String() == "error" {
					b.WriteString("err: ")
					b.WriteString(m.Returns[0].Name)
				}
			} else {
				for i, r := range m.Returns {
					if i > 0 {
						b.WriteString(", ")
					}
					b.WriteString(strconv.Quote(strcase.ToSnake(r.Name)))
					b.WriteString(": ")
					b.WriteString(r.Name)
				}
			}
			b.WriteString("},\n")
		}
		b.WriteString("},\n")

		b.WriteString("}, nil)\n")
	}
	b.WriteString("}")

	return b.String(), nil
}
