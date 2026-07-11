package captcha

import "github.com/mojocn/base64Captcha"

type StringCaptchaOption func(*Captcha)

type Captcha struct {
	captcha *base64Captcha.Captcha
}

func NewStringCaptcha(driver base64Captcha.Driver, store base64Captcha.Store) *Captcha {
	captcha := base64Captcha.NewCaptcha(driver, store)
	return &Captcha{captcha: captcha}
}

func (c *Captcha) Generate() (string, string, error) {
	id, b64s, _, err := c.captcha.Generate()
	if err != nil {
		return "", "", err
	}
	return id, b64s, nil
}

func (c *Captcha) Verify(id, answer string) bool {
	return c.captcha.Verify(id, answer, false)
}
