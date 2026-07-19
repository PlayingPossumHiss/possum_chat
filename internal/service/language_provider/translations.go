package language_provider

import "github.com/PlayingPossumHiss/possum_chat/internal/entity"

// TODO: вообще лучше, если локали и тесты лежат в отдельном файлике
// но их пока два, так что и так пойдет
var (
	defaultTranslations = map[entity.LanguageTextConstant]map[entity.ConfigLang]string{
		entity.LanguageTextConstantConnectionsTab: {
			entity.ConfigLangEn: "Connections",
			entity.ConfigLangRu: "Подключения",
		},
		entity.LanguageTextConstantCSSTab: {
			entity.ConfigLangEn: "CSS style",
			entity.ConfigLangRu: "CSS стили",
		},
		entity.LanguageTextConstantSettingsTab: {
			entity.ConfigLangEn: "Settings",
			entity.ConfigLangRu: "Настройки",
		},
		entity.LanguageTextConstantConnectionSourcesHead: {
			entity.ConfigLangEn: "Sources",
			entity.ConfigLangRu: "Источники",
		},
		entity.LanguageTextConstantConnectionSwitchesHead: {
			entity.ConfigLangEn: "Switch",
			entity.ConfigLangRu: "Тумблеры",
		},
		entity.LanguageTextConstantConnectionKeysHead: {
			entity.ConfigLangEn: "Key",
			entity.ConfigLangRu: "Ключ",
		},
		entity.LanguageTextConstantUnknownScraperIsOn: {
			entity.ConfigLangEn: "on",
			entity.ConfigLangRu: "запущен",
		},
		entity.LanguageTextConstantUnknownScraperIsOff: {
			entity.ConfigLangEn: "off",
			entity.ConfigLangRu: "остановлен",
		},
		entity.LanguageTextConstantConnectionSwitchButton: {
			entity.ConfigLangEn: "Turn",
			entity.ConfigLangRu: "Переключить",
		},
		entity.LanguageTextConstantSettingsTimeToHideMessage: {
			entity.ConfigLangEn: "Hide messages after (sec.)",
			entity.ConfigLangRu: "Скрывать сообщения через (сек.)",
		},
		entity.LanguageTextConstantSettingsTimeToDeleteMessage: {
			entity.ConfigLangEn: "Delete messages after (min.)",
			entity.ConfigLangRu: "Удалять сообщения через (мин.)",
		},
		entity.LanguageTextConstantSettingsPort: {
			entity.ConfigLangEn: "Port (need restart)",
			entity.ConfigLangRu: "Порт (нужен перезапуск)",
		},
		entity.LanguageTextConstantSettingsLang: {
			entity.ConfigLangEn: "Language (need restart)",
			entity.ConfigLangRu: "Язык (нужен перезапуск)",
		},
		entity.LanguageTextConstantSettingsShowOnline: {
			entity.ConfigLangEn: "Show online",
			entity.ConfigLangRu: "Отображать онлайн",
		},
		entity.LanguageTextConstantAppVersion: {
			entity.ConfigLangEn: "Version",
			entity.ConfigLangRu: "Версия",
		},
		entity.LanguageTextConstantTestMessageButton: {
			entity.ConfigLangEn: "Test",
			entity.ConfigLangRu: "Тест",
		},
		entity.LanguageTextConstantTestMessageContent: {
			entity.ConfigLangEn: "The quick brown fox jumps over the lazy dog",
			entity.ConfigLangRu: "Съешь ещё этих мягких французских булок, да выпей же чаю",
		},
		entity.LanguageTextConstantMainStyleField: {
			entity.ConfigLangEn: "Main style",
			entity.ConfigLangRu: "Основной стиль",
		},
		entity.LanguageTextConstantWidgetOBS: {
			entity.ConfigLangEn: "OBS widget",
			entity.ConfigLangRu: "Виджет OBS",
		},
		entity.LanguageTextConstantMessagePanel: {
			entity.ConfigLangEn: "All messages",
			entity.ConfigLangRu: "Все сообщения",
		},
		entity.LanguageTextConstantMyGithub: {
			entity.ConfigLangEn: "My GitHub",
			entity.ConfigLangRu: "Мой GitHub",
		},
		entity.LanguageTextConstantYoutubeConnPlaceholder: {
			entity.ConfigLangEn: "Channel name (without @) or stream ID",
			entity.ConfigLangRu: "Имя канала (без @) или ID трансляции",
		},
		entity.LanguageTextConstantTwitchConnPlaceholder: {
			entity.ConfigLangEn: "Channel name (without @)",
			entity.ConfigLangRu: "Имя канала (без @)",
		},
		entity.LanguageTextConstantDaConnPlaceholder: {
			entity.ConfigLangEn: "Widget url token",
			entity.ConfigLangRu: "Токен из виджета",
		},
	}
)
