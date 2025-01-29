package andromeda

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	host             = "http://192.168.0.18:9002/api"
	endpointGetSites = "/Sites"

	defaultTimeout = 5 * time.Second
)

type (
	GetSitesInput struct {
		Id       int
		UserName string
		ApiKey   string
	}

	request struct {
		URL    string
		body   []byte
		apiKey string
	}

	RespErr400 struct {
		Message      string `json:"Message"`
		SpResultCode int    `json:"SpResultCode"`
	}

	GetSitesResponse struct {
		RowNumber                  int     `json:"RowNumber"`
		Id                         string  `json:"Id"`
		AccountNumber              int     `json:"AccountNumber"`
		CloudObjectID              int     `json:"CloudObjectID"`
		Name                       string  `json:"Name"`
		ObjectPassword             string  `json:"ObjectPassword"`
		Address                    string  `json:"Address"`
		Phone1                     string  `json:"Phone1"`
		Phone2                     string  `json:"Phone2"`
		TypeName                   string  `json:"TypeName"`
		IsFire                     bool    `json:"IsFire"`
		IsArm                      bool    `json:"IsArm"`
		IsPanic                    bool    `json:"IsPanic"`
		DeviceTypeName             string  `json:"DeviceTypeName"`
		EventTemplateName          string  `json:"EventTemplateName"`
		ContractNumber             string  `json:"ContractNumber"`
		ContractPrice              float64 `json:"ContractPrice"`
		MoneyBalance               float64 `json:"MoneyBalance"`
		PaymentDate                string  `json:"PaymentDate"`
		DebtInformLevel            int     `json:"DebtInformLevel"`
		Disabled                   bool    `json:"Disabled"`
		DisableReason              int     `json:"DisableReason"`
		DisableDate                string  `json:"DisableDate"`
		AutoEnable                 bool    `json:"AutoEnable"`
		AutoEnableDate             string  `json:"AutoEnableDate"`
		CustomersComment           string  `json:"CustomersComment"`
		CommentForOperator         string  `json:"CommentForOperator"`
		CommentForGuard            string  `json:"CommentForGuard"`
		MapFileName                string  `json:"MapFileName"`
		WebLink                    string  `json:"WebLink"`
		ControlTime                int     `json:"ControlTime"`
		CTIgnoreSystemEvent        bool    `json:"CTIgnoreSystemEvent"`
		IsContractPriceForceUpdate bool    `json:"IsContractPriceForceUpdate"`
		IsMoneyBalanceForceUpdate  bool    `json:"IsMoneyBalanceForceUpdate"`
		IsPaymentDateForceUpdate   bool    `json:"IsPaymentDateForceUpdate"`
		IsStateArm                 bool    `json:"IsStateArm"`
		IsStateAlarm               bool    `json:"IsStateAlarm"`
		IsStatePartArm             bool    `json:"IsStatePartArm"`
		StateArmDisArmDateTime     string  `json:"StateArmDisArmDateTime"`
	}
)

func (i GetSitesInput) validate() error {
	if i.Id < 1 {
		return errors.New("неверно задан номер объекта")
	}

	if i.ApiKey == "" {
		return errors.New("неверно задан API ключ")
	}

	return nil
}

func (i GetSitesInput) generateRequest() request {
	baseURL, _ := url.Parse(host + endpointGetSites)
	param := url.Values{}
	param.Add("id", strconv.Itoa(i.Id))
	if i.UserName != "" {
		param.Add("userName", i.UserName)
	}
	baseURL.RawQuery = param.Encode()

	return request{
		URL:    baseURL.String(),
		body:   []byte{},
		apiKey: i.ApiKey,
	}
}

type Client struct {
	client *http.Client
}

func NewClient() (*Client, error) {
	return &Client{
		client: &http.Client{Timeout: defaultTimeout},
	}, nil
}

func (c *Client) GetSites(ctx context.Context, input GetSitesInput) (GetSitesResponse, error) {
	if err := input.validate(); err != nil {
		return GetSitesResponse{}, err
	}

	req := input.generateRequest()
	body, err := c.doHTTP(ctx, req)
	if err != nil {
		return GetSitesResponse{}, err
	}

	var resp GetSitesResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return GetSitesResponse{}, errors.WithMessage(err, "Не удалось парсить ответ")
	}

	return resp, nil
}

func (c *Client) doHTTP(ctx context.Context, r request) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL, bytes.NewBuffer(r.body))
	if err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалось создать запрос")
	}

	req.Header.Set("apiKey", r.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалось выполнить запрос")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return []byte{}, errors.WithMessage(err, "Не удалось выполнить запрос")
		}
		err400 := RespErr400{}
		err = json.Unmarshal(buf.Bytes(), &err400)
		if err != nil {
			return []byte{}, err
		}
		return []byte{}, errors.New(err400.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return []byte{}, errors.New("Не удалось выполнить запрос")
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return []byte{}, errors.WithMessage(err, "Не удалcя парсинг ответа")
	}

	return buf.Bytes(), nil
}
