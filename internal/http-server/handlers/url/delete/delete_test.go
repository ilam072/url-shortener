package delete_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

// TODO: change test under "/url/{alias}"
func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "field Alias is a required field",
		},
		{
			name:      "DeleteURL error",
			alias:     "test_alias",
			respError: "failed to delete url",
			mockError: errors.New("unexpected error"),
		},
		{
			name:      "Alias doesn't exists",
			alias:     "test_alias",
			respError: "url not found",
			mockError: fmt.Errorf("%s: %w", "storage.sqlite.DeleteURL", storage.ErrURLNotFound),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleterMock.On("DeleteURL", mock.AnythingOfType("string")).
					Return(tc.mockError).
					Once()
			}

			handler := delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock)

			input := fmt.Sprintf(`{"alias": "%s"}`, tc.alias)

			req, err := http.NewRequest(http.MethodDelete, "/delete", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp delete.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
