package character

type Race struct {
	Id      int    `yaml:"id"`
	Name    string `yaml:"name"`
	Force   int    `yaml:"force"`
	Reflex  int    `yaml:"reflex"`
	Acuity  int    `yaml:"acuity"`
	Insight int    `yaml:"insight"`
	Heart   int    `yaml:"heart"`
	BuffIds []int  `yaml:"buff_ids"`
}
