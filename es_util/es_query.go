package g_es

import (
	"context"
	"fmt"
	"github.com/olivere/elastic/v7"
	"reflect"
)

type QueryOption struct {
	Ctx          context.Context
	Query        elastic.Query
	Index        []string
	ReturnType   reflect.Type
	StartIndex   int
	OnceQueryNum int
	Sorters      []elastic.Sorter
	Desc         bool
	UpdateMap    map[string]string
	UpdateScript string
}

type EsQueryProxy struct {
	option *QueryOption
}

func EsSearch(option *QueryOption) ([]interface{}, error) {
	searchService := esClient.Search().
		Index(option.Index...)
	if len(option.Sorters) > 0 {
		searchService.SortBy(option.Sorters...)
	}
	if option.Query != nil {
		searchService.Query(option.Query)
	}
	//默认单页10个数据
	if option.OnceQueryNum == 0 {
		option.OnceQueryNum = 10
	}
	searchService.From(option.StartIndex).Size(option.OnceQueryNum)
	searchResult, err := searchService.Do(option.Ctx)
	if err != nil {
		return nil, err
	}
	return searchResult.Each(option.ReturnType), nil
}

func (p *EsQueryProxy) WithPagination(startIndex, onceQueryNum int) *EsQueryProxy {
	p.option.StartIndex = startIndex
	p.option.OnceQueryNum = onceQueryNum
	return p
}

func (p *EsQueryProxy) WithQuery(query elastic.Query) *EsQueryProxy {
	p.option.Query = query
	return p
}
func (p *EsQueryProxy) WithFieldSorter(field string, desc bool) *EsQueryProxy {
	sorter := elastic.NewFieldSort(field)
	if desc {
		sorter.Desc()
	}
	sorters := p.option.Sorters
	if sorters == nil {
		sorters = make([]elastic.Sorter, 0)
	}
	sorters = append(sorters, sorter)
	p.option.Sorters = sorters
	return p
}
func (p *EsQueryProxy) WithScoreSorter() *EsQueryProxy {
	sorter := elastic.NewScoreSort()
	sorters := p.option.Sorters
	if sorters == nil {
		sorters = make([]elastic.Sorter, 0)
		sorters = append(sorters, sorter)
	} else {
		//scoreSorter默认第一优先级
		sorters = append([]elastic.Sorter{sorter}, sorters...)
	}
	p.option.Sorters = sorters
	return p
}

func (p *EsQueryProxy) First() (interface{}, error) {
	p.option.StartIndex = 0
	p.option.OnceQueryNum = 1
	list, err := EsSearch(p.option)
	return list[0], err
}
func (p *EsQueryProxy) Results() ([]interface{}, error) {
	return EsSearch(p.option)
}

func UseEsSearch(ctx context.Context, query elastic.Query, rt reflect.Type, index ...string) *EsQueryProxy {
	return &EsQueryProxy{option: &QueryOption{
		Ctx:        ctx,
		Index:      index,
		Query:      query,
		ReturnType: rt,
	}}
}
func Delete(option *QueryOption) (deleted int64, err error) {
	res, err := esClient.DeleteByQuery().Index(option.Index[0]).
		Query(option.Query).
		ProceedOnVersionConflict().Do(option.Ctx)
	deleted = res.Deleted
	return
}

type EsDelProxy struct {
	option *QueryOption
}

func (d *EsDelProxy) Result() (int64, error) {
	return Delete(d.option)
}

func UseEsDel(ctx context.Context, query elastic.Query, index ...string) *EsDelProxy {
	return &EsDelProxy{option: &QueryOption{
		Ctx:   ctx,
		Index: index,
		Query: query,
	}}
}
func Update(option *QueryOption) (updated int64, err error) {
	queryService := esClient.UpdateByQuery().Index(option.Index[0]).
		Query(option.Query)
	if len(option.UpdateMap) > 0 {
		for k, v := range option.UpdateMap {
			queryService.Script(elastic.NewScript(fmt.Sprintf("ctx._source['%s']='%s'", k, v)))
		}
	}
	if option.UpdateScript != "" {
		queryService.Script(elastic.NewScript(option.UpdateScript))
	}
	res, err := queryService.ProceedOnVersionConflict().Do(option.Ctx)
	if err != nil {
		return 0, err
	}
	updated = res.Updated
	return
}

type EsUpdateProxy struct {
	option *QueryOption
}

func (d *EsUpdateProxy) Result() (int64, error) {
	return Update(d.option)
}
func (d *EsUpdateProxy) WithUpdateMap(update map[string]string) *EsUpdateProxy {
	d.option.UpdateMap = update
	return d
}

func (d *EsUpdateProxy) WithScript(script string) *EsUpdateProxy {
	d.option.UpdateScript = script
	return d
}
func UseEsUpdate(ctx context.Context, query elastic.Query, index ...string) *EsUpdateProxy {
	return &EsUpdateProxy{option: &QueryOption{
		Ctx:   ctx,
		Index: index,
		Query: query,
	}}
}
