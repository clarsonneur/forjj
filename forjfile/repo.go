package forjfile

import (
	"github.com/forj-oss/goforjj"
	"forjj/drivers"
	"github.com/forj-oss/forjj-modules/trace"
	"strings"
	"fmt"
)

type RepoStruct struct {
	name         string
	is_infra     bool
	infra        *InfraStruct
	forge        *ForgeYaml
	owner        string
	driverOwner  *drivers.Driver
	Upstream     string `yaml:"upstream-app,omitempty"` // Name of the application upstream hosting this repository.
	GitRemote    string `yaml:"git-remote,omitempty"`
	remote       goforjj.PluginRepoRemoteUrl // Git remote string to use/set
	Title        string `yaml:",omitempty"`
	RepoTemplate string `yaml:"repo-template,omitempty"`
	Flow         RepoFlow `yaml:",omitempty"`
	More         map[string]string `yaml:",inline"`
	apps         map[string]*AppStruct // List of applications connected to this repo. Defaults are added automatically.
	Apps         map[string]string `yaml:"in-relation-with"`// key: <AppRelName>, value: <appName>
}

type RepoFlow struct {
	Name string
	objects map[string]map[string]string
}

func (r *RepoStruct)Model() RepoModel {
	model := RepoModel{
		repo: r,
		Apps: make(map[string]RepoAppModel),
	}

	if r == nil {
		return model
	}

	for appRelName, app := range r.apps {
		repoModel := RepoAppModel{
			AppName: app.name,
		}
		if _, found := r.Apps[appRelName] ; !found {
			repoModel.Default = true
		}
		model.Apps[appRelName] = repoModel
	}
	return model
}

func (r *RepoStruct)Owner() string {
	if r == nil {
		return ""
	}

	return r.owner
}

func (r *RepoStruct)setFromInfra(infra *InfraStruct) {
	if r == nil {
		return
	}
	*r = infra.RepoStruct
	delete(r.More, "name")
	r.is_infra = true
	r.infra = infra
}

// setToInfra Copy each repo information to the infra, if each related infra data is not set.
func (r *RepoStruct)setToInfra(infra *InfraStruct) {
	if r == nil {
		return
	}

	if infra.Title == "" {
		infra.Title = r.Title
	}
	if infra.Upstream == "" {
		infra.Upstream = r.Upstream
	}
	if infra.GitRemote == "" {
		infra.GitRemote = r.GitRemote
	}
	if infra.RepoTemplate== "" {
		infra.RepoTemplate = r.RepoTemplate
	}
	if infra.Flow.Name == "" {
		infra.Flow.Name = r.Flow.Name
	}
	for app_name, app := range r.Apps {
		if v, found := infra.Apps[app_name] ; ! found || v == "" {
			infra.Apps[app_name] = app
		}
	}
	for key, value := range r.More {
		if v, found := infra.More[key]; ! found || v == "" {
			infra.Apps[key] = value
		}
	}
}

func (r *RepoStruct)GetString(field string) (string) {
	if r == nil {
		return ""
	}

	if v, found := r.Get(field) ; found {
		return v.GetString()
	}
	return ""
}

func (r *RepoStruct)RemoteUrl() string {
	if r == nil {
		return ""
	}

	return r.remote.Url
}

func (r *RepoStruct)RemoteGit() string {
	if r == nil {
		return ""
	}

	return r.remote.Ssh
}

