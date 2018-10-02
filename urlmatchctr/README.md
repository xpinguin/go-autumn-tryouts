# URLMatchCtr: A concurrent RegExp matches counter #
A toy application used as a personal playground for learning Golang's concurrency patterns.

## Description / Task ##
Count the number of non-overlapping RegExp matches within the stream of documents, provided through the std-input as a line-separated list of corresponding URLs.  
The number of documents processed simultaneously is explicitly limited (`-k` flag).  
"No-prefork": document processing worker ("goroutine") could only be started when there is a work to do, i.e. an unprocessed document available from the input stream.

## Features ##
* HTTP(S) and local file support
* Pattern specified as a regular expression (RE2 dialect)

## Usage ##
... TODO
