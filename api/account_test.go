package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/atanda0x/FintechConnect/db/mock"
	db "github.com/atanda0x/FintechConnect/db/sqlc"
	"github.com/atanda0x/FintechConnect/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}


func TestGetAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		expectedCode  int
		expectedBody  string
		buildStubs    func(store *mockdb.MockStore)
	}{
		{
			name:         "OK",
			accountID:    account.ID,
			expectedCode: http.StatusOK,
			expectedBody: accountJSON(account),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
		},
		{
			name:         "NotFound",
			accountID:    account.ID,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
		},
		{
			name:         "InternalError",
			accountID:    account.ID,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
		},
		{
			name:         "InvalidID",
			accountID:    0,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// Check response status code
			require.Equal(t, tc.expectedCode, recorder.Code)

			// Check response body if expected
			if tc.expectedBody != "" {
				body, err := ioutil.ReadAll(recorder.Body)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}

func TestListAccount(t *testing.T) {
    accounts := []db.Account{
        {ID: 1, Name: "Alice", Balance: 100},
        {ID: 2, Name: "Bob", Balance: 200},
    }

    testCases := []struct {
        name          string
        pageID        int
        pageSize      int
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
    }{
        {
            name:     "OK",
            pageID:   1,
            pageSize: 10,
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().ListAccounts(gomock.Any(), gomock.Any()).Times(1).Return(accounts, nil)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                requireBodyMatchAccounts(t, recorder.Body, accounts)
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            // build stubs
            store := mockdb.NewMockStore(ctrl)
            tc.buildStubs(store)

            // start test server and send request
            server := NewServer(store)
            recorder := httptest.NewRecorder()

            url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.pageID, tc.pageSize)
            request, err := http.NewRequest(http.MethodGet, url, nil)
            require.NoError(t, err)

            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(t, recorder)
        })
    }
}


func TestUpdateAccount(t *testing.T) {
    account := randomAccount()
    req := updateAccountRequest{
        ID:      account.ID,
        Balance: 50,
    }

    testCases := []struct {
        name          string
        request       updateAccountRequest
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
    }{
        {
            name:    "OK",
            request: req,
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(req.ID)).Times(1).Return(account, nil)
                store.EXPECT().UpdateAccount(gomock.Any(), gomock.Any()).Times(1).Return(nil)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                // You can check the response body here if needed
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            // build stubs
            store := mockdb.NewMockStore(ctrl)
            tc.buildStubs(store)

            // start test server and send request
            server := NewServer(store)
            recorder := httptest.NewRecorder()

            url := fmt.Sprintf("/accounts/%d", req.ID)
            body, err := json.Marshal(req)
            require.NoError(t, err)
            request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
            require.NoError(t, err)

            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(t, recorder)
        })
    }
}

func TestDeleteAccount(t *testing.T) {
    account := randomAccount()
    req := deleteAccountRequest{
        ID: account.ID,
    }

    testCases := []struct {
        name          string
        request       deleteAccountRequest
        buildStubs    func(store *mockdb.MockStore)
        checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
    }{
        {
            name:    "OK",
            request: req,
            buildStubs: func(store *mockdb.MockStore) {
                store.EXPECT().DeleteAccount(gomock.Any(), gomock.Eq(req.ID)).Times(1).Return(nil)
            },
            checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
                require.Equal(t, http.StatusOK, recorder.Code)
                // You can check the response body here if needed
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            // build stubs
            store := mockdb.NewMockStore(ctrl)
            tc.buildStubs(store)

            // start test server and send request
            server := NewServer(store)
            recorder := httptest.NewRecorder()

            url := fmt.Sprintf("/accounts/%d", req.ID)
            request, err := http.NewRequest(http.MethodDelete, url, nil)
            require.NoError(t, err)

            server.router.ServeHTTP(recorder, request)
            tc.checkResponse(t, recorder)
        })
    }
}


func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: "NGN",
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var getAccount db.Account
	err = json.Unmarshal(data, &getAccount)
	require.NoError(t, err)
	require.Equal(t, account, getAccount)
}


func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, expected []db.Account) {
    var accounts []db.Account
    err := json.NewDecoder(body).Decode(&accounts)
    require.NoError(t, err)

    require.Equal(t, len(expected), len(accounts))
    for i := range expected {
        require.Equal(t, expected[i].ID, accounts[i].ID)
        require.Equal(t, expected[i].Name, accounts[i].Name)
        require.Equal(t, expected[i].Balance, accounts[i].Balance)
    }
}
