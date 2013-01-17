# Mockity - For mocking out those pesky external services

### Quick summary:

A little HTTP server that will respond to a specific request (see below) with
whatever canned response you'd like.

### The pitch:

Does your application make lots of calls to external services?  Ever need to test your application or service without being connected to the network?  Ever want to test how your application responds to poor network conditions or poorly written services?

Mockity is a local stand-in web service that can simulate many of the real world problems that your application may face.

Some perliminary details:

Configuration is done via a JSON configuration file.

## Examples

### Mocking services

Simple mocking:

	[
		{"request": { 
	  			"url": "/assets",
	  			"method": "GET",
	  	 		"response": {
		  			"content-type": "application/json",
	  				"status": 200,
	  				"body": "{\"success\": true}"
	  			}
	  		}
	  	}
	]

How does it handle large response bodies you ask?  See "Returning a file as a response body" below.

Getting more specific:

	[
		{"request": { 
	  			"url": "/assets",
	  			"method": "POST",
	  			"params": {"mockity": "mocks you!"},
	  	 		"response": {
		  			"content-type": "text/html",
	  				"status": 200,
	  				"body": "<html><head><title>Mock You!</title></head><body>Geek says what?</body></html>"
	  			}
	  		}
	  	}
	]

Even more specific:

	[
		{"request": { 
	  			"url": "/assets",
	  			"method": "POST",
	  			"params": {"mockity": "mocks you!"},
	  			"headers": {"Accept": "*/*"},
	  	 		"response": {
		  			"content-type": "text/html",
	  				"status": 200,
	  				"body": "<html><head><title>Mock You!</title></head><body>You'll take anything?</body></html>"
	  			}
	  		}
	  	}
	]

Need to set a cookie?

	...
		"response": {
			"cookies": {"cookie": "Yum Yum"}
			...
		}
	
Need to set a header?

	...
		"response": {
			"headers": {"X-Mockity": ["Mocking you"]}
			...
		}

Returning a file as a response body

	...
		"response": {
			"body": "!file:/path/to/file"
			...
		}


### Directives - Where the real fun begins


Specify a slow response.  Response is written after the specified delay in ms.

	"response": {
		...
		"!directive": {"delay": 1000}
	}

Specify a hung server.  It accepts connections but never responds. (Technically it will answer in about 10 years)

	"response": {
		...
		"!directive": {"delay": -1}
	}

Specify a response that prematurely closes the connection.  Response headers and a random portion of the response body are output and the connection is closed.

	"response": {
		...
		"!directive": {"partial": true}
	}

Specify a response that redirects recursively

	"response": {
		...
		"!directive": {"loop": true}
	}


Specify a response that never terminates.  It just keeps outputing data forever.

	"response": {
		...
		"!directive": {"firehose": true}
	}

Specify a response that works sometimes but not always.

	"response": {
		...
		"!directive": {"flaky": true}
	}
