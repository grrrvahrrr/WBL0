package api

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"wbl0/recieveMsg/internal/entities"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func TestGetOrder(t *testing.T) {
	testid := "testid"

	r := chi.NewRouter()

	hm := &HandlerMock{}

	ret := &ApiChiMock{
		hm: hm,
	}
	r.Post("/order", ret.GetOrder)

	ret.Mux = r

	form := url.Values{}
	form.Add("orderid", testid)

	req, err := http.NewRequest("POST", "/order", strings.NewReader(form.Encode()))
	if err != nil {
		t.Errorf("failed to create an http request: %v", err)
		return
	}

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	ret.Mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected code: %d, got: %d", http.StatusOK, rr.Code)
	}

}

type ApiChiMock struct {
	*chi.Mux
	hm *HandlerMock
}

func (rt *ApiChiMock) GetOrder(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("error parsing form:", err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}

	orderId := r.Form.Get("orderid")
	if orderId == "" {
		log.Println("search query empty")
		err := render.Render(w, r, ErrNotFound)
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}

	order, err := rt.hm.GetOrderByIdHandler(r.Context(), orderId)
	if err != nil {
		log.Println(err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}

	t, err := template.ParseFiles("pages/order.html")
	if err != nil {
		log.Println("template parsing error: ", err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}
	err = t.Execute(w, *order)
	if err != nil {
		log.Println("template executing error: ", err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}

}

type HandlerMock struct {
}

func (hm *HandlerMock) GetOrderByIdHandler(ctx context.Context, orderId string) (*entities.Order, error) {
	testdelivery := entities.Delivery{
		Name:    "testname",
		Phone:   "testphone",
		Zip:     "testzip",
		City:    "testcity",
		Address: "testaddress",
		Region:  "testregion",
		Email:   "testemail",
	}

	testpayment := entities.Payment{
		Transaction:  "testid",
		RequestID:    "testreqid",
		Currency:     "cur",
		Provider:     "testprv",
		Amount:       1,
		PaymentDt:    2,
		Bank:         "testbank",
		DeliveryCost: 3,
		GoodsTotal:   4,
		CustomFee:    5,
	}

	testcartitem := entities.CartItem{
		ChrtId:      0,
		TrackNumber: "testtracknum",
		Price:       1,
		Rid:         "testrid",
		Name:        "testname",
		Sale:        2,
		Size:        "testsize",
		TotalPrice:  3,
		NmId:        4,
		Brand:       "testbrand",
		Status:      5,
	}

	var testitems []entities.CartItem
	testitems = append(testitems, testcartitem)

	return &entities.Order{
		OrderId:           "testid",
		TrackNumber:       "testnumber",
		Entry:             "testentry",
		Delivery:          testdelivery,
		Payment:           testpayment,
		Items:             testitems,
		Locale:            "testlocale",
		InternalSignature: "testinternalsig",
		CustomerId:        "testcustomerid",
		DeliveryService:   "testdelser",
		ShardKey:          "testshardkey",
		SmID:              0,
		DateCreated:       "testdate",
		OofShard:          "testoof",
	}, nil

}
