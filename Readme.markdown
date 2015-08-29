# besticon (The Icon Finder)

Source code powering The Icon Finder at <http://icons.better-idea.org>.

[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/mat/besticon/besticon)
[![Build Status](http://img.shields.io/travis/mat/besticon/master.svg?style=flat-square)](http://travis-ci.org/mat/besticon)

[![Screenshot of The Icon Finder](the-icon-finder.png)](http://icons.better-idea.org)


## Server Executable

### Pre-built Binaries

Binaries for different operating systems are available at <https://gobuilder.me/github.com/mat/besticon/besticon/iconserver>

### Building

If you have Go already installed on your system using `go get` is probably the easiest way to fetch the source code and build the server:

	$ go get -u github.com/mat/besticon/...

You may also add the `GOOS` and `GOARCH` environment variables to build the executable for a different target operating system:

	$ GOOS=linux go get -u github.com/mat/besticon/...

### Running

To start the server on port 8080 use

	$ iconserver --port=8080



## Command Line Tool

### Dependencies

 - <http://golang.org>

### Setup

    $ go get -u github.com/mat/besticon/...

### Usage

Finding the biggest icon:

	$ besticon http://github.com 
	http://github.com:  https://github.com/apple-touch-icon-144.png

Finding all icons, sorted by biggest icon first:

	$ besticon --all http://github.com 
	http://github.com:  https://github.com/apple-touch-icon-144.png
	http://github.com:  https://github.com/apple-touch-icon.png
	http://github.com:  https://github.com/apple-touch-icon-114.png
	http://github.com:  https://github.com/apple-touch-icon-precomposed.png
	http://github.com:  https://assets-cdn.github.com/favicon.ico
	http://github.com:  https://github.com/favicon.ico

## License

The *source code* of besticon is released under the [MIT License](http://www.opensource.org/licenses/MIT).

The *icon* used for the website/server is licensed under [these terms](http://sixrevisions.com/freebies/icons/free-icons-1000/).
