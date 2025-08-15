module autotests

go 1.24.1

require github.com/yourname/your_project v0.0.0
require github.com/yourname/your_project/errors v0.0.0

require (
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)

replace github.com/yourname/your_project => ../
replace github.com/yourname/your_project/errors => ../errors
