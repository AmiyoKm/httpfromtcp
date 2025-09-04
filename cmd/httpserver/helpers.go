package main

import (
	"bytes"
	"fmt"
)

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
		headers.Replace("Content-Type", "text/html")
  </body>
</html>`)
}

func toStr(byt []byte) string {
	out := bytes.Buffer{}

	for _, b := range byt {
		out.Write([]byte(fmt.Sprintf("%02x", b)))
	}
	return out.String()
}
