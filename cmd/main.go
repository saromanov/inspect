package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Output struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
}

type Options struct {
	
}

func command(global *globalOptions) cli.Command {
	sharedFlags, sharedOpts := sharedImageFlags()
	imageFlags, imageOpts := imageFlags(global, sharedOpts, "", "")
	opts := inspectOptions{
		global: global,
		image:  imageOpts,
	}
	return cli.Command{
		Name:  "inspect",
		Usage: "Inspect image IMAGE-NAME",
		Description: "Description"
		ArgsUsage: "NAME",
		Flags: append(append([]cli.Flag{
		}, sharedFlags...), imageFlags...),
		Action: commandAction(opts.run),
	}
}

func (opts *inspectOptions) run(args []string, stdout io.Writer) (retErr error) {
	ctx, cancel := opts.global.commandTimeoutContext()
	defer cancel()
	return nil
}

func main(){

}