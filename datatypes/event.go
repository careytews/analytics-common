package datatypes

// DNS request
type DnsQuery struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Class string `json:"class,omitempty"`
}

// DNS response
type DnsAnswer struct {
	// Name
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Class string `json:"class,omitempty"`

	// IP address
	Address string `json:"address,omitempty"`
}

// Geo
type Posn struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type Place struct {
	City    string `json:"city,omitempty"`
	IsoCode string `json:"iso,omitempty"`
	Country string `json:"country,omitempty"`
	ASNum   uint   `json:"asnum,omitempty"`
	ASOrg   string `json:"asorg,omitempty"`

	Position       *Posn  `json:"position,omitempty"`
	AccuracyRadius int    `json:"accuracy,omitempty"`
	PostCode       string `json:"postcode,omitempty"`
}

type LocationInfo struct {
	Src  *Place `json:"src,omitempty"`
	Dest *Place `json:"dest,omitempty"`
}

type UnrecDg struct {
	Payload          string `json:"payload,omitempty"`
	PayloadLength    int    `json:"payload-length,omitempty"`
	PayloadB64Length int    `json:"payload-b64length,omitempty"`
	PayloadHash      string `json:"payload-sha1,omitempty"`
}

type UnrecStrm struct {
	Payload          string `json:"payload,omitempty"`
	PayloadLength    int    `json:"payload-length,omitempty"`
	PayloadB64Length int    `json:"payload-b64length,omitempty"`
	PayloadHash      string `json:"payload-sha1,omitempty"`
	Position         int64  `json:"position"`
}

type Icmp struct {
	Type    int    `json:"type"`
	Code    int    `json:"code"`
	Payload string `json:"payload,omitempty"`
}

type HttpRequest struct {
	Method string            `json:"method,omitempty"`
	Header map[string]string `json:"header,omitempty"`
	Body   string            `json:"body,omitempty"`
}

type HttpResponse struct {
	Code   int               `json:"code"`
	Status string            `json:"status,omitempty"`
	Header map[string]string `json:"header,omitempty"`
	Body   string            `json:"body,omitempty"`
}

type DnsMessage struct {
	Type   string      `json:"type"`
	Query  []DnsQuery  `json:"query,omitempty"`
	Answer []DnsAnswer `json:"answer,omitempty"`
}

type FtpCommand struct {
	Command string `json:"command,omitempty"`
}

type FtpResponse struct {
	Status int      `json:"status"`
	Text   []string `json:"text,omitempty"`
}

type SipRequest struct {
	Method  string `json:"method,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type SipResponse struct {
	Code    int    `json:"code"`
	Status  string `json:"status,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type Payload struct {
	Payload string `json:"payload,omitempty"`
}

type SmtpCommand struct {
	Command string `json:"command,omitempty"`
}

type SmtpResponse struct {
	Status int      `json:"status,omitempty"`
	Text   []string `json:"text,omitempty"`
}

type SmtpData struct {
	From string   `json:"from,omitempty"`
	To   []string `json:"to,omitempty"`
	Data string   `json:"data,omitempty"`
}

type Ntp struct {
	Version int `json:"version"`
	Mode    int `json:"mode"`
}

type Operation struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

// Network event
type Event struct {
	Id      string `json:"id,omitempty"`
	Action  string `json:"action,omitempty"`
	Device  string `json:"device,omitempty"`
	Network string `json:"network,omitempty"`
	Time    string `json:"time,omitempty"`
	Origin  string `json:"origin,omitempty"`

	UnrecognisedDatagram *UnrecDg   `json:"unrecognised_datagram,omitempty"`
	UnrecognisedStream   *UnrecStrm `json:"unrecognised_stream,omitempty"`

	Icmp         *Icmp         `json:"icmp,omitempty"`
	HttpRequest  *HttpRequest  `json:"http_request,omitempty"`
	HttpResponse *HttpResponse `json:"http_response,omitempty"`
	DnsMessage   *DnsMessage   `json:"dns_message,omitempty"`
	FtpCommand   *FtpCommand   `json:"ftp_command,omitempty"`
	FtpResponse  *FtpResponse  `json:"ftp_response,omitempty"`
	SipRequest   *SipRequest   `json:"sip_request,omitempty"`
	SipResponse  *SipResponse  `json:"sip_response,omitempty"`
	SipSsl       *Payload      `json:"sip_ssl,omitempty"`
	Imap         *Payload      `json:"imap,omitempty"`
	ImapSsl      *Payload      `json:"imap_ssl,omitempty"`
	Pop3         *Payload      `json:"pop3,omitempty"`
	Pop3Ssl      *Payload      `json:"pop3_ssl,omitempty"`
	SmtpCommand  *SmtpCommand  `json:"smtp_command,omitempty"`
	SmtpResponse *SmtpResponse `json:"smtp_response,omitempty"`
	SmtpData     *SmtpData     `json:"smtp_data,omitempty"`
	NtpTimestamp *Ntp          `json:"ntp_timestamp,omitempty"`
	NtpControl   *Ntp          `json:"ntp_control,omitempty"`
	NtpPrivate   *Ntp          `json:"ntp_private,omitempty"`

	Url string `json:"url,omitempty"`

	Src  []string `json:"src,omitempty"`
	Dest []string `json:"dest,omitempty"`

	Location   *LocationInfo `json:"location,omitempty"`
	Indicators *[]*Indicator `json:"indicators,omitempty"`

	Risk float64 `json:"risk"`

	// map for operation name -> k,v pairs
	Operations *[]*Operation `json:"operations,omitempty"`
}
