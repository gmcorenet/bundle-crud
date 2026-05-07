package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	gmcore_crud "github.com/gmcorenet/sdk-gmcore-crud"
	"github.com/gmcorenet/bundle-crud/internal/export"
	"github.com/gmcorenet/bundle-crud/internal/filter"
	"github.com/gmcorenet/bundle-crud/internal/model"
	"github.com/gmcorenet/bundle-crud/internal/registry"
	"github.com/gmcorenet/bundle-crud/internal/service"
)

var (
	reg     *registry.CrudRegistry
	defs    *model.ConfigDefaults
	filters *filter.FilterTypeRegistry
)

func SetRegistry(r *registry.CrudRegistry) { reg = r }
func SetDefaults(d *model.ConfigDefaults) { defs = d }
func SetFilterRegistry(f *filter.FilterTypeRegistry) { filters = f }

type JSONResponse struct {
	Ok      bool                   `json:"ok"`
	Message string                 `json:"message,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type AjaxInspector interface {
	IsAjaxRequest(*http.Request, string) bool
}

type DefaultAjaxInspector struct{}

func (d *DefaultAjaxInspector) IsAjaxRequest(r *http.Request, marker string) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		r.Header.Get("X-CRUD-Ajax") != "" ||
		r.Header.Get("Content-Type") == "application/json"
}

type CrudRouter struct {
	mux       *http.ServeMux
	ajaxCheck AjaxInspector
}

func NewCrudRouter(
	crudReg *registry.CrudRegistry,
	lists *service.ListResponder,
	forms *service.FormResponder,
	bulkService *service.BulkActionService,
	serverActionService *service.ServerActionService,
	relationService *service.RelationOptionsService,
	designerService *service.DesignerService,
	diagnosticsService *service.DiagnosticsService,
	defaults *model.ConfigDefaults,
	filterReg *filter.FilterTypeRegistry,
) *CrudRouter {
	c := &CrudRouter{
		mux:       http.NewServeMux(),
		ajaxCheck: &DefaultAjaxInspector{},
	}

	reg = crudReg
	defs = defaults
	filters = filterReg

	c.mux.HandleFunc("/_gmcore/crud/", c.routeDispatch)

	return c
}

func (c *CrudRouter) Register(mux *http.ServeMux) {
	mux.Handle("/_gmcore/crud/", c.mux)
}

func (c *CrudRouter) routeDispatch(w http.ResponseWriter, r *http.Request) {
	if !c.ajaxCheck.IsAjaxRequest(r, "_crud_ajax") {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/_gmcore/crud/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	dispatch := func(resource string, id string, extra []string) {
		switch {
		case len(extra) == 0 && id == "":
			c.handleIndex(resource, w, r)
		case len(extra) == 0 && id != "":
			c.handleView(resource, id, w, r)
		}
	}

	_ = dispatch

	switch {
	case len(segments) == 1:
		c.handleIndex(segments[0], w, r)
	case len(segments) == 2:
		switch segments[1] {
		case "filters":
			c.handleFilters(segments[0], w, r)
		case "debug":
			c.handleDebug(segments[0], w, r)
		case "designer":
			c.handleDesigner(segments[0], w, r)
		case "bulk":
			c.handleBulk(segments[0], w, r)
		case "new":
			c.handleCreate(segments[0], w, r)
		default:
			c.handleView(segments[0], segments[1], w, r)
		}
	case len(segments) == 3:
		switch {
		case segments[1] == "designer" && segments[2] == "modal":
			c.handleDesigner(segments[0], w, r)
		case segments[1] == "designer":
			c.handleDesignerAPI(segments[0], segments[2], w, r)
		case segments[1] == "filters" && segments[2] == "modal":
			c.handleFilters(segments[0], w, r)
		case segments[1] == "debug" && segments[2] == "modal":
			c.handleDebug(segments[0], w, r)
		case segments[2] == "new":
			c.handleCreate(segments[0], w, r)
		case segments[1] == "relations":
			c.handleRelationOptions(segments[0], segments[2], w, r)
		default:
			http.NotFound(w, r)
		}
	case len(segments) == 4:
		switch segments[3] {
		case "edit":
			c.handleEdit(segments[0], segments[2], w, r)
		case "clone":
			c.handleClone(segments[0], segments[2], w, r)
		case "delete":
			c.handleDelete(segments[0], segments[2], w, r)
		default:
			http.NotFound(w, r)
		}
	case len(segments) == 5 && segments[3] == "action":
		c.handleServerAction(segments[0], segments[2], segments[4], w, r)
	case len(segments) == 6 && segments[3] == "action" && segments[5] == "form":
		c.handleServerActionForm(segments[0], segments[2], segments[4], w, r)
	default:
		http.NotFound(w, r)
	}
}

func (c *CrudRouter) denyAccess(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusForbidden, JSONResponse{Ok: false, Message: msg, Type: "error"})
}

func (c *CrudRouter) denyNotFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, JSONResponse{Ok: false, Message: msg, Type: "error"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (c *CrudRouter) handleIndex(resource string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	params := parseListParams(r)
	records, err := crud.List(params, getUser(r))
	if err != nil {
		c.denyAccess(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"records": records,
		"params":  params,
	})
}

func (c *CrudRouter) handleView(resource, id string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	record, err := crud.Find(id, getUser(r))
	if err != nil {
		c.denyNotFound(w, "Record not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"record": record,
	})
}

func (c *CrudRouter) handleCreate(resource string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)

	record, err := crud.Create(payload, getUser(r))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, JSONResponse{Ok: false, Message: err.Error(), Type: "error"})
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "crud.flash.created",
		Type:    "success",
		Payload: map[string]interface{}{"record": record},
	})
}

func (c *CrudRouter) handleEdit(resource, id string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	record, err := crud.Find(id, getUser(r))
	if err != nil {
		c.denyNotFound(w, "Record not found")
		return
	}

	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)

	err = crud.Update(record, payload, getUser(r))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, JSONResponse{Ok: false, Message: err.Error(), Type: "error"})
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "crud.flash.updated",
		Type:    "success",
	})
}

func (c *CrudRouter) handleClone(resource, id string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	record, err := crud.Find(id, getUser(r))
	if err != nil {
		c.denyNotFound(w, "Record not found")
		return
	}

	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)

	newRecord, err := crud.Create(payload, getUser(r))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, JSONResponse{Ok: false, Message: err.Error(), Type: "error"})
		return
	}

	_ = record

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "crud.flash.created",
		Type:    "success",
		Payload: map[string]interface{}{"record": newRecord},
	})
}

func (c *CrudRouter) handleDelete(resource, id string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	record, err := crud.Find(id, getUser(r))
	if err != nil {
		c.denyNotFound(w, "Record not found")
		return
	}

	err = crud.Delete(record, getUser(r))
	if err != nil {
		c.denyAccess(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "crud.flash.deleted",
		Type:    "success",
	})
}

func (c *CrudRouter) handleBulk(resource string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	var payload struct {
		Action string   `json:"action"`
		IDs    []string `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&payload)

	if len(payload.IDs) == 0 || payload.Action == "" {
		writeJSON(w, http.StatusOK, JSONResponse{Ok: true, Message: "crud.flash.bulk_empty", Type: "info"})
		return
	}

	result, err := crud.Bulk(payload.Action, payload.IDs, getUser(r))
	if err != nil {
		c.denyAccess(w, err.Error())
		return
	}

	if strings.HasPrefix(payload.Action, "export_") {
		format := strings.TrimPrefix(payload.Action, "export_")
		exporter, ok := export.GlobalRegistry().Get(format)
		if ok {
			content := fmt.Sprintf("%v", result)
			exportResult, err := exporter.Export(export.ExportRequest{
				Resource: resource,
				Action:   payload.Action,
				Content:  content,
			})
			if err == nil {
				w.Header().Set("Content-Type", exportResult.ContentType)
				w.Header().Set("Content-Disposition", "attachment; filename=\""+exportResult.Filename+"\"")
				w.Write(exportResult.Data)
				return
			}
		}
		writeJSON(w, http.StatusOK, JSONResponse{
			Ok:      true,
			Message: "crud.flash.bulk_done",
			Type:    "success",
			Payload: map[string]interface{}{"result": fmt.Sprintf("%v", result)},
		})
		return
	}

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "crud.flash.bulk_done",
		Type:    "success",
	})
}

