package model

import (
	"context"

	"github.com/gofrs/uuid"
)

type Member struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Role           string
}

func ContextWithMember(ctx context.Context, member *Member) context.Context {
	return context.WithValue(ctx, ContextKeyMember, member)
}

func MemberFromContext(ctx context.Context) (*Member, bool) {
	v, ok := ctx.Value(ContextKeyMember).(*Member)
	return v, ok
}
