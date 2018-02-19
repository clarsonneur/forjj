package forjfile

type InfraStruct struct {
	Maintain MaintainStruct
	RepoStruct
}

type MaintainStruct struct { // Would contains any kind of specific data required to manage forjj-maintain
	DataName string `yaml:"data-name"`
}
