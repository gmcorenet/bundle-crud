package service

import (
	"encoding/json"
	"net/http"
)

type DesignerService struct{}

func NewDesignerService() *DesignerService {
	return &DesignerService{}
}

func (d *DesignerService) Meta(resource string, mode, owner interface{}) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"ok":   true,
		"mode": mode,
	}
}

func (d *DesignerService) Save(resource string, mode, owner interface{}, payload json.RawMessage) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Designer layout saved",
	}
}

func (d *DesignerService) Reset(resource string, mode, owner interface{}) (int, interface{}) {
	return http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Designer layout reset",
	}
}
