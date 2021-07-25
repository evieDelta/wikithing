package wikithing

type Article struct {
	Pages []Page
}

type Page struct {
	Title string
	Table Table
	Body  string

	// Either markdown or formatthing
	Format string
}

type Table struct {
	Fields []string
}

type Field struct {
	Rows []string
}
