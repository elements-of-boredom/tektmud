package configs

type Paths struct {
	RootDataDir  string `yaml:"root_data_dir"`
	UserData     string `yaml:"user_data"`
	WorldFiles   string `yaml:"world_files"`
	Localization string `yaml:"localization"`
	Logs         string `yaml:"logs"`
	Templates    string `yaml:"templates"`
	Races        string `yaml:"races"`
	Classes      string `yaml:"classes"`
}

func (p *Paths) Check() {

	if p.UserData == `` {
		p.UserData = `user_data`
	}

	if p.WorldFiles == `` {
		p.WorldFiles = `world_files`
	}

	if p.Localization == `` {
		p.Localization = `localization`
	}

	if p.Races == `` {
		p.Races = `races`
	}

	if p.Classes == `` {
		p.Classes = `classes`
	}

	if p.Logs == `` {
		p.Logs = `logs`
	}
}
