# libinjection 
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![codecov](https://codecov.io/gh/corazawaf/libinjection-go/branch/master/graph/badge.svg?token=RTCQXUDZQQ)](https://codecov.io/gh/corazawaf/libinjection-go)
[![CodeQL](https://github.com/corazawaf/libinjection-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/corazawaf/libinjection-go/actions/workflows/codeql.yml)

libinjection is a Go porting of the libinjection([http://www.client9.com/projects/libinjection/](http://www.client9.com/projects/libinjection/)) and it's thread safe.

## How to use
### SQLi Example
```go
package main

import (
    "fmt"
    "github.com/corazawaf/libinjection-go"
)

func main() {
    result, fingerprint := libinjection.IsSQLi("-1' and 1=1 union/* foo */select load_file('/etc/passwd')--")
    fmt.Println("=========result==========: ", result)
    fmt.Println("=======fingerprint=======: ", string(fingerprint))
}
```

### XSS Example
```go
package main

import (
	"fmt"
	"github.com/corazawaf/libinjection-go"
)

func main() {
	fmt.Println("result: ", libinjection.IsXSS("<script>alert('1')</script>"))
}
```

## License
libinjection-go is distributed under the same license as the [libinjection](http://www.client9.com/projects/libinjection/).