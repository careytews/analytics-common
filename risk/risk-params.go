package risk

// Params are the parameters which go with one risk.
type Params struct {
	A     float64 `json:"a,omitempty"`
	B     float64 `json:"b,omitempty"`
	Risk  float64 `json:"risk,omitempty"`
	Decay float64 `json:"decay,omitempty"`
}

// Risks holds the categories of risk
var Risks = map[string]Params{
	"location.unexpected":     Params{A: 0.30, B: 0.70, Risk: 0.30, Decay: 86400},
	"lost.laptop":             Params{A: 0.30, B: 0.70, Risk: 0.90, Decay: 86400},
	"ntp.private":             Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"apt":                     Params{A: 0.30, B: 0.70, Risk: 0.70, Decay: 86400},
	"redirector":              Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"spyware":                 Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"test":                    Params{A: 0.30, B: 0.10, Risk: 0.00, Decay: 86400},
	"tor.entry":               Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"tor.exit":                Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"warez":                   Params{A: 0.30, B: 0.10, Risk: 0.30, Decay: 86400},
	"insecure":                Params{A: 0.30, B: 0.70, Risk: 0.30, Decay: 86400},
	"dynamic":                 Params{A: 0.30, B: 0.10, Risk: 0.30, Decay: 86400},
	"hacking":                 Params{A: 0.30, B: 0.10, Risk: 0.30, Decay: 86400},
	"anonvpn":                 Params{A: 0.30, B: 0.10, Risk: 0.70, Decay: 86400},
	"compromised-credentials": Params{A: 0.30, B: 0.70, Risk: 0.90, Decay: 86400},
	"covert.dns-tunnel":       Params{A: 0.30, B: 0.70, Risk: 0.90, Decay: 86400},
	"fraud.launder":           Params{A: 0.30, B: 0.70, Risk: 0.90, Decay: 86400},
	"fraud.watch":             Params{A: 0.30, B: 0.70, Risk: 0.90, Decay: 86400},
	"anomaly.useragent":       Params{A: 0.30, B: 0.70, Risk: 0.30, Decay: 86400},
	"anomaly.domain":          Params{A: 0.30, B: 0.70, Risk: 0.30, Decay: 86400},
	"anomaly.ja3":             Params{A: 0.30, B: 0.70, Risk: 0.30, Decay: 86400},
	"rat.dark-comet":          Params{A: 0.30, B: 0.70, Risk: 0.70, Decay: 86400},
	"exploit.eternal-blue":    Params{A: 0.30, B: 0.70, Risk: 0.70, Decay: 86400},
}

// Removed:
// 	"oppression.rights":       Params{A: 0.30, B: 0.70, Risk: 0.70, Decay: 86400},
// 	"porn":                    Params{A: 0.30, B: 0.10, Risk: 0.05, Decay: 86400},
// 	"sex.lingerie":            Params{A: 0.30, B: 0.10, Risk: 0.05, Decay: 86400},
// 	"violence":                Params{A: 0.30, B: 0.10, Risk: 0.05, Decay: 86400},
// 	"drugs":                   Params{A: 0.70, B: 0.10, Risk: 0.70, Decay: 86400},
// 	"gamble":                  Params{A: 0.30, B: 0.10, Risk: 0.05, Decay: 86400},
// 	"aggressive":              Params{A: 0.30, B: 0.10, Risk: 0.05, Decay: 86400},
