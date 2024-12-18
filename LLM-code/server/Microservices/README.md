# Running the Example

1. Open three separate terminals.

2. **Start the User Service:**

   ```bash
   cd usersvc
   go run main.go
   ```

   Logs: `[usersvc] Starting on :8081`

3. **Start the Order Service:**

   ```bash
   cd ordersvc
   go run main.go
   ```

   Logs: `[ordersvc] Starting on :8082`

4. **Start the Gateway Service:**

   ```bash
   cd gateway
   go run main.go
   ```

   Logs: `[gateway] Starting on :8080`

5. **Test the Gateway:**

   ```bash
   curl http://localhost:8080/all
   ```

   You should see combined user and order data. Initially:

   ```json
   {
     "orders": [
       {"id": "o1", "item": "Book", "user_id": "u1"}
     ],
     "users": [
       {"id": "u1", "name": "Alice"},
       {"id": "u2", "name": "Bob"}
     ]
   }
   ```

6. **Create a new User:**

   ```bash
   curl -X POST -d '{"id":"u3","name":"Charlie"}' http://localhost:8081/users
   ```

7. **Create a new Order:**

   ```bash
   curl -X POST -d '{"id":"o2","item":"Laptop","user_id":"u3"}' http://localhost:8082/orders
   ```

8. **Check all combined data again:**

   ```bash
   curl http://localhost:8080/all
   ```

   Now it should include the new user and order.

---

**Extending This Example:**

- Add authentication middleware to each service.
- Add rate limiting or caching at the gateway.
- Integrate observability tools (OpenTelemetry, Prometheus).
- Use Docker or Kubernetes for deployment.
- Introduce service discovery (e.g., Consul, etcd) instead of hardcoded URLs.
- Implement retry or circuit breaker patterns in the gateway’s `fetchJSON` logic.

---
