// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"utropicmedia/cpanel_storj_interface/cpanel"
	"utropicmedia/cpanel_storj_interface/storj"

	"github.com/urfave/cli"
)

const cpanelConfigFile = "./config/cpanel_property.json"
const storjConfigFile = "./config/storj_config.json"

// Create command-line tool to read from CLI.
var app = cli.NewApp()

// SetAppInfo sets information about the command-line application.
func setAppInfo() {
	app.Name = "Storj cPanel Connector"
	app.Usage = "Backup your cPanel file to the decentralized Storj network"
	app.Authors = []*cli.Author{{Name: "Satyam Shivam - Utropicmedia", Email: "development@utropicmedia.com"}}
	app.Version = "1.0.0"

}

// setCommands sets various command-line options for the app.
func setCommands() {

	app.Commands = []*cli.Command{
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Command to read and parse JSON information about Storj network and upload sample data",
			//\n arguments- 1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information if this fileName is not given, then data is read from ./config/storj_config.json example = ./storj-cpanel s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) error {

				// Default Storj configuration file name.
				var fullFileName = storjConfigFile
				var foundFirstFileName = false
				var foundSecondFileName = false
				var keyValue string
				var restrict string

				//Process arguments
				if len(cliContext.Args().Slice()) > 0 {
					for i := 0; i < len(cliContext.Args().Slice()); i++ {

						if !foundFirstFileName {
							fullFileName = cliContext.Args().Slice()[i]
							foundFirstFileName = true
						} else {
							if !foundSecondFileName {
								keyValue = cliContext.Args().Slice()[i]
								foundSecondFileName = true
							} else {
								restrict = cliContext.Args().Slice()[i]
							}
						}
					}
				}

				testdata := "test"
				fileName := "test.txt"
				data := []byte(testdata)

				// Create a buffer as an io.Reader implementor.
				buf := bytes.NewBuffer(data)
				_, err := storj.ConnectStorjReadUploadData(fullFileName, buf, fileName, keyValue, restrict)

				if err != nil {
					fmt.Println("Error while uploading data to the Storj bucket")
				}
				return err
			},
		},
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "Command to connect and transfer a back-up file from a desired cPanel instance to given Storj Bucket",
			//\n    arguments-\n      1. fileName [optional] = provide full file name (with complete path), storing cPanel properties in JSON format\n   if this fileName is not given, then data is read from ./config/cpanel_property.json\n      2. fileName [optional] = provide file name, storing Storj configuration in JSON format\n     if this fileName is not given, then data is read from ./config/storj_config.json\n   example = ./storj_cpanel c ./config/cpanel_property.json ./config/storj_config.json\n",
			Action: func(cliContext *cli.Context) error {

				// Default configuration file names.
				var fullFileNameStorj = storjConfigFile
				var fullFileNamecPanel = cpanelConfigFile
				var keyValue string
				var restrict string

				// process arguments - Reading fileName from the command line.
				var foundFirstFileName = false
				var foundSecondFileName = false
				var foundThirdFileName = false
				if len(cliContext.Args().Slice()) > 0 {
					for i := 0; i < len(cliContext.Args().Slice()); i++ {

						if !foundFirstFileName {
							fullFileNamecPanel = cliContext.Args().Slice()[i]
							foundFirstFileName = true
						} else {
							if !foundSecondFileName {
								fullFileNameStorj = cliContext.Args().Slice()[i]
								foundSecondFileName = true
							} else {
								if !foundThirdFileName {
									keyValue = cliContext.Args().Slice()[i]
									foundThirdFileName = true
								} else {
									restrict = cliContext.Args().Slice()[i]
								}
							}
						}
					}

				}

				// Establish connection with cPanel and get io.Reader implementor.
				cpanelReader, err := cpanel.ConnectToCpanel(fullFileNamecPanel)
				if err != nil {
					fmt.Println("Failed to establish connection with cPanel:\n")
					return err
				}

				// Fetch fullbackup from cPanel instance
				// and simultaneously store them into desired Storj bucket.
				scope, err := storj.ConnectStorjReadUploadData(fullFileNameStorj, cpanelReader.FileHandle, cpanelReader.FileName, keyValue, restrict)
				if err != nil {
					fmt.Println("Error while fetching cPanel backup data and uploading them to bucket:")
					return err
				}
				fmt.Println(" ")
				if keyValue == "key" {
					if restrict == "restrict" {
						fmt.Println("Restricted Serialized Scope Key: ", scope)
						fmt.Println(" ")
					} else {
						fmt.Println("Serialized Scope Key: ", scope)
						fmt.Println(" ")
					}
				}
				return err
			},
		},
	}
}

func main() {

	setAppInfo()

	setCommands()

	err := app.Run(os.Args)

	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
