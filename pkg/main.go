package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/mewc/posthog-grafana-plugin/pkg/plugin"
)

func main() {
	if err := datasource.Manage("mewc-posthog-datasource", plugin.NewPostHogDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
