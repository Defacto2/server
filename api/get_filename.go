package api

// func GetFilename(c echo.Context) error {
// 	// Bind the data to ExampleRequest
// 	exampleRequest := new(model.ExampleRequest)
// 	if err := c.Bind(exampleRequest); err != nil {
// 		return err
// 	}

// 	// Manipulate the input data
// 	greeting := exampleRequest.FileName + " <-"

// 	return c.JSONBlob(
// 		http.StatusOK,
// 		[]byte(
// 			fmt.Sprintf(`{
// 				"file_name": %q,
// 				"greeting": %q,
// 			}`, exampleRequest.FileName, greeting),
// 		),
// 	)
// }
