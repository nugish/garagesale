# 3. JSON

- Create a type `Product` with fields for Name, Cost, And Quantity
- Rename the `Echo` handler to `ListProducts`
- Create a slice of Product values with some dummy data.
- Marshal the slice to HSON and write it to the client.
- Use `w.WriteHeader` to explicity set the response status code.
- Include the Content-Type header so clients understand the response.
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
- See what happens when a nil slice is provided