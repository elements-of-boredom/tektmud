package language

import (
	"errors"
	"path/filepath"
	configs "tektmud/internal/config"
	"tektmud/internal/logger"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	ErrMessageFallback = errors.New("translation message fallback to default language")
)

var translation *Translation

type Translation struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
}

func Initialize() {
	translation = NewTranslation()
}

func NewTranslation() *Translation {
	t := &Translation{}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	t.bundle = bundle
	c := configs.GetConfig()
	t.bundle.LoadMessageFile(filepath.Join(c.Paths.RootDataDir, c.Paths.Localization, "en.yaml"))
	t.localizer = i18n.NewLocalizer(t.bundle, "en")

	return t
}

func T(msgId string, tplData ...map[any]any) string {
	lng := language.Make("en")

	msg, err := translation.Translate(lng, msgId, tplData...)
	if err != nil {
		if !IsMessageFallbackErr(err) && !IsMessageNotFoundErr(err) {
			logger.Error("Translation", "msgId", msgId, "error", err)
		}
	}
	return msg
}

func (t *Translation) Translate(lng language.Tag, msgId string, tplData ...map[any]any) (string, error) {

	if t.localizer == nil {
		return msgId, nil
	}

	cfg := &i18n.LocalizeConfig{
		MessageID: msgId,
	}

	if len(tplData) > 0 {
		cfg.TemplateData = tplData[0]
	}

	msg, l, err := translation.localizer.LocalizeWithTag(cfg)
	if err != nil {
		//Fallback to english
		if !l.IsRoot() {
			return msg, ErrMessageFallback
		}

		//We couldn't find the id, so return it to be used.
		return msgId, err
	}

	if l != lng {
		return msg, ErrMessageFallback
	}

	return msg, nil
}

func IsMessageNotFoundErr(err error) bool {
	_, ok := err.(*i18n.MessageNotFoundErr)

	return ok
}

func IsMessageFallbackErr(err error) bool {
	return errors.Is(err, ErrMessageFallback)
}
