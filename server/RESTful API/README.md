# Testing the Endpoints

With the server running, you can test the endpoints using `curl`:

- **List Users**:

  ```bash
  curl http://localhost:8080/users
  ```

- **Create User**:

  ```bash
  curl -X POST -d '{"id":"3","name":"Charlie"}' http://localhost:8080/users
  ```

- **Get User**:

  ```bash
  curl http://localhost:8080/users/3
  ```

- **Update User**:

  ```bash
  curl -X PUT -d '{"name":"Charles"}' http://localhost:8080/users/3
  ```

- **Delete User**:

  ```bash
  curl -X DELETE http://localhost:8080/users/3
  ```
