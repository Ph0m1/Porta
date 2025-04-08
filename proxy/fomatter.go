package proxy

import "strings"

// EntityFormatter formats the response data
type EntityFormatter interface {
	Format(entity Request) Response
}

type propertyFilter func(entity *Response)

type entityFilter struct {
	Target string
	Prefix string
	PropertyFilter propertyFilter
	Mapping map[string]string
}

// NewEntityFormatter creates an entity formatter with the received params
func NewEntityFormatter(target string, whitelist, blacklist []string, group string, mappings map[string]string) EntityFormatter {
	var propertyFilter propertyFilter
	if len(whitelist) > 0 {
		propertyFilter = newWhitelistingFilter(whitelist)
	}
}

func newWhitelistingFilter(whitelist []string) propertyFilter {
	wl := make(map[string]map[string]interface{}, len(whitelist))
	for _, k := range whitelist {
		keys := strings.Split(k, ".")
		tmp := make(map[string]interface{}, len(keys)-1)
		if len(keys) > 1 {
			if _, ok := wl[keys[0]]; ok {
				for _, key := range keys[1:] {
					wl[keys[0]][key] = nil
				}
			} else {
				for _, key := range keys[1:] {
					tmp[key] = nil
				}
				wl[keys[0]] = tmp
			}
		} else {
			wl[keys[0]] = tmp
		}
	}
	return func(entity *Response){
		accumulator := make(map[string]interface{}, len(whitelist))
		for k,v := range entity.Data{
			if sub, ok := wl[k]; ok{
				if len(sub) > 0{
					if tmp:=whitelistFilterSub(v,sub); len(tmp) > 0{
						accumulator[k] = tmp
					}
				}else{
					accumulator[k] = v
				}
			}
		}
		*entity = Response{accumulator,entity.IsComplete}
	}
}

func whitelistFilterSub(v interface{}, whitelist map[string]interface{}) map[string]interface{} {
	entity, ok := v.(map[string]interface{})
	if !ok{
		return map[string]interface{}{}
	}
	tmp := make(map[string]interface{}, len(whitelist))
	for k,v := range entity{
		if _, ok := whitelist[k]; ok{
			tmp[k] = v
		}
	}
	return tmp
}

func newBlacklistingFilter(blacklist []string) propertyFilter {