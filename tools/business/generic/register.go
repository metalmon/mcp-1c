package generic

import (
	"github.com/feenlace/mcp-1c/onec"
	businesscommon "github.com/feenlace/mcp-1c/tools/business/common"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Профиль `generic` — это «конфигурация, под которую фасада нет»: сюда
// попадает всё, кроме «Бухгалтерии предприятия 3.0» (`profile.Detect`), в том
// числе УНФ.
//
// Именованные бизнес-инструменты здесь НЕ регистрируются, и это не упрощение,
// а исправление неправды. Фасад `generic`
// (`extension/profiles/generic/src/CommonModules/MCPBusinessFacade`) —
// заглушка целиком: 15 из 17 функций бросают «generic profile: … is not
// implemented», а две оставшиеся — внутренние помощники, не инструменты.
// Раньше список инструментов у `generic` совпадал с `buh_3_0`, и агент,
// доверившись ему, звал `read_nomenclature` и получал в ответ исключение.
// Инструмент, который объявлен и не работает, хуже отсутствующего: он уводит
// агента с рабочего пути и заставляет человека думать, что сломана база.
//
// Что остаётся: инструменты, которые идут прямо в платформу, а не через
// фасад, — язык запросов и метаданные. Они работают против ЛЮБОЙ
// конфигурации, и на них же построен рабочий путь для незнакомой базы: узнать
// состав (`get_metadata_tree`), взять имена полей (`get_object_structure`),
// проверить запрос (`validate_query`), выполнить (`execute_query`).
//
// Когда у профиля появится настоящий фасад, сюда вернутся и именованные
// инструменты — но вместе с реализацией, а не прежде неё.
func RegisterTools(s *mcp.Server, onecClient *onec.Client, readOnly bool) {
	businesscommon.RegisterQueryTools(s, onecClient)
	_ = readOnly // пишущих инструментов у этого профиля нет вовсе: фасад их не реализует
}
