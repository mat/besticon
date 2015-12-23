# favicon-service (besticon)

Source code powering The Favicon Finder at <http://icons.better-idea.org>.

[![Build Status](https://travis-ci.org/mat/besticon.svg?branch=master)](https://travis-ci.org/mat/besticon)

## API

### GET /icon

This endpoint always returns an icon image for the given site — it redirects to an official icon if possible or creates and returns a fallback image if needed.

Parameter | Example         | Description    | Default
--------  | --------        | --------       | ---- 
url       | http://yelp.com |                                   | required
size      | 120             | Desired **minimum** icon size | required
formats   | png,ico         | Comma-separated list of accepted image formats: png, ico, gif | `png,ico,gif`
fallback\_icon\_url   | *HTTP image URL*         | If provided, a redirect to this image will be returned in case no suitable icon could be found. This overrides the default fallback image behaviour.  | 


#### Examples

|Input URL | Icon |
|----------|------|
|<http://icons.better-idea.org/icon?url=yelp.com&size=32>|![Icon for yelp.com](http://icons.better-idea.org/icon?url=yelp.com&size=32)|
|<http://icons.better-idea.org/icon?url=yelp.com&size=64>|![Icon for yelp.com](http://icons.better-idea.org/icon?url=yelp.com&size=64)|
|<http://icons.better-idea.org/icon?url=yelp.com>|size missing|



### GET /allicons.json

This endpoint returns all icons for a given site.

Parameter | Example         | Description | Default
--------  | --------        | ---------   | ---- 
url       | http://yelp.com |             | required
formats   | png,ico         | Comma-separated list of accepted image formats: png, ico, gif | `png,ico,gif`
pretty    | true            | Whether to return a pretty-printed JSON response, for debugging | `false`

#### Examples

* <http://icons.better-idea.org/allicons.json?url=github.com>
* <http://icons.better-idea.org/allicons.json?url=github.com&formats=png>
* <http://icons.better-idea.org/allicons.json?url=github.com&formats=png&pretty=true>


## Server Executable

### Download binaries

Binaries for some operating systems can be downloaded from <https://github.com/mat/besticon/releases/latest>

Even more binaries are available from the excellent GoBuilder community site <https://gobuilder.me/github.com/mat/besticon/besticon/iconserver>

### Build your own

If you have Go 1.5 installed on your system you can use `go get` to fetch the source code and build the server:

	$ go get -u github.com/mat/besticon/...

If you want to build executables for a different target operating system you can add the `GOOS` and `GOARCH` environment variables:

	$ GOOS=linux GOARCH=amd64 go get -u github.com/mat/besticon/...

### Running

To start the server on default port 8080 just do

	$ iconserver
	
To use a different port use

	$ PORT=80 iconserver

Now when you open <http://localhost:8080/icons?url=instagram.com> you should see something like
![Screenshot of The Favicon Finder](the-icon-finder.png)


## Libraries etc.

Package | Description | License
------  | ----------  | ------
<http://github.com/NYTimes/gziphandler> | net/http gzip compression | [Apache License, Version 2.0](https://github.com/NYTimes/gziphandler/blob/master/LICENSE.md) |
<http://github.com/PuerkitoBio/goquery> |  |[BSD style](https://github.com/PuerkitoBio/goquery/blob/master/LICENSE) |
<http://github.com/andybalholm/cascadia> | CSS selectors| [License](https://github.com/andybalholm/cascadia/blob/master/LICENSE) |
<http://github.com/golang/groupcache> | | [Apache License 2.0](https://github.com/golang/groupcache/blob/master/LICENSE)
<http://github.com/golang/protobuf> | | [License](https://github.com/golang/protobuf/blob/master/LICENSE)
<http://github.com/golang/freetype> | | [FreeType License](https://github.com/golang/freetype/blob/master/LICENSE)
<http://golang.org/x/image> | supplementary image libraries | [BSD style](https://github.com/golang/image/blob/master/LICENSE) |
<http://golang.org/x/net> | | [BSD style](https://github.com/golang/net/blob/master/LICENSE)|
<http://golang.org/x/text> | | [BSD style](https://github.com/golang/text/blob/master/LICENSE)|
| [Open Sans font](https://www.google.com/fonts/specimen/Open+Sans) used for the generated icons | | [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html) |
| [The icon](http://sixrevisions.com/freebies/icons/free-icons-1000/) used on [icons.better-idea.org](http://icons.better-idea.org) | | [License](http://sixrevisions.com/freebies/icons/free-icons-1000/) |


## License

MIT License (MIT)

Copyright (c) 2015 Matthias Lüdtke, Hamburg - <http://github.com/mat>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
