package client

type SearchQueryService struct {
	Term              string
	SearchQuery       *SearchQuery
	PageQuery         *PageQuery
	ResultFieldsQuery *ResultFieldsQuery
}

// https://www.elastic.co/guide/en/app-search/current/filters.html#filters-nesting-filters
func NewSearchQueryService(term string, query *SearchQuery) *SearchQueryService {
	return &SearchQueryService{
		Term:        term,
		SearchQuery: query,
	}
}

// https://www.elastic.co/guide/en/app-search/current/result-fields-highlights.html
func (s *SearchQueryService) WithPageQuery(pq *PageQuery) *SearchQueryService {
	s.PageQuery = pq
	return s
}

// https://www.elastic.co/guide/en/app-search/current/result-fields-highlights.html#result-fields-highlights-snippet-result-field
func (s *SearchQueryService) WithResultFieldsQuery(rfq *ResultFieldsQuery) *SearchQueryService {
	s.ResultFieldsQuery = rfq
	return s
}

func (s *SearchQueryService) Source() (r map[string]interface{}, err error) {
	r = make(map[string]interface{})
	r["query"] = s.Term

	qs, err := s.SearchQuery.Source()
	if err != nil {
		return
	}
	if qs != nil {
		r["filters"] = qs
	}

	if s.PageQuery != nil {
		var err error
		pqs, err := s.PageQuery.Source()
		if err != nil {
			return r, err
		}
		if pqs != nil {
			r["page"] = pqs
		}
	}

	if s.ResultFieldsQuery != nil {
		var err error
		rfqs, err := s.ResultFieldsQuery.Source()
		if err != nil {
			return r, err
		}
		if rfqs != nil {
			r["result_fields"] = rfqs
		}
	}

	return
}
