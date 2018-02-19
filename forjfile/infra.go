package forjfile

import (
	"strings"
	"github.com/forj-oss/goforjj"
)

type InfraStruct struct {
	Maintain MaintainStruct
	RepoStruct
}

type MaintainStruct struct { // Would contains any kind of specific data required to manage forjj-maintain
	DataName string `yaml:"data-volume"`
}

func (r *InfraStruct)Get(field string) (value *goforjj.ValueStruct, _ bool) {
	if r == nil {
		return
	}
	switch field_sel := strings.Split(field, ":"); field_sel[0] {
	case "maintain-data-volume":
		return value.SetIfFound(r.Maintain.DataName, (r.Maintain.DataName != ""))
	default:
		return r.RepoStruct.Get(field)
	}
	return
}

func (r *InfraStruct)SetHandler(from func(field string) (string, bool), keys...string) {
	if r == nil {
		return
	}

	for _, key := range keys {
		if v, found := from(key); found {
			r.Set(key, v)
		}
	}
}

func (r *InfraStruct)Set(field, value string) {
	if r == nil {
		return
	}
	switch field_sel := strings.Split(field, ":"); field_sel[0] {
	case "maintain-data-volume":
		if r.Maintain.DataName != value {
			r.Maintain.DataName= value
			r.forge.dirty()
		}
	default:
		r.RepoStruct.Set(field, value)
	}
}
