# Besticon

Source code powering The Icon Finder at <http://icons.better-idea.org>.

[![Build Status](http://img.shields.io/travis/mat/besticon/master.svg?style=flat-square)](http://travis-ci.org/mat/besticon)
[![GoDoc](https://godoc.org/github.com/mat/besticon?status.svg)](https://godoc.org/github.com/mat/besticon)


## Command Line Tool

### Dependencies

 - <http://golang.org>
 - libxml2

### Setup

    go get github.com/mat/besticon/besticon/besticon

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
