package captcha

import "github.com/mojocn/base64Captcha"

func NewStringDriver() base64Captcha.Driver {
	source := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return base64Captcha.NewDriverString(
		80, 240, 6, 1, 4, source, nil, nil, nil,
	).ConvertFonts()
}
