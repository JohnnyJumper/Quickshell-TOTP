package store

type Account struct {
	ID        string `json:"id"`
	Issuer    string `json:"issuer"`
	Label     string `json:"label"`
	Secret    string `json:"secret"`
	Algorithm string `json:"algorithm"`
	Digits    int    `json:"digits"`
	Period    int    `json:"period"`
}
