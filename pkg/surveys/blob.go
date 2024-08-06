package surveys

import (
	"github.com/plutov/formulosity/pkg/services"
)

func UploadCustomTheme(svc services.Services, urlSlug string, themeContents string) (string, error) {
	if themeContents == "" {
		return "", nil
	}

	// TODO: implement custom theme upload

	return "", nil
}
