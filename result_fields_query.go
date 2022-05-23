package client

type ResultFieldsQuery struct {
	Fields map[string]interface{}
}

func NewResultFieldsQuery(f map[string]interface{}) *ResultFieldsQuery {
	return &ResultFieldsQuery{
		Fields: f,
	}
}

func (p *ResultFieldsQuery) Source() (r map[string]interface{}, err error) {
	return p.Fields, nil
}
