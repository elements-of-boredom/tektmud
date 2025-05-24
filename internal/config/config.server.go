package configs

type Server struct {
	Name        string `yaml:"name"`
	Seed        string `yaml:"seed"`
	MaxCPUCores int    `yaml:"max_cpu_cores"`
	Ports       []int  `yaml:"ports"`
	MaxPlayers  int    `yaml:"max_players"`
	IdleTimeout int    `yaml:"idle_timeout_minutes"`
	LogLevel    string `yaml:"log_level"`
}

func (s *Server) Check() {
	if s.Seed == `` {
		s.Seed = `8008135`
	}

	if s.MaxCPUCores < 0 {
		s.MaxCPUCores = 0
	}
}
