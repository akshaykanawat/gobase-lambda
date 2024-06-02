package example

type Address struct {
	AddressLine1 string
	AddressLine2 string
	AddressLine3 string
}

type Payload struct {
	Name    string
	Gender  string
	Address Address
}

type Upload struct {
	URL      string
	FileName string
}

type QueryParams struct {
	Status  []string `structs:"status" json:"status"`
	Name    string   `structs:"name" json:"name"`
	Age     int      `structs:"age" json:"age"`
	NumList []int    `structs:"numList" json:"numList"`
	Flag    bool     `structs:"flag" json:"flag"`
}

type PathParams struct {
	CustomerId string
}
