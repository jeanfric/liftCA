package ht

import (
	"encoding/json"
	"fmt"
	"io"
	"github.com/jeanfric/liftca"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	replyTypeError = iota
	replyTypeNotFound
	replyTypeJSON
	replyTypeRedirect
	replyTypeReader
	replyTypeNoContent
)

type Request struct {
	httpRequest *http.Request
}

type Answer struct {
	replyType   int
	data        interface{}
	contentType string
}

type Handler struct {
	f     func(*liftca.Store, *Request) *Answer
	store *liftca.Store
	log   logger
}

func NewHandler(store *liftca.Store, f func(*liftca.Store, *Request) *Answer) *Handler {
	return &Handler{
		f:     f,
		store: store,
		log:   &SimpleLogger{},
	}
}

func RedirectTo(url string) *Answer {
	return &Answer{
		replyType: replyTypeRedirect,
		data:      url,
	}
}

func JSONDocument(x interface{}) *Answer {
	return &Answer{
		replyType: replyTypeJSON,
		data:      x,
	}
}

func Failure(x interface{}) *Answer {
	return &Answer{
		replyType: replyTypeError,
		data:      x,
	}
}

func NoContent() *Answer {
	return &Answer{
		replyType: replyTypeNoContent,
	}
}

func NotFound() *Answer {
	return &Answer{
		replyType: replyTypeNotFound,
	}
}

func Read(contentType string, reader io.Reader) *Answer {
	return &Answer{
		replyType:   replyTypeReader,
		data:        reader,
		contentType: contentType,
	}
}

func (r *Request) BodyAsJSON(to interface{}) error {
	dec := json.NewDecoder(r.httpRequest.Body)
	if err := dec.Decode(to); err != nil {
		return err
	}
	return nil
}

func (r *Request) VarInt64(key string) (int64, error) {
	vars := mux.Vars(r.httpRequest)
	val, found := vars[key]
	if !found {
		return 0, fmt.Errorf("Not found")
	}
	return strconv.ParseInt(val, 10, 64)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()

	sw := NewDetailedResponseWriter(w)

	req := Request{
		httpRequest: r,
	}

	reply := h.f(h.store, &req)

	switch reply.replyType {
	case replyTypeRedirect:
		http.Redirect(sw, r, reply.data.(string), http.StatusFound)
	case replyTypeJSON:
		replyJSON(reply.data, sw)
	case replyTypeError:
		replyError(reply.data.(error), sw)
	case replyTypeReader:
		replyReader(reply.data.(io.Reader), reply.contentType, sw)
	case replyTypeNotFound:
		replyNotFound(r, sw)
	case replyTypeNoContent:
		replyNoContent(w)
	default:
		replyError(fmt.Errorf("Incorrect response handling"), sw)
	}

	h.log.Log(time.Since(timeStart), sw, r)

	return
}

func replyError(err error, w http.ResponseWriter) {
	log.Print(err)
	http.Error(w, "Server Error", http.StatusInternalServerError)
}

func replyNoContent(w http.ResponseWriter) {
	http.Error(w, "No Content", http.StatusNoContent)
}

func replyNotFound(r *http.Request, w http.ResponseWriter) {
	http.NotFound(w, r)
}

func replyJSON(reply interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reply); err != nil {
		replyError(err, w)
		return
	}
}

func replyReader(reader io.Reader, contentType string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentType)
	io.Copy(w, reader)
}
