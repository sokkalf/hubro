package helpers

import (
	"html/template"

	"github.com/sokkalf/hubro/config"
)

func GetLogoImage() template.HTML {
	switch config.Config.LogoImage {
	case "":
		return template.HTML(`<span class="text-6xl">🦉</span>`)
	default:
		return template.HTML(`<img src="` + config.Config.RootPath +
			"userfiles/" +config.Config.LogoImage + `" alt="Logo" class="avatar">`)
	}
}
