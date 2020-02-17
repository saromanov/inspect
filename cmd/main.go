package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/containers/common/pkg/unshare"
	"github.com/containers/image/types"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/syndtr/gocapability/capability"
	"github.com/urfave/cli"
)

type Output struct {
	Name string `json:",omitempty"`
	Tag  string `json:",omitempty"`
}

type dockerImageOptions struct {
	global         *globalOptions
	shared         *sharedImageOptions
	authFilePath   optionalString
	credsOption    optionalString
	dockerCertPath string
	tlsVerify      optionalBool
	noCreds        bool
}

type inspectOptions struct {
	global *globalOptions
	image  *imageOptions
	raw    bool
	config bool
}

type imageOptions struct {
	dockerImageOptions
	sharedBlobDir    string // A directory to use for OCI blobs, shared across repositories
	dockerDaemonHost string // docker-daemon: host to connect to
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

	src, err := parseImageSource(ctx, opts.image, imageName)
	if err != nil {
		return fmt.Errorf("Error parsing image name %q: %v", imageName, err)
	}

	defer func() {
		if err := src.Close(); err != nil {
			retErr = errors.New(fmt.Sprintf("(could not close image: %v) ", err))
		}
	}()

	rawManifest, _, err := src.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to get manifest: %v", err)
	}

	if opts.raw && !opts.config {
		_, err := stdout.Write(rawManifest)
		if err != nil {
			return fmt.Errorf("unable to write manifest: %v", err)
		}
		return nil
	}

	return nil
}

func parseImageSource(ctx context.Context, opts *imageOptions, name string) (types.ImageSource, error) {
	ref, err := alltransports.ParseImageName(name)
	if err != nil {
		return nil, err
	}
	sys, err := opts.newSystemContext()
	if err != nil {
		return nil, err
	}
	return ref.NewImageSource(ctx, sys)
}

func main() {

}
