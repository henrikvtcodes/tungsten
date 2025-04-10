package main

import "github.com/henrikvtcodes/tungsten/cmd"

//go:generate pkl-gen-go config/server-config.pkl
func main() {
	cmd.Execute()
}