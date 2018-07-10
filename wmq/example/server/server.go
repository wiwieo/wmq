package main

import (
  "flag"
  "wmq/server"
)

var(
  PORT = flag.String("port", "44444", "set env [-port=44444]")
)

func init() {
  flag.Parse()
}

func main() {
  server.StartTCP(*PORT)
}
