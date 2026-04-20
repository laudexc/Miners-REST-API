package myHttp

import "prj2/logic"

// internal/transport/http
// HTTP-обработчики, роутинг, валидация входных данных, маппинг ошибок в HTTP-ответы.
type HTTPHandlers struct {
	enterprise *logic.Enterprise
}

func NewHTTPHandlers(Entp *logic.Enterprise) *HTTPHandlers {
	return &HTTPHandlers{
		enterprise: Entp,
	}
}

