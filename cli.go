package main

import (
	"fmt"
	"io"
	"log"
	"os"

	flag "github.com/linyows/mflag"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1
)

// Options is structure
type Options struct {
	Config  string
	Version bool
}

// CLI is the command line object
type CLI struct {
	outStream, errStream io.Writer
	inStream             *os.File
	opt                  Options
}

var usageText = `Usage: octopass [options] <command> [args]

Commands:
  keys   get public keys for AuthorizedKeysCommand in sshd(8)
  pam    authorize with github authentication for pam_exec(8)

Options:`

var exampleText = `
Examples:
  $ octopass keys <user@github>
  $ echo <token@github> | env PAM_USER=<user@github> octopass pam

`

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	f := flag.NewFlagSet(Name, flag.ContinueOnError)
	f.SetOutput(cli.outStream)

	f.Usage = func() {
		fmt.Fprintf(cli.outStream, usageText)
		f.PrintDefaults()
		fmt.Fprint(cli.outStream, exampleText)
	}

	f.StringVar(&cli.opt.Config, []string{"c", "-config"}, "/etc/octopass.conf", "the path to the configuration file")
	f.BoolVar(&cli.opt.Version, []string{"v", "-version"}, false, "print the version and exit")

	if err := f.Parse(args[1:]); err != nil {
		return ExitCodeError
	}
	parsedArgs := f.Args()

	if len(parsedArgs) == 0 {
		f.Usage()
		return ExitCodeOK
	}

	if parsedArgs[0] != "keys" && parsedArgs[0] != "pam" {
		fmt.Fprintf(cli.errStream, "invalid argument: %s\n", parsedArgs[0])
		return ExitCodeError
	}

	if cli.opt.Version {
		fmt.Fprintf(cli.outStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	c, err := LoadConfig(cli.opt.Config)
	if err != nil {
		fmt.Fprintf(cli.errStream, "%s\n", err)
		return ExitCodeError
	}
	c.SetDefault()

	oct := NewOctopass(c, nil, cli)
	if err := oct.Run(parsedArgs); err != nil {
		log.Print("[ERR] " + fmt.Sprintf("%s", err))
		return ExitCodeError
	}

	return ExitCodeOK
}
