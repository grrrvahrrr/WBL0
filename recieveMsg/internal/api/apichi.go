package api

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type ApiChi struct {
	*chi.Mux
	hs *Handlers
}

func NewApiChiRouter(hs *Handlers) *ApiChi {
	r := chi.NewRouter()

	ret := &ApiChi{
		hs: hs,
	}

	r.Get("/home", ret.Homepage)
	r.Post("/order", ret.GetOrder)

	ret.Mux = r

	return ret
}

// (GET /home)
func (rt *ApiChi) Homepage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("pages/homepage.html")
	if err != nil {
		log.Println("template parsing error: ", err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		log.Println("template execute error: ", err)
		err := render.Render(w, r, ErrRender(err))
		if err != nil {
			log.Println("api rendering error: ", err)
		}
		return
	}
}

//(Post /order)
func (rt *ApiChi) GetOrder(w http.ResponseWriter, r *http.Request) {
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

	order, err := rt.hs.GetOrderByIdHandler(r.Context(), orderId)
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
