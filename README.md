_Repo metadata_


[![GitHub tag](https://img.shields.io/github/tag/Xumeiquer/n43?include_prereleases=&sort=semver&color=blue)](https://github.com/Xumeiquer/n43/releases/)
[![License](https://img.shields.io/badge/License-MIT-blue)](#license)


_Call-to-Action buttons_

<div align="center">

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/M4M625UW0)

</div>

## Documentation

The N43 library parser is reqlly simple, just load the lines from the N43 file into a new parser and
parse it. After that, you'll have a nice structure with all the needed data accessible.

```golang
package main

import (
    "os"
    "log"
    "string"

    "github.com/Xumeiquer/n43"
)

func main() {
    data, err := os.ReadFile(fin)
    if err != nil {
        log.Fatal(err.Error())
    }

    dataLines := strings.Split(string(data), "\n")
    parser := n43.NewParser(dataLines, ops)
    res, err := parser.Parse()
    if err != nil {
        log.Fatal(err.Error())
    }

    printOutput(*res)
}
```

## License

Released under [MIT](/LICENSE) by [@Xumeiquer](https://github.com/Xumeiquer).