func (r *RepoStruct)Get(field string) (value *goforjj.ValueStruct, _ bool) {
	if r == nil {
		return
	}
	switch field_sel := strings.Split(field, ":") ; field_sel[0] {
	case "name":
		return value.SetIfFound(r.name, (r.name != ""))
	case "apps", "upstream":
		if field == "upstream" {
			gotrace.Warning("*RepoStruct.Get(): Field '%s' is obsolete. Change the code to use 'apps:upstream'.", field)
		} else if len(field_sel) >1 {
			field = field_sel[1]
		}
		if v, found := r.apps[field] ; found {
			return value.SetIfFound(v.name, v != nil && v.name != "")
		}
		if v, found := r.Apps[field] ; found {
			return value.SetIfFound(v, found && (v != ""))
		}
		if field == "upstream" && r.Upstream != "" {
			gotrace.Warning("the '%s' in /repositories/%s is obsolete. Define it as /repositories/%s/apps/%s", field,
				r.name, r.name, field)
			return value.SetIfFound(r.Upstream, (r.Upstream != ""))
		}
		return value.SetIfFound("", false)
	case "git-remote":
		return value.SetIfFound(r.GitRemote, (r.GitRemote != ""))
	case "remote":
		return value.SetIfFound(r.remote.Ssh, (r.remote.Ssh != ""))
	case "remote-url":
		return value.SetIfFound(r.remote.Url, (r.remote.Url != ""))
	case "title":
		return value.SetIfFound(r.Title, (r.Title != ""))
	case "flow":
		return value.SetIfFound(r.Flow.Name, (r.Flow.Name != ""))
	case "repo-template":
		return value.SetIfFound(r.RepoTemplate, (r.RepoTemplate != ""))
	default:
		v, f := r.More[field]
		return value.SetIfFound(v, f)
	}
	return
}

func (r *RepoStruct)SetHandler(from func(field string)(string, bool), keys...string) {
	if r == nil {
		return
	}

	for _, key := range keys {
		if v, found := from(key) ; found {
			r.Set(key, v)
		}
	}
}

// SetInternalRelApp will set the appName connected to the repo
// But if the Forjfile has a setup, this one will forcefully used.
// return:
// updated bool : true if the app has been updated.
// error : set if error has been found. updated is then nil.
func (r *RepoStruct)SetInternalRelApp(appRelName, appName string) (updated *bool, _ error) {
	if err := r.initApp() ; err != nil {
		return nil, err
	}

	if v, found := r.Apps[appRelName] ; found {
		appName = v // Set always declared one.
	}

	if app, err := r.forge.Apps.Found(appName) ; err != nil {
		return nil, fmt.Errorf("Unable to set %s:%s. %s.", appRelName, appName, err)
	} else if v, found :=  r.apps[appRelName] ; !found || (found && v.name != appName) {
		r.apps[appRelName] = app
		updated = new(bool)
		*updated = true
	}

	return
}

// SetApp Define the Forjfile application name to link with the repo.
//
// return:
// updated bool : true if the app has been updated.
// error : set if error has been found. updated is then nil.
func (r *RepoStruct)SetApp(appRelName, appName string) (updated *bool, _ error) {
	if err := r.initApp() ; err != nil {
		return nil, err
	}

	updated = new(bool)
	if v, found := r.Apps[appRelName] ; !found || (found && v != appName) {
		*updated = true
	}
	r.Apps[appRelName] = appName

	if set, err := r.SetInternalRelApp(appRelName, appName) ; set == nil {
		return set, err
	}
	return
}

func (r *RepoStruct)initApp() (_ error) {
	if r == nil {
		return fmt.Errorf("Internal: repo object is nil.")
	}
	if r.forge == nil {
		return fmt.Errorf("Internal: forge ref not set.")
	}

	if r.apps == nil {
		r.apps = make(map[string]*AppStruct)
	}
	if r.Apps == nil {
		r.Apps = make(map[string]string)
	}
	return
}

func (r *RepoStruct)Set(field, value string) {
	if r == nil {
		return
	}
	switch field_sel := strings.Split(field, ":") ; field_sel[0] {
	case "name":
		if r.name != value {
			r.name = value
		}
	case "apps":
		if len(field_sel) > 1{
			r.SetApp(field_sel[1], value)
			r.forge.dirty()
		}
	case "upstream":
		if v, found := r.Apps["upstream"] ; !found || (found && v != value) {
			r.SetApp("upstream", value)
			r.forge.dirty()
		}
	case "git-remote":
		if r.GitRemote != value {
			r.GitRemote = value
			r.forge.dirty()
		}
	case "remote":
		if r.remote.Ssh != value {
			r.remote.Ssh = value
		}
	case "remote-url":
		if r.remote.Url != value {
			r.remote.Url = value
		}
	case "repo-template":
		if r.RepoTemplate != value {
			r.RepoTemplate = value
			r.forge.dirty()
		}
	case "title":
		if r.Title != value {
			r.Title = value
			r.forge.dirty()
		}
	case "flow":
		if r.Flow.Name != value {
			r.Flow.Name = value
			r.forge.dirty()
		}
	default:
		if r.More == nil {
			r.More = make(map[string]string)
		}
		if v, found := r.More[field] ; found && value == "" {
			delete(r.More,field)
			r.forge.dirty()
		} else {
			if v != value {
				r.forge.dirty()
				r.More[field] = value
			}

		}
	}
}

