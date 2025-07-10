package app

import (
	"fmt"

	"github.com/superboomer/tg-spec-cli/internal/generator"
	"github.com/superboomer/tg-spec-cli/internal/telegram"

	"go.uber.org/zap"
)

type App struct {
	log        *zap.Logger
	url        string
	outputPath string
	typeFlag   string
}

func New(log *zap.Logger, url, outputPath string) *App {
	return &App{
		log:        log,
		url:        url,
		outputPath: outputPath,
	}
}

func NewWithType(log *zap.Logger, url, outputPath, typeFlag string) *App {
	return &App{
		log:        log,
		url:        url,
		outputPath: outputPath,
		typeFlag:   typeFlag,
	}
}

func (a *App) Run() error {
	a.log.Info("starting app")

	if a.typeFlag != "botapi" && a.typeFlag != "gateway" {
		a.log.Error("unsupported API type", zap.String("type", a.typeFlag))
		return fmt.Errorf("unsupported API type: %s", a.typeFlag)
	}

	a.log.Debug("fetching Telegram API page", zap.String("url", a.url), zap.String("type", a.typeFlag))

	page, err := telegram.GetPage(a.url)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}
	a.log.Debug("successfully fetched page")

	version, err := page.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}
	a.log.Info("got version", zap.String("version", version))

	types, err := page.GetTypes()
	if err != nil {
		return fmt.Errorf("failed to get types: %w", err)
	}
	a.log.Info("got types", zap.Int("count", len(types)))
	if a.log.Core().Enabled(zap.DebugLevel) {
		typeNames := make([]string, 0, len(types))
		for name := range types {
			typeNames = append(typeNames, name)
		}
		a.log.Debug("type names", zap.Strings("types", typeNames))
	}

	methods, err := page.GetMethods()
	if err != nil {
		return fmt.Errorf("failed to get methods: %w", err)
	}
	a.log.Info("got methods", zap.Int("count", len(methods)))
	if a.log.Core().Enabled(zap.DebugLevel) {
		methodNames := make([]string, 0, len(methods))
		for _, m := range methods {
			methodNames = append(methodNames, m.Name)
		}
		a.log.Debug("method names", zap.Strings("methods", methodNames))
	}

	gen := generator.NewWithType(a.log, version, types, methods, a.typeFlag)
	a.log.Debug("generating OpenAPI schema")
	openAPI, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate openapi: %w", err)
	}
	a.log.Debug("OpenAPI schema generated")

	a.log.Debug("saving OpenAPI schema", zap.String("outputPath", a.outputPath))
	if err := gen.Save(openAPI, a.outputPath); err != nil {
		return fmt.Errorf("failed to save openapi: %w", err)
	}

	a.log.Info("finished app")
	return nil
}
