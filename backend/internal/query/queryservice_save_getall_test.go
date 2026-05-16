package query_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"gohole/internal/database"
	mockdb "gohole/internal/mock/database"
	mockfilter "gohole/internal/mock/filter"
	"gohole/internal/query"
)

func newService(t *testing.T) (query.Service, *mockdb.MockRepository, *mockfilter.MockFilter, *mockfilter.MockFilter) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mockdb.NewMockRepository(ctrl)
	blockFilter := mockfilter.NewMockFilter(ctrl)
	allowFilter := mockfilter.NewMockFilter(ctrl)
	svc := query.NewService(blockFilter, allowFilter, repo)
	return svc, repo, blockFilter, allowFilter
}

// ---- Save ----

func TestSave_OK(t *testing.T) {
	svc, repo, _, _ := newService(t)
	q := database.Query{Name: "example.com", Blocked: false}
	repo.EXPECT().SaveQuery(gomock.Any(), q).Return(nil)

	if err := svc.Save(context.Background(), q); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSave_Error(t *testing.T) {
	svc, repo, _, _ := newService(t)
	q := database.Query{Name: "example.com"}
	repo.EXPECT().SaveQuery(gomock.Any(), q).Return(errors.New("db error"))

	if err := svc.Save(context.Background(), q); err == nil {
		t.Error("expected error, got nil")
	}
}

// ---- GetAll ----

func TestGetAll_OK(t *testing.T) {
	svc, repo, _, _ := newService(t)
	expected := []database.Query{{Name: "a.com"}, {Name: "b.com"}}
	repo.EXPECT().FindAllLimit(gomock.Any(), 10, "").Return(expected, nil)

	got, err := svc.GetAll(context.Background(), 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 results, got %d", len(got))
	}
}

func TestGetAll_Error(t *testing.T) {
	svc, repo, _, _ := newService(t)
	repo.EXPECT().FindAllLimit(gomock.Any(), 5, "foo").Return(nil, errors.New("db error"))

	_, err := svc.GetAll(context.Background(), 5, "foo")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
