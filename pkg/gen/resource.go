package gen

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type App struct {
	Name         string
	BuildCommand string
	RunCommand   string
}

func NewAppResource(app App) *schema.ResourceData {
	var d *schema.ResourceData

	d.Set("name", app.Name)
	d.Set("build_command", app.BuildCommand)
	d.Set("run_command", app.RunCommand)

	return d
}
