package main

import (
	"github.com/tombenke/go-12f-common/v2/apprun"
	"github.com/tombenke/go-12f-common/v2/must"
	"github.com/tombenke/ports-and-adapters-example/internal/infrastructure/di/server12f"
)

func main() {
	// Make and run an application via ApplicationRunner
	must.Must(apprun.MakeAndRun(&server12f.Config{}, server12f.NewApplication))
}
