package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/containers/common/pkg/unshare"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/syndtr/gocapability/capability"
	"github.com/urfave/cli"
)

type Output struct {
	Name string `json:",omitempty"`
	Tag  string `json:",omitempty"`
}

type inspectOptions struct {
	global *globalOptions
	image  *imageOptions
	raw    bool
	config bool
}

type Options struct {
}

func reexecIfNecessaryForImages(imageNames ...string) error {
	for _, imageName := range imageNames {
		transport := alltransports.TransportFromImageName(imageName)
		if transport != nil && transport.Name() == "containers-storage" {
			return maybeReexec()
		}
	}
	return nil
}

func maybeReexec() error {
	capabilities, err := capability.NewPid(0)
	if err != nil {
		return errors.Wrapf(err, "error reading the current capabilities sets")
	}
	for _, cap := range neededCapabilities {
		if !capabilities.Get(capability.EFFECTIVE, cap) {
			unshare.MaybeReexecUsingUserNamespace(true)
			return nil
		}
	}
	return nil
}

func command(global *globalOptions) cli.Command {
	sharedFlags, sharedOpts := sharedImageFlags()
	imageFlags, imageOpts := imageFlags(global, sharedOpts, "", "")
	opts := inspectOptions{
		global: global,
		image:  imageOpts,
	}
	return cli.Command{
		Name:        "inspect",
		Usage:       "Inspect image IMAGE-NAME",
		Description: "Description",
		ArgsUsage:   "NAME",
		Flags:       append(append([]cli.Flag{}, sharedFlags...), imageFlags...),
		Action:      commandAction(opts.run),
	}
}

func (opts *inspectOptions) run(args []string, stdout io.Writer) (retErr error) {
	ctx, cancel := opts.global.commandTimeoutContext()
	defer cancel()

	if len(args) != 1 {
		return errors.New("Exactly one argument expected")
	}
	imageName := args[0]

	if err := reexecIfNecessaryForImages(imageName); err != nil {
		return err
	}

	sys, err := opts.image.newSystemContext()
	if err != nil {
		return err
	}

	defer func() {
		if err := src.Close(); err != nil {
			retErr = errors.New(fmt.Sprintf("(could not close image: %v) ", err))
		}
	}()

	return nil
}

func main() {

}
