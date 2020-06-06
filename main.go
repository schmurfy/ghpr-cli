package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
)

var cfgFile string
var owner string
var repository string
var ctx context.Context

var githubToken string

var globalFlagsHelpShown bool

func showUsage(cmd string, fs *pflag.FlagSet) string {
	if !globalFlagsHelpShown {
		globalFlagsHelpShown = true
		globalFs := pflag.NewFlagSet("", pflag.ExitOnError)
		addGlobalFlags(globalFs)

		fmt.Printf("\nGlobal Flags:\n")
		globalFs.PrintDefaults()
	}

	fmt.Printf("\nUsage: %s %s\n", os.Args[0], cmd)
	fs.PrintDefaults()
	fmt.Printf("\n")
	return ""
}

var registry map[string]cli.CommandFactory

func Register(name string, fn cli.CommandFactory) {
	if registry == nil {
		registry = make(map[string]cli.CommandFactory)
	}

	registry[name] = fn
}

func parseFlags(ui cli.Ui, fs *pflag.FlagSet, args []string) {
	addGlobalFlags(fs)
	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	requiredMissing := []string{}
	invalidValues := []string{}

	// check required
	fs.VisitAll(func(flag *pflag.Flag) {
		// load from env if unset
		if !flag.Changed {
			if val, present := os.LookupEnv(strings.ToUpper(flag.Name)); present {
				flag.Value.Set(val)
				flag.Changed = true
			}
		}

		if flag.Annotations != nil {
			if _, required := flag.Annotations["required"]; required {
				if flag.Changed {
					// fmt.Printf(" - %s: %s\n", flag.Name, flag.Value.String())
				} else {
					// the flag is required and has not been changed
					requiredMissing = append(requiredMissing, flag.Name)
				}
			}
		}

		// check value if we have one
		if flag.Changed && (flag.Annotations != nil) {
			if values, present := flag.Annotations["enum"]; present {
				valid := false

				for _, v := range values {
					if v == flag.Value.String() {
						valid = true
						break
					}
				}

				if !valid {
					invalidValues = append(invalidValues, flag.Name)
				}
			}
		}
	})

	if len(requiredMissing) > 0 {
		ui.Error(fmt.Sprintf("Missing required flags: %s\n", strings.Join(requiredMissing, ",")))
		os.Exit(1)
	}

	if len(invalidValues) > 0 {
		ui.Error(fmt.Sprintf("Invalid values for %s\n", strings.Join(invalidValues, ",")))
		os.Exit(1)
	}

}

func setAllowedValues(fs *pflag.FlagSet, name string, acceptableValues []string) {
	fs.SetAnnotation(name, "enum", acceptableValues)
}

func setRequiredFlags(fs *pflag.FlagSet, flags ...string) {
	for _, flag := range flags {
		fs.SetAnnotation(flag, "required", []string{"true"})
	}
}

func addGlobalFlags(fs *pflag.FlagSet) {
	flags := []string{"token", "login", "owner", "repository"}

	fs.StringVarP(&githubToken, "token", "", "", "github token")
	fs.StringVarP(&owner, "owner", "o", "", "owner")
	fs.StringVarP(&repository, "repository", "r", "", "repository")

	setRequiredFlags(fs, flags...)
}

func main() {
	c := cli.NewCLI("ghpr", "1.0.0")
	c.Args = os.Args[1:]

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	ui := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr}

	coloredUI := &cli.ColoredUi{
		Ui:         ui,
		ErrorColor: cli.UiColorRed,
	}

	// initGithubClient()
	initStatuses(coloredUI)
	initComments(coloredUI)

	c.Commands = registry

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
