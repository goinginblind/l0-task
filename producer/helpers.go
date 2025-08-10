package main

import (
	"encoding/json"
	"os"
)

// getEnv retrieves enviroment variable by its key. If it fails
// to retrieve the value, the 'fallback' value will be returned.
// A wrapper for os.LookupEnv(key).
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// streamJSONObjects reads from a file via a stream of data.
// The file should be a list of JSON objects, in such format:
//
//	[{"field1":"val1", "field2":"val2"}, {"field1":"val3", "field2":"val4"}, ... ]
//
// The returned json.RawMessage would be an object contained inside the {curly braces}.
func streamJSONObjects(filename string) (<-chan json.RawMessage, <-chan error) {
	out := make(chan json.RawMessage)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		file, err := os.Open(filename)
		if err != nil {
			errs <- err
			close(errs)
			return
		}

		dec := json.NewDecoder(file)

		// Reads the first '['
		_, err = dec.Token()
		if err != nil {
			errs <- err
			close(errs)
			return
		}

		// Reads every other object as a separate entity
		// and parses it outside
		for dec.More() {
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				errs <- err
				close(errs)
				return
			}

			out <- raw
		}

		if _, err := dec.Token(); err != nil {
			errs <- err
			close(errs)
			return
		}

		close(errs)
	}()

	return out, errs
}
