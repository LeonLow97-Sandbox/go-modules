package toolkit

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module. Any variable of this type will have access
// to all the methods with the reciever *Tools
type Tools struct {
	MaxFileSize        int
	AllowedFileTypes   []string
	MaxJSONSize        int
	AllowUnknownFields bool
}

// RandomString returns a string of random characters of length n, using randomStringSource
// as the source for the string
func (t *Tools) RandomString(n int) string {
	// Create a slice of runes (Unicode characters) of length n
	s, r := make([]rune, n), []rune(randomStringSource)

	// Loop through each position in the slice of runes
	for i := range s {
		// Generate a random prime number using the length of the rune slice as a seed
		p, _ := rand.Prime(rand.Reader, len(r))

		// Get two unsigned 64-bit integers: one from the generated prime number
		// and the other from the length of the rune slice
		x, y := p.Uint64(), uint64(len(r))

		// Assign a random rune from the rune slice (randomStringSource) to the current position
		// in the 's' slice based on the modulo operation between x and y
		s[i] = r[x%y]
	}

	// Convert the slice of runes 's' into a string and return it
	return string(s)
}

// UploadedFile is a struct used to save information about an uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile is just a convenience method that calls UploadFiles, but expects only one file to
// be in the upload.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles uploads one or more file to a specified directory, and gives the files a random name.
// It returns a slice containing the newly named files, the original file names, the size of the files,
// and potentially an error. If the optional last parameter is set to true, then we will not rename
// the files, but will use the original file names.
// UploadFiles handles the process of uploading files via HTTP Request
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	// Determine whether to rename the uploaded files or not
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	// Initialize a slice to hold information about the uploaded files
	var uploadedFiles []*UploadedFile

	// If MaxFileSize is not set, default to 1GB
	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := t.CreateDirIfNotExist(uploadDir)
	if err != nil {
		return nil, err
	}

	// Parse the multipart form data from the HTTP Request
	err = r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		// Return an error if the uploaded file exceeds the maximum allowed size
		return nil, errors.New("the uploaded file is too big")
	}

	// Iterate through each file in the multipart form data
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			// Process each file individually
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile

				// Open the uploaded file for reading
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				// Read the first 512 bytes of the file to determine its type
				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				// Check if the file type is permitted based on AllowedFileTypes
				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				// If the file type is not permitted, return an error
				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}

				// Reset file read pointer to the beginning
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				// Determine the new file name
				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				// Store the original file name
				uploadedFile.OriginalFileName = hdr.Filename

				// Create a new file in the upload directory
				var outfile *os.File
				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				}
				defer outfile.Close()

				// Copy the content of the uploaded file to the newly created file
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize

				// Append information about the uploaded file to the slice
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)

			// Check for any errors during file processing
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	// Return the slice containing information about uploaded files
	return uploadedFiles, nil
}

// CreateDirIfNotExist creates a directory, and all necessary parents, if it does not exist
func (t *Tools) CreateDirIfNotExist(path string) error {
	// Define file mode (permissions for the directory)
	const mode = 0755

	// Check if the directory already exists or not
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If the directory does not exist, create it along with any necessary parent directories
		err := os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}
	// If the directory already exists or if it has been successfully created, return nil (no error)
	return nil
}

// Slugify is a (very) simple means of creating a slug from a string
// Takes a string 's' and converts it into a slug, which is a URL-friendly version of the string
func (t *Tools) Slugify(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty string not permitted")
	}

	// Defining a regular expression pattern to match any characters that are not lowercase letters or digits.
	var re = regexp.MustCompile(`[^a-z\d]+`)

	// Convert the input string to lowercase and replace any characters that do not match the pattern with a hyphen ("-").
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(slug) == 0 {
		return "", errors.New("after removing characters, slug is zero length")
	}

	// If all checks pass, return the slug, which is the URL-friendly version of the input string, and nil (indicating no error).
	return slug, nil
}

// DownloadStaticFile downloads a file and tries to force the browser to avoid displaying it in the browser window
// by setting content disposition. It also allows specification of the display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	// Combine the base directory (p) and file name (file) to get the complete file path
	fp := path.Join(p, file)

	// Set the Content-Disposition header in the HTTP response.
	// This header indicates that the content should be treated as an attachment for download.
	// It specifies the filename that will be suggested to the user when downloading the file.
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))

	// ServeFile sends the specified file to the response writer.
	// It reads the file specified by 'fp' and writes it to the HTTP response.
	http.ServeFile(w, r, fp)
}

// JSONResponse is the type used for sending JSON around
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ReadJSON tries to read the body of a request and converts from json into a go data variable
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024 // 1 MB
	if t.MaxJSONSize != 0 {
		maxBytes = t.MaxJSONSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	if !t.AllowUnknownFields {
		dec.DisallowUnknownFields()
	}

	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("error unmarshaling JSON: %s", err.Error())
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{}) // decode more JSON from that file
	if err != io.EOF {
		return errors.New("body must contain only one JSON value")
	}

	return nil
}

// WriteJSON takes a response status code and arbitrary data and write json to the client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// ErrorJSON takes an error, and optionally a status code, generates and sends an error response json
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return t.WriteJSON(w, statusCode, payload)
}

// PushJSONToRemote posts arbitrary data to some URL as JSON, and returns the response, status code and error (if any)
// The final parameter, client, is optional. If none is specified, we use the standard http.Client
func (t *Tools) PushJSONToRemote(uri string, data interface{}, client ...*http.Client) (*http.Response, int, error) {
	// create json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}

	// check for custom http client
	httpClient := &http.Client{}
	if len(client) > 0 {
		httpClient = client[0]
	}

	// build the request and set the header
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	// call the remote URI
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()

	// send response back
	return response, response.StatusCode, nil
}