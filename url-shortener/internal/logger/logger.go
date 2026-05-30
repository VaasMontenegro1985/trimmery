package logger

import "go.uber.org/zap"

func New(level string) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}

	if level != "" {
		if err := config.Level.UnmarshalText([]byte(level)); err != nil {
			return nil, err
		}
	}

	return config.Build()
}
