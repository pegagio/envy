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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load an environment profile",
	Long:  `Specify a profile and load it into the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		doLoad(args[0], func(profile Profile) string {
			return profile.GenerateLoadScript()
		})
	},
	Args: cobra.ExactArgs(1),
}

var unloadCmd = &cobra.Command{
	Use:   "unload",
	Short: "Unload an environment profile",
	Long:  `Removes a profile from the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		doLoad(args[0], func(profile Profile) string {
			return profile.GenerateUnloadScript()
		})
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(unloadCmd)
}

func doLoad(name string, scriptf func(Profile) string) {
	verbose, err := rootCmd.Flags().GetBool(verboseFlag)
	cobra.CheckErr(err)

	profile, err := readProfile(name, verbose)
	cobra.CheckErr(err)

	// Generate script
	script := scriptf(profile)

	// Write the script to a temporary file
	filename, err := writeScriptFile(verbose, script)
	cobra.CheckErr(err)

	fmt.Println(filename)
}

func readProfile(name string, verbose bool) (Profile, error) {
	// Locate the profile file
	dir := viper.GetString(cfgProfileDir)
	profilePath := fmt.Sprintf("%s/%s.yaml", dir, name)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return Profile{}, fmt.Errorf("profile %s does not exist", name)
	}
	if verbose {
		fmt.Fprintf(os.Stderr, "Reading profile from %s\n", profilePath)
	}

	// Read in the profile file
	if profile, err := ReadProfile(profilePath); err != nil {
		return Profile{}, fmt.Errorf("failed to read profile from %s: %w", profilePath, err)
	} else {
		return profile, nil
	}
}

func writeScriptFile(verbose bool, script string) (string, error) {
	tempFile, err := os.CreateTemp("", "envy_*")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer func(tempFile *os.File) {
		if err := tempFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing temp file '%s'", tempFile.Name())
		}
	}(tempFile)

	if verbose {
		fmt.Fprintln(os.Stderr, "Temp file created:", tempFile.Name())
	}
	if _, err := tempFile.WriteString(script); err != nil {
		return "", fmt.Errorf("error writing to temp file '%s': %w", tempFile.Name(), err)
	}

	return tempFile.Name(), nil
}
