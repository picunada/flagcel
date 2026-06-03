module github.com/picunada/flagcel/sdks/go

go 1.26.2

require (
	github.com/open-feature/go-sdk v1.17.2
	github.com/picunada/flagcel/evalcore v0.0.0
)

require (
	cel.dev/expr v0.25.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/google/cel-go v0.28.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20240823005443-9b4947da3948 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240826202546-f6391c0de4c7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240826202546-f6391c0de4c7 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/picunada/flagcel/evalcore => ../../evalcore
