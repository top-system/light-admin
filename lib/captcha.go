package lib

import (
	"image/color"
	"time"

	"go.uber.org/zap"

	"github.com/top-system/light-admin/constants"
	"github.com/mojocn/base64Captcha"
)

type Captcha struct {
	*base64Captcha.Captcha
}

type CaptchaStore struct {
	key    string
	cache  Cache
	logger *zap.SugaredLogger
}

func NewCaptcha(cache Cache, logger Logger) Captcha {
	ds := base64Captcha.NewDriverString(
		46,
		140,
		2,
		2,
		4,
		"234567890abcdefghjkmnpqrstuvwxyz",
		&color.RGBA{R: 240, G: 240, B: 246, A: 246},
		nil,
		[]string{"wqy-microhei.ttc"},
	)

	driver := ds.ConvertFonts()
	store := &CaptchaStore{
		cache:  cache,
		key:    constants.CaptchaKeyPrefix,
		logger: logger.Zap.With(zap.String("module", "captcha")),
	}

	return Captcha{Captcha: base64Captcha.NewCaptcha(driver, store)}
}

func (a *CaptchaStore) getKey(v string) string {
	return a.key + ":" + v
}

func (a *CaptchaStore) Set(id string, value string) error {
	err := a.cache.Set(a.getKey(id), value, time.Second*constants.CaptchaExpireTimes)
	if err != nil {
		a.logger.Errorf("captcha - error writing cache: %v", err)
	}
	return err
}

func (a *CaptchaStore) Get(id string, clear bool) string {
	var (
		key = a.getKey(id)
		val string
	)

	err := a.cache.Get(key, &val)
	if err != nil {
		a.logger.Errorf("captcha - error reading cache: %v", err)
		return ""
	}

	if clear {
		_, err := a.cache.Delete(key)
		if err != nil {
			a.logger.Errorf("captcha - error deleting item from cache: %v", err)
		}
	}

	return val
}

func (a *CaptchaStore) Verify(id, answer string, clear bool) bool {
	v := a.Get(id, clear)
	return v == answer
}
