/*
Copyright © 2025 pegagio

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagAll      = "all"
	flagLoaded   = "loaded"
	flagUnloaded = "unloaded"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists envy profiles",
	Long:  `Lists envy profiles that are currently loaded, or unloaded, or available to be loaded`,
	Run: func(cmd *cobra.Command, args []string) {
		err := doList(cmd)
		cobra.CheckErr(err)
	},
	Args: cobra.NoArgs,
}

func doList(cmd *cobra.Command) error {
	verbose, err := cmd.Flags().GetBool(verboseFlag)
	if err != nil {
		return err
	}

	dir := viper.GetString(cfgProfileDir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("profile directory does not exist: %s", dir)
	}

	if b, err := cmd.Flags().GetBool(flagLoaded); err != nil {
		cobra.CheckErr(err)
	} else if b {
		return listLoaded(dir, verbose)
	}

	if b, err := cmd.Flags().GetBool(flagUnloaded); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else if b {
		return listUnloaded(dir, verbose)
	}

	// list all by default
	return listAll(dir, verbose)
}

func listAll(dir string, verbose bool) error {
	if verbose {
		fmt.Fprintln(os.Stderr, "Listing profiles in ", dir)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading profile directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") {
			envName := strings.TrimSuffix(file.Name(), ".yaml")
			fmt.Println(envName)
		}
	}

	return nil
}

func listLoaded(dir string, verbose bool) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Listing loaded profiles (dir: %s)\n", dir)
	}
	return fmt.Errorf("not implemented")
}

func listUnloaded(dir string, verbose bool) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Listing unloaded profiles (dir: %s)\n", dir)
	}
	return fmt.Errorf("not implemented")
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP(flagAll, "a", false, "List all available profiles")
	listCmd.Flags().BoolP(flagLoaded, "l", false, "List all loaded profiles")
	listCmd.Flags().BoolP(flagUnloaded, "u", false, "List all profiles that are not loaded")
	listCmd.MarkFlagsMutuallyExclusive(flagAll, flagLoaded, flagUnloaded)

}
