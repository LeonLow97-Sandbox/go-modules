## Building Go Modules

- Modules make life easier
- Create them for your own use, or for others to use
- Updates become easier
- Sharing new functionality becomes easier

## Commands to Create a Go Module

```
go work init toolkit app
cat go.work

// inside `trevor_sawler` directory
go work use app-upload
go work use app-dir
go work use app-slug
go work use app-download
go work use app-json
```

## Git Repository Name for Go Modules

- It has to be exactly `toolkit` because we specified like this as shown below.

```
module github.com/LeonLow97/toolkit

go 1.19
```

## Uploading 1 or More Files

- Uploading 1 or more files from the browser to the server
- Limiting uploads by file size
- Limiting uploads by mime type
- Writing a test for our new functionality
- Writing a simple application to try things out

## Creating Directories if they do not exist

- Create a directory if it does not exist
- Writing a test for our new functionality
- Writing a simple application to try things out

## Generating Slugs

- Generate Slugs: "Hello, world!" becomes hello-world
- Convert a string to one that is safe to use in a URL
- Writing a test for new functionality
- Writing a simple application to try things out

## Downloading s Static File

- Download a file from the server to the user's browser
  - Files for authenticated users
- Writing a test for new functionality
- Writing a simple application to try things out

## Working with JSON

- Reading & Writing JSON
- Good JSON error responses
- Posting JSON to a remote API
- Update our tests
- Writing a simple application to try things out

## Tagging a Release and Semantic Versioning

- Overview of semantic versioning
- Tagging a release
- Updating our code and tests
- Tagging a version 2.0.0 release

---

#### Semantic Versioning

- Given a version number MAJOR.MINOR.PATCH, increment the:
  1. MAJOR version when you make incompatible API changes
  2. MINOR version when you add **functionality** in a backwards compatible manner
  3. PATCH version when you make backwards compatible **bug fixes**

---

## Creating a Release on GitHub

- Create First Release
  1. Go to GitHub and Look for "Create a New Release"
  2. Tag it with Version Number, i.e., v1.0.0
- After making changes and want to publish as MAJOR version
  0. MUST change go.mod to `github.com/LeonLow97/toolkit/v2`
  1. Go to GitHub and look for "Releases"
  2. Click "Draft a new release"
  3. Tag it with Version Number, i.e., v2.0.0

## Importing our Module

- Create a dummy API with 2 endpoints
- Import our module

```
go get github.com/LeonLow97/toolkit/v2
```
