package fatsecret

// State holds state related to FatSecret API
type State struct {
	AuthCallbackURL string
}

func StateInit(authCallbackURL string) *State {

	return &State{
		AuthCallbackURL: authCallbackURL,
	}
}