func (r *RepoStruct)SetInstanceOwner(owner string) {
	if r == nil {
		return
	}
	r.owner = owner
}

func (r *RepoStruct)SetPluginOwner(d *drivers.Driver) {
	if r == nil {
		return
	}
	r.driverOwner = d
}

func (r *RepoStruct)RemoteType() string {
	if r == nil {
		return "git"
	}
	if r.driverOwner == nil {
		return "git"
	}
	return r.driverOwner.Name
}

func (r *RepoStruct)UpstreamAPIUrl() string {
	if r == nil {
		return ""
	}
	if r.driverOwner == nil {
		return ""
	}
	return r.driverOwner.DriverAPIUrl
}

func (r *RepoStruct)set_forge(f *ForgeYaml) {
	if r == nil {
		return
	}

	r.forge = f
}

// HasApps return a bool if rules are all true on at least one application.
// a rule is a string formatted as '<key>:<value>'
// a rule is true on an application if it has the key value set to <value>
//
// If the rule is not well formatted, an error is returned.
// If the repo has no application, HasApps return false.
// If no rules are provided and at least one application exist, HasApps return true.
//
// TODO: Write Unit test of HasApps
func (r *RepoStruct)HasApps(rules ...string) (found bool, err error) {
	if r.apps == nil {
		return
	}
	for appRelName, app:= range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if ruleToCheck[0] == "appRelName" {
			    if appRelName == ruleToCheck[0] {
					continue
				}
				found = false
				break
			}
			v, found2 := app.Get(ruleToCheck[0])
			if ruleToCheck[1] == "*" {
				if found2 {
					continue
				}
				found = false
				break
			} else if found2 && v.GetString() != ruleToCheck[1] {
				found = false
				break
			}
		}
		if found {
			gotrace.Trace("Found an application which meets '%s'", rules)
			return
		}
	}
	gotrace.Trace("NO application found which meets '%s'", rules)
	return
}

func (r *RepoStruct)GetApps(rules ...string) (apps map[string]*AppStruct , err error) {
	if r.apps == nil {
		return
	}
	apps = make(map[string]*AppStruct)
	found := false
	for app_name, app:= range r.apps {
		found = true
		for _, rule := range rules {
			ruleToCheck := strings.Split(rule, ":")
			if len(ruleToCheck) != 2 {
				err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
				return
			}
			if ruleToCheck[0] == "appRelName" {
				if app_name == ruleToCheck[0] {
					continue
				}
				found = false
				break
			}
			v, found2 := app.Get(ruleToCheck[0])
			if ruleToCheck[1] == "*" {
				if found2 {
					continue
				}
				found = false
				break
			} else if found2 && v.GetString() != ruleToCheck[1] {
				found = false
				break
			}
		}
		if found {
			apps[app_name] = app
		}
	}
	return
}

func (r *RepoStruct)HasValues(rules ...string) (found bool, err error) {
	found = true
	for _, rule := range rules {
		ruleToCheck := strings.Split(rule, ":")
		if len(ruleToCheck) != 2 {
			err = fmt.Errorf("rule '%s' is invalid. Format supported is '<key>:<value>'.", rule)
			return
		}
		if v, found2 := r.Get(ruleToCheck[0]); found2 && v.GetString() != ruleToCheck[1] {
			found = false
			break
		}
	}
	if found {
		return
	}
	return
}
