package language_provider

import (
	"fmt"

	"github.com/PlayingPossumHiss/possum_chat/internal/entity"
	"github.com/PlayingPossumHiss/possum_chat/internal/service/logger"
)

type LanguageProvider struct {
	translations map[entity.LanguageTextConstant]map[entity.ConfigLang]string
	lang         entity.ConfigLang
}

func New(lang entity.ConfigLang) *LanguageProvider {
	return &LanguageProvider{
		lang:         lang,
		translations: defaultTranslations,
	}
}

func (lp *LanguageProvider) Local(name entity.LanguageTextConstant) string {
	translations, exist := lp.translations[name]
	if !exist {
		logger.Error(fmt.Sprintf("cant find text constant %s", name))
	}

	translation, exist := translations[lp.lang]
	if !exist {
		return translations[entity.ConfigLangEn]
	}

	return translation
}
