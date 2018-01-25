# favicon-service (besticon)

This is a favicon service:

  * Supports `favicon.ico` and `apple-touch-icon.png`
  * Simple URL API
  * Fallback icon generation
  * Docker image & single binary download for [easy hosting](#hosting)

Try it out at <https://besticon-demo.herokuapp.com> or find out how to [deploy your own version](#self-hosting).

[![Build Status](https://travis-ci.org/mat/besticon.svg?branch=master)](https://travis-ci.org/mat/besticon)
[![Donate at PayPal](https://img.shields.io/badge/paypal-donate-orange.svg?style=flat)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=6F9YBSSCM6KCW "Donate once-off to this project using Paypal")


## What's this?

Websites used to have a `favicon.ico`, or not. With the introduction of the `apple-touch-icon.png` finding “the icon” for a website became more complicated. This service finds and — if necessary — generates icons for web sites.


## API

### GET /icon

This endpoint always returns an icon image for the given site — it redirects to an official icon if possible or creates and returns a fallback image if needed.

Parameter | Example         | Description    | Default
--------  | --------        | --------       | ----
url       | http://yelp.com |                                   | required
size      | 32..50..100             | Desired size range (min..perfect..max) If no image of size perfect..max nor perfect..min can be found a fallback icon will be generated. | required
formats   | png,ico         | Comma-separated list of accepted image formats: png, ico, gif | `png,ico,gif`
fallback\_icon\_url   | *HTTP image URL*         | If provided, a redirect to this image will be returned in case no suitable icon could be found. This overrides the default fallback image behaviour.  |
fallback\_icon\_color | ff0000 | If provided, letter icons will be colored with the hex value provided, rather than be grey, when no color can be found for any icon.


#### Examples

|Input URL | Icon |
|----------|------|
|<https://besticon-demo.herokuapp.com/icon?url=yelp.com&size=32..50..120>|![Icon for yelp.com](https://besticon-demo.herokuapp.com/icon?url=yelp.com&size=32..50..120)|
|<https://besticon-demo.herokuapp.com/icon?url=yelp.com&size=64..64..120>|![Icon for yelp.com](https://besticon-demo.herokuapp.com/icon?url=yelp.com&size=64..64..120)|
|<https://besticon-demo.herokuapp.com/icon?url=yelp.com>|size missing|
|<https://besticon-demo.herokuapp.com/icon?url=httpbin.org/status/404&size=32..64..120>|![Icon for non-existent page](https://besticon-demo.herokuapp.com/icon?url=httpbin.org/status/404&size=32..64..120)|
|<https://besticon-demo.herokuapp.com/icon?url=httpbin.org/status/404&size=32..64..120&fallback_icon_color=ff0000>|![Icon for non-existent page](https://besticon-demo.herokuapp.com/icon?url=httpbin.org/status/404&size=32..64..120&fallback_icon_color=ff0000)|
|<https://besticon-demo.herokuapp.com/icon?url=фминобрнауки.рф&size=32..64..120>|![Icon with cyrillic letter ф](https://besticon-demo.herokuapp.com/icon?url=фминобрнауки.рф&size=32..64..120)|


### GET /allicons.json

This endpoint returns all icons for a given site.

Parameter | Example         | Description | Default
--------  | --------        | ---------   | ----
url       | http://yelp.com |             | required
formats   | png,ico         | Comma-separated list of accepted image formats: png, ico, gif | `png,ico,gif`

#### Examples

* <https://besticon-demo.herokuapp.com/allicons.json?url=github.com>
* <https://besticon-demo.herokuapp.com/allicons.json?url=github.com&formats=png>

## Bugs & limitations

I tried hard to make this useful but please note there are some known limitations:

- Poor i18n support for letter icons ([#13](https://github.com/mat/besticon/issues/13))

Feel free to file other bugs - and offer your help - at <https://github.com/mat/besticon/issues>.

## Hosting

An easy way to host this service is to use Heroku, just go to [https://heroku.com/deploy](https://heroku.com/deploy?template=https://github.com/mat/besticon) to get started.

## Docker

A docker image is available at <https://hub.docker.com/r/matthiasluedtke/iconserver/>, generated from the [Dockerfile](https://github.com/mat/besticon/blob/master/Dockerfile) in this repo. I try to keep it updated for every release.

Note that this docker image is not used to run <https://besticon-demo.herokuapp.com> and therefore not well tested.


## Server Executable

### Download binaries

Binaries for some operating systems can be downloaded from <https://github.com/mat/besticon/releases/latest>

### Build your own

If you have Go 1.9 installed on your system you can use `go get` to fetch the source code and build the server:

	$ go get -u github.com/mat/besticon/...

If you want to build executables for a different target operating system you can add the `GOOS` and `GOARCH` environment variables:

	$ GOOS=linux GOARCH=amd64 go get -u github.com/mat/besticon/...

### Running

To start the server on default port 8080 just do

	$ iconserver

To use a different port use

	$ PORT=80 iconserver

Now when you open <http://localhost:8080/icons?url=instagram.com> you should see something like
![Screenshot of The Favicon Finder](https://github.com/mat/besticon/raw/master/the-icon-finder.png)


## Configuration

There is not a lot to configure but these environment variables exist

Variable         | Description            | Default Value
--------         | -----------            | -------------
`PORT`           | HTTP server port       | 8080
`CACHE_SIZE_MB`  | Size for the [groupcache](http://github.com/golang/groupcache)|32
`POPULAR_SITES`  | Comma-separated list of domains used on /popular page |
`HOST_ONLY_DOMAINS`           | Comma-separated list of domains where requests for http://example.com/foobar will be rewritten to http://example.com. [`HOST_ONLY_DOMAINS`](https://github.com/mat/besticon/blob/master/HOST_ONLY_DOMAINS) used by https://besticon-demo.herokuapp.com. See [#27](https://github.com/mat/besticon/issues/27). |


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
| [Noto Sans font](https://www.google.com/get/noto/) used for the generated icons | | [SIL Open Font License 1.1](http://scripts.sil.org/OFL) |
| [The icon](http://sixrevisions.com/freebies/icons/free-icons-1000/) | | [License](http://sixrevisions.com/freebies/icons/free-icons-1000/) |

## Contributors

  * Erkie - https://github.com/erkie
  * mmkal - https://github.com/mmkal
  * kspearrin - https://github.com/kspearrin

## License

MIT License (MIT)

Copyright (c) 2015-2018 Matthias Lüdtke, Hamburg - <https://github.com/mat>

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

## Donate

If you find this useful and want to donate... you would make my day :-)

[![Donate at PayPal](https://img.shields.io/badge/paypal-donate-orange.svg?style=flat)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=6F9YBSSCM6KCW "Donate once-off to this project using Paypal")
