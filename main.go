package main

import (
	"fmt"
	"os"

	"github.com/bata94/apiright/cmd/apiright"

	"github.com/bata94/apiright/pkg/config"
	"github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/database"
	"github.com/bata94/apiright/pkg/generator"
	"github.com/bata94/apiright/pkg/server"
)

type (
	Schema                = core.Schema
	Table                 = core.Table
	Column                = core.Column
	Index                 = core.Index
	ForeignKey            = core.ForeignKey
	Query                 = core.Query
	Param                 = core.Param
	Enum                  = core.Enum
	Type                  = core.Type
	Service               = core.Service
	ServiceMethod         = core.ServiceMethod
	ErrorResponse         = core.ErrorResponse
	GenerationContext     = core.GenerationContext
	ContentNegotiatorImpl = core.ContentNegotiatorImpl
	ContentInfo           = core.ContentInfo
	Logger                = core.Logger
	ZapLogger             = core.ZapLogger
	StatusCode            = core.StatusCode

	Generator        = generator.Generator
	GenerateOptions  = generator.GenerateOptions
	ProgressReporter = generator.ProgressReporter

	Config           = config.Config
	ProjectConfig    = config.ProjectConfig
	DatabaseConfig   = config.DatabaseConfig
	ServerConfig     = config.ServerConfig
	GenerationConfig = config.GenerationConfig
	PluginConfig     = config.PluginConfig
	TLSConfig        = config.TLSConfig

	DualServer = server.DualServer

	Database        = database.Database
	Migration       = database.Migration
	MigrationResult = database.MigrationResult
)

const (
	StatusOK                  = core.StatusOK
	StatusNotFound            = core.StatusNotFound
	StatusBadRequest          = core.StatusBadRequest
	StatusUnauthorized        = core.StatusUnauthorized
	StatusForbidden           = core.StatusForbidden
	StatusInternalServerError = core.StatusInternalServerError
	StatusConflict            = core.StatusConflict
	StatusUnprocessableEntity = core.StatusUnprocessableEntity

	ContentTypeJSON     = core.ContentTypeJSON
	ContentTypeXML      = core.ContentTypeXML
	ContentTypeYAML     = core.ContentTypeYAML
	ContentTypeProtobuf = core.ContentTypeProtobuf
	ContentTypeText     = core.ContentTypeText
	ContentTypeHTML     = core.ContentTypeHTML
	ContentTypeForm     = core.ContentTypeForm

	DebugLevel  = core.DebugLevel
	InfoLevel   = core.InfoLevel
	WarnLevel   = core.WarnLevel
	ErrorLevel  = core.ErrorLevel
	DPanicLevel = core.DPanicLevel
	PanicLevel  = core.PanicLevel
	FatalLevel  = core.FatalLevel
)

var NewLogger = core.NewLogger
var NewLoggerWithLevel = core.NewLoggerWithLevel
var SyncLogger = core.SyncLogger
var NewContentNegotiator = core.NewContentNegotiator
var LoadConfig = config.LoadConfig
var DefaultConfig = config.DefaultConfig
var ValidateConfig = config.ValidateConfig
var SaveConfig = config.SaveConfig
var NewDatabase = database.NewDatabase
var NewGenerator = generator.NewGenerator
var NewProgressReporter = generator.NewProgressReporter
var NewServer = server.NewServer
var NewGenerationContext = core.NewGenerationContext
var Close = core.Close
var Rollback = core.Rollback
var SQLToGoType = core.SQLToGoType
var SQLToProtoType = core.SQLToProtoType
var SQLTypeToOpenAPI = core.SQLTypeToOpenAPI
var GetExampleValue = core.GetExampleValue
var ToPascalCase = core.ToPascalCase
var ParseContentHeader = core.ParseContentHeader

var Version = core.Version

func main() {
	if err := apiright.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
