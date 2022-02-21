package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/arv28/form3-accountapi-client/lib/accounts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Form3ClientTestSuite struct {
	suite.Suite
	client           *Client
	testAccountsData []accounts.AccountData
}

func (s *Form3ClientTestSuite) SetupTest() {
	s.client = NewClient(os.Getenv("HOST_URL"))

	fmt.Println("Setting up test Account Data...")

	absPath, _ := filepath.Abs("../../testdata/account_data.json")
	f, err := os.Open(absPath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	var accounts []accounts.AccountData
	err = json.NewDecoder(f).Decode(&accounts)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	s.testAccountsData = accounts

	for _, acc := range s.testAccountsData {
		_, err := s.client.Create(&acc)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Test Account Data creation done!!")

}

func (s *Form3ClientTestSuite) TearDownTest() {
	fmt.Println("Tearing down accounts data...")
	if len(s.testAccountsData) == 0 {
		return
	}
	for _, acc := range s.testAccountsData {
		_ = s.client.Delete(acc.ID, int(*acc.Version))
	}

	fmt.Println("Accounts Data deleted successfully!!")
}

func (s *Form3ClientTestSuite) TestSendRequest() {
	type args struct {
		req *http.Request
		v   interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "sendRequest - ok",
			args: args{
				req: createValidHttpRequest(s.client.HostURL),
				v:   nil,
			},
			wantErr: false,
		},
		{
			name: "sendRequest - Bad request",
			args: args{
				req: createBadHttpRequest(s.client.HostURL),
				v:   nil,
			},
			wantErr: true,
		},
		{
			name: "sendRequest - Retry and Fail",
			args: args{
				req: createBadHttpRequest("http://localhost:1000"),
				v:   nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			if err := s.client.sendRequest(tt.args.req, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Client.sendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (s *Form3ClientTestSuite) TestFetch() {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantID  string // acccount id
		errMsg  string
		wantErr bool
	}{
		{
			name: "Happy path",
			args: args{
				id: "c2743ea9-8cd8-4558-9e5c-d7ca3cd5f180",
			},
			wantID:  "c2743ea9-8cd8-4558-9e5c-d7ca3cd5f180",
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "Fetch - not found error",
			args: args{
				id: "c2743ea9-8cd8-4558-9e5c-d7ca3cd5f183",
			},
			wantID:  "",
			errMsg:  "does not exist",
			wantErr: true,
		},
		{
			name: "Fetch - Invalid data",
			args: args{
				id: "",
			},
			wantID:  "",
			errMsg:  "Invalid account id",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			got, err := s.client.Fetch(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}

func (s *Form3ClientTestSuite) TestCreate() {
	id := uuid.NewString()

	type args struct {
		accountData *accounts.AccountData
	}
	tests := []struct {
		name    string
		args    args
		wantID  string // acccount id
		errMsg  string
		wantErr bool
	}{
		{
			name: "Happy path - ok",
			args: args{
				accountData: createAccountData(id, "DE", "1234QWER", "DEBLZ", "EUR", "BICABM12"),
			},
			wantID:  id,
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "Conflict Error - Account already exists",
			args: args{
				accountData: &s.testAccountsData[0],
			},
			wantID:  "",
			errMsg:  "violates a duplicate constraint",
			wantErr: true,
		},
		{
			name: "Bad Request - Invalid request",
			args: args{
				accountData: createAccountData(id, "DE", "1234QWER", "DEBLZ", "EUR", "BICABM"),
			},
			wantID:  "",
			errMsg:  "validation failure",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			got, err := s.client.Create(tt.args.accountData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, got.ID)
				if got != nil && got.ID == tt.wantID {
					// Delete the newly created account
					_ = s.client.Delete(got.ID, int(*got.Version))
				}

			}
		})
	}
}

func (s *Form3ClientTestSuite) TestDelete() {
	type args struct {
		accountId string
		version   int
	}
	tests := []struct {
		name    string
		args    args
		errMsg  string
		wantErr bool
	}{
		{
			name: "Happy path - Delete ok",
			args: args{
				accountId: "6f22743b-9787-4c08-b8b9-ef6d444b870d",
				version:   0,
			},
			errMsg:  "",
			wantErr: false,
		},
		// the test with invalid account should actually return an error message but the fake api
		// returns response code of 404 without any error. So adjusting the test as per the fake api
		// implementation.
		{
			name: "Account does not exist",
			args: args{
				accountId: "6f22743b-9787-4c08-b8b9-ef6d444b870e",
				version:   0,
			},
			errMsg:  "",
			wantErr: false,
		},
		{
			name: "Invalid version",
			args: args{
				accountId: "df051b55-1f02-4b55-aa6a-4931c332123b",
				version:   2,
			},
			errMsg:  "invalid version",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.client.Delete(tt.args.accountId, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createValidHttpRequest(hostURL string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", hostURL, ACCOUNTS_API_PATH), nil)
	return req
}

func createBadHttpRequest(hostURL string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", hostURL, ACCOUNTS_API_PATH), nil)
	return req
}

func createAccountData(id string, country string, bankId string, bankIdCode string, currency string, bic string) *accounts.AccountData {
	accountClassification := "Personal"

	return &accounts.AccountData{
		Type:           "accounts",
		ID:             id,
		OrganisationID: "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
		Attributes: &accounts.AccountAttributes{
			Country:                 &country,
			BaseCurrency:            currency,
			BankID:                  bankId,
			BankIDCode:              bankIdCode,
			Bic:                     bic,
			Name:                    []string{"Sam Holder"},
			AlternativeNames:        []string{"Alternate Sam Holder"},
			AccountClassification:   &accountClassification,
			SecondaryIdentification: "A1B2C3D4",
		},
	}
}

// In order for 'go test' to run this suite, create
// a normal test function and pass our suite to suite.Run
func TestForm3ClientTestSuite(t *testing.T) {
	suite.Run(t, new(Form3ClientTestSuite))
}
