package middleware

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/slipros/roamer/examples/custom_parser/model"
)

func Member(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(model.ContextWithMember(r.Context(), &model.Member{
			ID:             uuid.Must(uuid.NewV7()),
			OrganizationID: uuid.Must(uuid.NewV7()),
			Name:           "Jack",
			Role:           "Admin",
		})))
	}
}
