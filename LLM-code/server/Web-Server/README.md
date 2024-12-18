# Running the Server

1. **Run the server**:

   ```bash
   go run main.go
   ```

   Or set a custom port:

   ```bash
   PORT=9000 go run main.go
   ```

2. **Access the home page**:
   [http://localhost:8080/](http://localhost:8080/)
   You’ll see a simple HTML page with a form.

3. **Create a new user via form**:
   Submit the form (e.g., ID: `3`, Name: `Charlie`) and you’ll get a JSON response appended to the page if using htmx.

4. **List users**:

   ```bash
   curl http://localhost:8080/users
   ```

5. **Check health**:

   ```bash
   curl http://localhost:8080/healthz
   ```
