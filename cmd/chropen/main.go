// chropen opens URLs in a specific Chrome profile, identified by email.
//
//	chropen <email-or-profile-dir> <url> [<url>...]
//	chropen --list
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/O6lvl4/chropen/profile"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `chropen — open URLs in a specific Chrome profile

Usage:
  chropen <email-or-profile-dir> <url> [<url>...]
  chropen --list

Examples:
  chropen alice@example.com https://gmail.com
  chropen "Profile 24" https://example.com https://other.com

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Environment:
  CHROME_BIN             override Chrome binary path
  CHROME_USER_DATA_DIR   override Chrome user-data directory
`)
	}
	listFlag := flag.Bool("list", false, "list all detected Chrome profiles and exit")
	resolveFlag := flag.Bool("resolve", false, "print the profile directory for <email-or-profile> and exit")
	flag.Parse()

	if *listFlag {
		profiles, err := profile.List()
		if err != nil {
			fail(err)
		}
		for _, p := range profiles {
			fmt.Printf("  %-12s  %-40s  %s\n", p.Dir, fallback(p.Email, "-"), p.DisplayName)
		}
		return
	}

	args := flag.Args()
	if *resolveFlag {
		if len(args) != 1 {
			fail(fmt.Errorf("--resolve takes exactly one argument"))
		}
		dir, err := profile.Resolve(args[0])
		if err != nil {
			fail(err)
		}
		fmt.Println(dir)
		return
	}

	if len(args) < 2 {
		flag.Usage()
		os.Exit(2)
	}
	target := args[0]
	urls := args[1:]
	if err := profile.OpenAs(target, urls...); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "chropen: "+err.Error())
	os.Exit(1)
}

func fallback(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