func (c *CrudRouter) handleFilters(resource string, w http.ResponseWriter, r *http.Request) {
	_, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"filters": []string{},
	})
}

func (c *CrudRouter) handleDebug(resource string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	diagnostics := map[string]interface{}{
		"name":       resource,
		"fields":     len(crud.Config().Fields),
		"relations":  len(crud.Config().Relations),
		"registered": true,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"diagnostics": diagnostics,
	})
}

func (c *CrudRouter) handleDesigner(resource string, w http.ResponseWriter, r *http.Request) {
	_, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"mode": "create",
	})
}

func (c *CrudRouter) handleDesignerAPI(resource, mode string, w http.ResponseWriter, r *http.Request) {
	_, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":   true,
			"meta": map[string]interface{}{"mode": mode},
		})
	case http.MethodPost:
		writeJSON(w, http.StatusOK, JSONResponse{
			Ok:      true,
			Message: "Designer saved",
			Type:    "success",
		})
	case http.MethodDelete:
		writeJSON(w, http.StatusOK, JSONResponse{
			Ok:      true,
			Message: "Designer reset",
			Type:    "success",
		})
	default:
		c.denyNotFound(w, "Method not allowed")
	}
}

func (c *CrudRouter) handleRelationOptions(resource, relation string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	options, hasMore, err := crud.RelationOptions(relation, query, page, limit)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"items":    options,
		"has_more": hasMore,
	})
}

func (c *CrudRouter) handleServerAction(resource, id, action string, w http.ResponseWriter, r *http.Request) {
	crud, err := reg.Get(resource)
	if err != nil {
		c.denyNotFound(w, err.Error())
		return
	}

	record, err := crud.Find(id, getUser(r))
	if err != nil {
		c.denyNotFound(w, "Record not found")
		return
	}

	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)

	result, err := crud.Bulk(action, []string{id}, getUser(r))
	if err != nil {
		c.denyAccess(w, err.Error())
		return
	}

	_ = record
	_ = payload

	writeJSON(w, http.StatusOK, JSONResponse{
		Ok:      true,
		Message: "Action executed",
		Type:    "success",
		Payload: map[string]interface{}{"result": result},
	})
}

func (c *CrudRouter) handleServerActionForm(resource, id, action string, w http.ResponseWriter, r *http.Request) {
	c.handleServerAction(resource, id, action, w, r)
}

func getUser(r *http.Request) interface{} {
	return r.Context().Value("user")
}

func parseListParams(r *http.Request) gmcore_crud.ListParams {
	params := gmcore_crud.ListParams{
		Page:    1,
		PerPage: 25,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}
	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 {
			params.PerPage = pp
		}
	}
	if search := r.URL.Query().Get("q"); search != "" {
		params.Search = search
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = strings.Split(sort, ",")
	}

	return params
}
