# storj-cpanel
### Developed using libuplink version : v0.34.0

## Install and configure- Go
* Install Go for your platform by following the instructions in given link
[Refer: Installing Go](https://golang.org/doc/install#install)

* Make sure your `PATH` includes the `$GOPATH/bin` directory, so that your commands can be easily used:
```
export PATH=$PATH:$GOPATH/bin
```

## Setting up Storj-cPanel project

* Put the utropicmedia folder in ***`go/src`*** folder in your home directoy.

* Put the storj-cpanel folder in ***`go/src`*** folder in your home directory.

* Now open `terminal`, navigate to the `storj-cpanel` project folder and download following dependencies one by one required by the project:

```
$ go get -u github.com/urfave/cli
$ go get -u storj.io/storj/lib/uplink
$ go get -u ./...
```

## Set-up Files
* Create a `cpanel_property.json` file with following contents about a cpanel instance:
    * hostname :- Host Name connect to cPanel
    * username :- User Name of cPanel
    * password :- Password of cPanel

```json
    { 
        "hostname" : "cpanelHostName",
        "username": "username",
        "password": "password"
  }
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
    * apiKey :- API key created in Storj satellite gui
    * satelliteURL :- Storj Satellite URL
    * encryptionPassphrase :- Storj Encryption Passphrase.
    * bucketName :- Split file into given size before uploading.
    * uploadPath :- Path on Storj Bucket to store data (optional) or "/"
    * serializedScope:- Serialized Scope Key shared while uploading data used to access bucket without API key
    * disallowReads:- Set true to create serialized scope key with restricted read access
    * disallowWrites:- Set true to create serialized scope key with restricted write access
    * disallowDeletes:- Set true to create serialized scope key with restricted delete access

```json
    { 
         
        "apiKey": "change-me-to-the-api-key-created-in-satellite-gui",
        "satelliteURL": "us-central-1.tardigrade.io:7777",
        "bucketName": "change-me-to-desired-bucket-name",
        "uploadPath": "optionalpath/requiredfilename ",
        "encryptionPassphrase": "you'll never guess this",
        "serializedScope": "change-me-to-the-api-key-created-in-encryption-access-apiKey",
        "disallowReads": "true/false-to-disallow-reads",
        "disallowWrites": "true/false-to-disallow-writes",
        "disallowDeletes": "true/false-to-disallow-deletes"
    }
```

* Store both these files in a `config` folder.  Filename command-line arguments are optional.  defualt locations are used.

## Steps to create executable based on server architecture

Change the following command according to the server requirment.

```
$ env GOOS=target-OS GOARCH=target-architecture go build package-import-path
```

| GOOS - Target Operating System |  GOARCH - Target Platform|
| ------------------------------ | ------------------------:|
| android                        |   arm                    |
| darwin                         |  386                     |  
| darwin                         |  amd64                   |
| darwin                         |   arm                    |
| darwin                         |  386                     |  
| darwin                         |  arm64                   |
| dragonfly                      |  amd64                   |
| freebsd                        |  386                     |  
| freebsd                        |  amd64                   |
| freebsd                        |  arm                     |
| linux                          |  386                     |  
| linux                          |  amd64                   |
| linux                          |  arm                     |  
| linux                          |  arm64                   |
| linux                          |  ppc64                   |
| linux                          | ppc64le                  |  
| linux                          |  mips                    |
| linux                          |  mipsle                  |
| linux                          |  mips64                  |  
| linux                          |  mips64le                |
| netbsd                         |  386                     |
| netbsd                         |  amd64                   |
| netbsd                         |  arm                     |
| openbsd                        |  386                     |
| openbsd                        |  amd64                   |
| openbsd                        |  arm                     |
| plan9                          |  386                     |
| plan9                          |  amd64                   |
| solaris                        |  amd64                   |
| windows                        |  386                     |
| windows                        |  amd64                   |

## Build ONCE

```
$ go build storj-cpanel.go
```

## Upload executable files on server

Place the executable file along with configuration files to the user's home directory. 

## Run the command-line tool

**NOTE**: The following commands operate
* Get help
```
    $ ./storj-cpanel -h
```

* Check version
```
    $ ./storj-cpanel -v
```

* Create and Read backup data from desired cPanel instance and upload it to given Storj network bucket using Serialized Scope Key.  [note: filename arguments are optional.  default locations are used.]
```
    $ ./storj-cpanel store ./config/cpanel_property.json ./config/storj_config.json  
```

* Create and Read  backup data from desired cPanel instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates an unrestricted shareable Serialized Scope Key.  [note: filename arguments are optional. default locations are used.]
```
    $ ./storj-panel store ./config/cpanel_property.json ./config/storj_config.json key
```

* Create and Read backup data from desired cPanel instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates a restricted shareable Serialized Scope Key.  [note: filename arguments are optional. default locations are used. `restrict` can only be used with `key`]
```
    $ ./storj-cpanel store ./config/cpanel_property.json ./config/storj_config.json key restrict
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object
```
    $ ./storj-cpanel.go test 
```