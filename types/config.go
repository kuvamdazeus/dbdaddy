package types

type CliConfig struct {
	MainConnConfig   ConnConfig
	ShadowConnConfig *ConnConfig
	Origins          map[string]ConnConfig

	// env vars file
	EnvVars CliConfigVars

	// state file
	State CliState
}

func NewCliConfig() CliConfig {
	return CliConfig{
		MainConnConfig: ConnConfig{
			Params: map[string]string{},
		},
		EnvVars: CliConfigVars{},
		State:   CliState{},
		Origins: map[string]ConnConfig{},
	}
}

type CliConfigVars map[string]string

type CliState struct {
	CurrentBranch string `json:"current_branch"`
}
