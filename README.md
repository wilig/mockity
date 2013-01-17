# Mockity - For mocking out those pesky third party API calls #


Add tests

Quick summary:

A little HTTP server that will response to a specific request (see below) with
whatever canned response you'd like.

The pitch:

Ever need to test your application or service without being connected to a 
network?  Ever want to test how your application responds to poor network conditions or poorly written services?

Mockity will let you run a local stand-in web service that can simulate many of the real world problems that your application may face.


Some perliminary details:

Configuration is done via a JSON configuration file.

Specifying a simple request/response:

	[
		{"request": { 
	  			"url": "/assets",
	  			"method": "GET",
	  	 		"response": {
		  			"content-type": "application/json",
	  				"status": 200,
	  				"set-cookie": [{"name": "mockity", "value": "is easy"}],
	  				"body": "{\"success\": true}"
	  			}
	  		}
	  	}
	]


More complex cases:

- Specify a slow response.  Headers are written immediately, body is written after the specified delay in ms.

	"response": {
		...
		"!directive": {"delay": 1000}
	}

- Specify a hung server.  It accepts connections but never responds. (Technically it will answer in about 10 years)

	"response": {
		...
		"!directive": {"delay": -1}
	}

- Specify a response that prematurely closes the connection.  Response headers are output and connection is closed.

	"response": {
		...
		"!directive": {"partial": true}
	}

- Infinite redirects?

- Firehose, just keep sending data forever.


