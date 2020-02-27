// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/macaroon"
)

// ConfigStorj depicts keys to search for within the storj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satelliteURL"`
	Bucket               string `json:"bucketName"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedScope      string `json:"serializedScope"`
	DisallowReads        string `json:"disallowReads"`
	DisallowWrites       string `json:"disallowWrites"`
	DisallowDeletes      string `json:"disallowDeletes"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) (ConfigStorj, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename.

	var configStorj ConfigStorj

	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)

	// Display read information.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Scope Key\t: ", configStorj.SerializedScope)

	return configStorj, nil
}

// ConnectStorjReadUploadData reads Storj configuration from given file,
// connects to the desired Storj network.
// It then reads data using io.Reader interface and
// uploads it as object to the desired bucket.
func ConnectStorjReadUploadData(fullFileName string, fileReader io.Reader, fileName string, keyValue string, restrict string) (string, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// fileReader is an io.Reader implementation that 'reads' desired data,
	// which is to be uploaded to storj V3 network.
	// fileName for adding file name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	var scope string = ""
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		log.Fatal("loadStorjConfiguration:", err)
	}

	fmt.Println("\nCreating New Uplink...")

	var cfg uplink.Config
	// Configure the user agent
	cfg.Volatile.UserAgent = "cPanel"
	ctx := context.Background()

	uplinkstorj, err := uplink.NewUplink(ctx, &cfg)
	if err != nil {
		uplinkstorj.Close()
		log.Fatal("Could not create new Uplink object:", err)
	}
	defer uplinkstorj.Close()
	var serializedScope string
	if keyValue == "key" {

		fmt.Println("Parsing the API key...")
		key, err := uplink.ParseAPIKey(configStorj.APIKey)
		if err != nil {
			uplinkstorj.Close()
			log.Fatal("Could not parse API key:", err)
		}

		fmt.Println("Opening Project...")
		proj, err := uplinkstorj.OpenProject(ctx, configStorj.Satellite, key)
		if err != nil {
			uplinkstorj.Close()
			log.Fatal("Could not open project:", err)
		}
		defer proj.Close()

		encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
		if err != nil {
			uplinkstorj.Close()
			proj.Close()
			log.Fatal("Could not create encryption key:", err)
		}

		// Creating an encryption context.
		access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)

		// Serializing the parsed access, so as to compare with the original key.
		serializedAccess, err := access.Serialize()
		if err != nil {
			uplinkstorj.Close()
			proj.Close()
			log.Fatal("Error Serialized key : ", err)
		}

		// Load the existing encryption access context
		accessParse, err := uplink.ParseEncryptionAccess(serializedAccess)
		if err != nil {
			log.Fatal(err)
		}

		if restrict == "restrict" {
			disallowRead, _ := strconv.ParseBool(configStorj.DisallowReads)
			disallowWrite, _ := strconv.ParseBool(configStorj.DisallowWrites)
			disallowDelete, _ := strconv.ParseBool(configStorj.DisallowDeletes)
			userAPIKey, err := key.Restrict(macaroon.Caveat{
				DisallowReads:   disallowRead,
				DisallowWrites:  disallowWrite,
				DisallowDeletes: disallowDelete,
			})
			if err != nil {
				log.Fatal(err)
			}
			userAPIKey, userAccess, err := accessParse.Restrict(userAPIKey,
				uplink.EncryptionRestriction{
					Bucket:     configStorj.Bucket,
					PathPrefix: configStorj.UploadPath,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			userRestrictScope := &uplink.Scope{
				SatelliteAddr:    configStorj.Satellite,
				APIKey:           userAPIKey,
				EncryptionAccess: userAccess,
			}
			serializedRestrictScope, err := userRestrictScope.Serialize()
			if err != nil {
				log.Fatal(err)
			}
			scope = serializedRestrictScope

		}
		userScope := &uplink.Scope{
			SatelliteAddr:    configStorj.Satellite,
			APIKey:           key,
			EncryptionAccess: access,
		}
		serializedScope, err = userScope.Serialize()
		if err != nil {
			log.Fatal(err)
		}
		if restrict == "" {
			scope = serializedScope
		}

		proj.Close()
		uplinkstorj.Close()
	} else {
		serializedScope = configStorj.SerializedScope

	}
	parsedScope, err := uplink.ParseScope(serializedScope)
	if err != nil {
		log.Fatal(err)
	}

	uplinkstorj, err = uplink.NewUplink(ctx, &cfg)
	if err != nil {
		log.Fatal("Could not create new Uplink object:", err)
	}
	proj, err := uplinkstorj.OpenProject(ctx, parsedScope.SatelliteAddr, parsedScope.APIKey)
	if err != nil {
		uplinkstorj.Close()
		proj.Close()
		log.Fatal("Could not open project:", err)
	}

	fmt.Println("Opening Bucket: ", configStorj.Bucket)

	// Open up the desired Bucket within the Project.
	bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
	if err != nil {
		fmt.Println("Could not open bucket", configStorj.Bucket, ":", err)
		fmt.Println("Trying to create new bucket....")
		_, err1 := proj.CreateBucket(ctx, configStorj.Bucket, nil)
		if err1 != nil {
			uplinkstorj.Close()
			proj.Close()
			bucket.Close()
			fmt.Printf("Could not create bucket %q:", configStorj.Bucket)
			log.Fatal(err1)
		} else {
			fmt.Println("Created Bucket", configStorj.Bucket)
		}
		fmt.Println("Opening created Bucket: ", configStorj.Bucket)
		bucket, err = proj.OpenBucket(ctx, configStorj.Bucket, parsedScope.EncryptionAccess)
		if err != nil {
			fmt.Printf("Could not open bucket %q: %s", configStorj.Bucket, err)
		}
	}

	defer bucket.Close()

	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	// Read data using io.Reader and upload it to Storj.
	fmt.Println("File path: ", configStorj.UploadPath+fileName)
	fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")

	err = bucket.UploadObject(ctx, configStorj.UploadPath+fileName, fileReader, nil)

	if err != nil {
		fmt.Printf("Could not upload: %s\t", err)
		return scope, err
	}

	fmt.Println("Uploading of the object to the Storj bucket: Completed!")

	return scope, nil
}
