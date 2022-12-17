package logger

import (
	"github.com/mavolin/dblog/generator/file"
)

type (
	// Logger is the abstraction of a logger.
	// It should implement at least one of the below interfaces.
	Logger interface {
		// Imports returns the import paths of the imports that are required
		// for the logger.
		Imports() []string
	}

	// PreRequestLogger is a [Logger] that logs before a request is made.
	//
	// It returns the code that is used for logging.
	PreRequestLogger interface {
		LogPreRequest(file.Method) (string, error)
	}

	// PostRequestLogger is a [Logger] that logs after a request is made,
	// regardless of whether an error occurred.
	//
	// It returns the code that is used for logging.
	PostRequestLogger interface {
		LogPostRequest(file.Method) (string, error)
	}

	// SuccessLogger is the same as a [PostRequestLogger], but only logs if no
	// error occurred.
	//
	// If the generator does not have an [error] as its last return value,
	// [LogSuccess] will always be called.
	//
	// It returns the code that is used for logging.
	SuccessLogger interface {
		LogSuccess(file.Method) (string, error)
	}

	// ErrorLogger is the same as a [PostRequestLogger], but only logs if an
	// error occurred.
	//
	// If the generator does not have an [error] as its last return value,
	// [LogError] will never be called.
	//
	// It returns the code that is used for logging.
	ErrorLogger interface {
		LogError(file.Method) (string, error)
	}
)
