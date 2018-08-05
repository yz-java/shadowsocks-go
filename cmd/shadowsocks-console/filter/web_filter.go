/**
 * Created with IntelliJ IDEA.
 * Description: 
 * User: yangzhao
 * Date: 2018-08-01
 * Time: 17:20
 */
 package web_filter

import (
	"net/http"
	"strings"
)

type FilterHandle func(rw http.ResponseWriter,r *http.Request) error
//拦截uri映射处理
var filterMapping = make(map[string]FilterHandle,0)
//保证有序uri
var uriArray = make([]string,0)

func Register(uri string,fh FilterHandle)  {
	uri = uri[:len(uri)-2]
	filterMapping[uri]=fh
	uriArray = append(uriArray,uri)
}

type WebHandle func(rw http.ResponseWriter,r *http.Request) error

func Handle(webHandle WebHandle) func(rw http.ResponseWriter,r *http.Request) {

	return func(rw http.ResponseWriter,r *http.Request){
		var uri=r.RequestURI
		uri+="/"
		for _,v:=range uriArray{
			if strings.Contains(uri,v) {
				e := filterMapping[v](rw, r)
				if e != nil {
					rw.Write([]byte(e.Error()))
					return
				}

			}
		}
		err := webHandle(rw, r)
		if err != nil {
			rw.Write([]byte(err.Error()))
		}
	}
}