package main

import "github.com/henrikvtcodes/tungsten/cmd"

//go:generate pkl-gen-go config/Server.pkl
func main() {
	cmd.Execute()
}
